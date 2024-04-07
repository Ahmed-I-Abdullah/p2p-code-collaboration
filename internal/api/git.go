package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	util "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils/dhtutil"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

var logger = log.Logger("grpcService")

type RepositoryService struct {
	pb.UnimplementedRepositoryServer
	Peer *p2p.Peer
	Git  *gitops.Git
}

func Init() {
	log.SetLogLevel("grpcService", "info")
}

func (s *RepositoryService) Init(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	log.SetLogLevel("grpcService", "info")
	replicationFactor := 2
	logger.Info("Received request to initialize a repository")

	if req.FromCli {
		existingRepository, err := s.Peer.DHT.GetValue(ctx, dhtutil.GetDHTPathFromRepoName(req.Name))

		if err != routing.ErrNotFound && err != nil {
			logger.Errorf("Failed to check if repository with key %s exists in DHT: %v", req.Name, err)
			return &pb.RepoInitResponse{
					Success: false,
					Message: "Failed to check if repository exists in DHT",
				},
				fmt.Errorf("Error fetching repository from DHT: %w", err)
		}

		if existingRepository != nil {
			return &pb.RepoInitResponse{
				Success: false,
				Message: "Repository already exits",
			}, errors.New("Repository already exits")
		}

		var onlinePeers []peer.ID
		allPeers := s.Peer.GetPeers()

		// Always include the current peer as one of the peers storing the repository
		onlinePeers = append(onlinePeers, s.Peer.Host.ID())

		for i := 0; i < len(allPeers); i++ {
			if allPeers[i] == s.Peer.Host.ID() || s.Peer.IsBootstrapNode(allPeers[i]) {
				continue
			}

			logger.Infof("Current peer ID is %s", allPeers[i].Pretty())
			onlinePeers = append(onlinePeers, allPeers[i])
		}

		if len(onlinePeers) < replicationFactor {
			return &pb.RepoInitResponse{Success: false, Message: fmt.Sprintf("only found %d online peers, expected %d", len(onlinePeers), replicationFactor)}, fmt.Errorf("only found %d online peers, expected %d", len(onlinePeers), replicationFactor)
		}

		successfulPeers := make([]peer.ID, 0)
		failedPeers := make([]peer.ID, 0)
		var errs []error

		for len(successfulPeers) < replicationFactor && len(onlinePeers) > 0 {
			p := onlinePeers[0]
			success, err := s.signalCreateNewRepository(ctx, p, req.Name)

			if success {
				successfulPeers = append(successfulPeers, p)
			} else if len(failedPeers) < replicationFactor {
				failedPeers = append(failedPeers, p)
				errs = append(errs, fmt.Errorf("failed processing peer %v: %w", p, err))
			}
			onlinePeers = onlinePeers[1:]
		}

		// If no one peer succeeded, we return an error
		if len(successfulPeers) == 0 {
			return &pb.RepoInitResponse{
				Success: false,
				Message: fmt.Sprintf("Could not create repo on all %v peers. Errors: %v", replicationFactor, multierr.Combine(errs...)),
			}, fmt.Errorf("Could not create repo on all %v peers. Errors: %w", replicationFactor, len(successfulPeers), multierr.Combine(errs...))
		}

		// If not enough peers succeed, we use some failed peers but don't include them in the ISR list
		repoPeers := successfulPeers
		if len(successfulPeers) < replicationFactor {
			remainingPeers := replicationFactor - len(successfulPeers)
			repoPeers = append(successfulPeers, failedPeers[:remainingPeers]...)
		}

		repo := p2p.RepositoryPeers{
			PeerIDs:        repoPeers,
			InSyncReplicas: successfulPeers,
			LeaderID:       successfulPeers[0],
			Version:        1,
		}

		if err := dhtutil.StoreRepoInDHT(ctx, s.Peer, req.Name, repo); err != nil {
			return &pb.RepoInitResponse{
				Success: false,
				Message: "Failed to store repository details in DHT",
			}, err
		}

		return &pb.RepoInitResponse{
			Success: true,
			Message: "Repo created",
		}, nil
	}

	// if the request is not from CLI
	if req.Name == "" {
		return &pb.RepoInitResponse{
			Success: false,
			Message: "Repository name is empty",
		}, errors.New("Repository name is empty")
	}

	err := s.Git.InitBare(req.Name)
	if err != nil {
		return &pb.RepoInitResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create file: %v", err),
		}, nil
	}

	return &pb.RepoInitResponse{
		Success: true,
		Message: "Repository created successfully",
	}, nil
}

func (s *RepositoryService) signalCreateNewRepository(ctx context.Context, p peer.ID, repoName string) (bool, error) {
	// there is a chance that the peer port is wrong and needs to be stored again. In this case we retry peer
	// retrieval by contacting the peer directly
	return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, false)
}

func (s *RepositoryService) signalCreateNewRepositoryRecursive(ctx context.Context, p peer.ID, repoName string, secondAttempt bool) (bool, error) {
	if p == s.Peer.Host.ID() {
		logger.Info("Initializing repository on current host")
		_, err := s.Init(ctx, &pb.RepoInitRequest{Name: repoName, FromCli: false})
		if err != nil {
			return false, fmt.Errorf("failed to initialize repository on current host: %w", err)
		}
		return true, nil
	}

	peerAddresses := s.Peer.Host.Peerstore().Addrs(p)

	if len(peerAddresses) == 0 {
		return false, fmt.Errorf("no known addresses for peer %s", p)
	}
	peerAddress := peerAddresses[0]

	var peerPorts *p2p.PeerInfo
	var err error
	if !secondAttempt {
		peerPorts, err = s.Peer.GetPeerPortsFromDB(p)
	} else {
		peerPorts, err = s.Peer.GetPeerPortsDirectly(p)
	}
	if err != nil {
		return false, fmt.Errorf("failed to get peer ports: %w", err)
	}

	ipAddress, err := util.ExtractIPAddr(peerAddress.String())
	if err != nil {
		return false, fmt.Errorf("failed to get peer IP Adress: %w", err)
	}

	target := fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)

	logger.Infof("Connecting to peer %s at address %s for gRPC communication\n", p, target)

	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcCancel()

	conn, err := grpc.DialContext(grpcCtx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if secondAttempt {
			return false, fmt.Errorf("failed to get peer ports: %w", err)
		}
		logger.Warnf("failed to connect to peer %s at address %s. Will attempt once more by grabbing Peer ports directly!", p, target)
		return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, true)
	}
	defer conn.Close()

	logger.Infof("Connected to gRPC server of peer %s.", p)

	c := pb.NewRepositoryClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.Init(ctx, &pb.RepoInitRequest{Name: repoName, FromCli: false})

	if err != nil {
		return false, fmt.Errorf("failed to initialize repository: %w", err)
	}

	return true, nil
}

func (s *RepositoryService) NotifyPushCompletion(ctx context.Context, req *pb.NotifyPushCompletionRequest) (*pb.NotifyPushCompletionResponse, error) {
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, err
	}
	// Clear the inSyncReplicas
	repo.InSyncReplicas = make([]peer.ID, 0)

	// Add the current peer (Leader) ID to inSyncReplicas
	repo.InSyncReplicas = append(repo.InSyncReplicas, s.Peer.Host.ID())

	// Update the version if needed
	repo.Version++

	// Slice to keep track of successful peers
	successfulPeers := make([]peer.ID, 0)

	// Iterate over peers in repo.PeerIDs and make pull requests
	for _, peerID := range repo.PeerIDs {
		if peerID == s.Peer.Host.ID() {
			continue
		}

		// Fetch peer information
		peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
		peerAddress := peerAddresses[0]
		ipAddress, err := util.ExtractIPAddr(peerAddress.String())
		if err != nil {
			continue
		}
		peerInfo, err := s.Peer.GetPeerPortsFromDB(peerID)
		if err != nil {
			logger.Warnf("failed to get peer ports for peer: %s", peerID)
		}

		peerInfo, err = s.Peer.GetPeerPortsDirectly(peerID)
		if err != nil {
			logger.Warnf("failed to get peer ports directly: %s", peerID)
			continue
		}

		targetGitAddress := fmt.Sprintf("git://%s:%d/%s", ipAddress, peerInfo.GitDaemonPort, req.Name)

		err = s.PushToPeer(req.Name, targetGitAddress)
		if err != nil {
			logger.Warnf("failed to push changes to peer: %s. Error: %v", peerID, err)
			continue
		}

		// If pull request is successful, add peer to successful peers slice
		successfulPeers = append(successfulPeers, peerID)
	}

	repo.InSyncReplicas = append(repo.InSyncReplicas, successfulPeers...)

	if err := dhtutil.StoreRepoInDHT(ctx, s.Peer, req.Name, *repo); err != nil {
		return &pb.NotifyPushCompletionResponse{
			Success: false,
			Message: "Failed to store repository details in DHT",
		}, err
	}
	return &pb.NotifyPushCompletionResponse{
		Success: true,
		Message: fmt.Sprintf("%d peers have successfully notified about the push change. ISR: %v", len(repo.InSyncReplicas), repo.InSyncReplicas),
	}, nil
}

func (s *RepositoryService) PushToPeer(repoName, peerGitAddress string) error {
	repoPath := fmt.Sprintf("%s/%s", s.Git.ReposDir, repoName)

	logger.Infof("repo dir: %v, %v", repoPath, peerGitAddress)

	cmd := exec.Command("git", "push", peerGitAddress, "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		fmt.Errorf("Failed to execute a git push with error: %v", err)
		return fmt.Errorf("failed to push changes: %v", err)
	}

	logger.Infof("Pushed changes successfully to %s", peerGitAddress)
	return nil
}

// func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
// 	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Make a copy of InSyncReplicas
// 	sortedReplicas := make([]peer.ID, len(repo.InSyncReplicas))
// 	copy(sortedReplicas, repo.InSyncReplicas)

// 	// Sort the copy in descending order
// 	sort.Slice(sortedReplicas, func(i, j int) bool {
// 		return sortedReplicas[i] > sortedReplicas[j]
// 	})

// 	logger.Infof("Sorted ISR List is: %v", sortedReplicas)

// 	for i := 0; i < len(sortedReplicas); i++ {
// 		var replica = sortedReplicas[i]
// 		logger.Infof("Trying to get leader with id %s", replica)

// 		if replica == s.Peer.Host.ID() {
// 			address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
// 			return &pb.LeaderUrlResponse{
// 				Success:        true,
// 				Name:           req.Name,
// 				GitRepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
// 				GrpcAddress:    fmt.Sprintf("%s:%d", address, s.Peer.GrpcPort),
// 			}, nil
// 		}
// 		peerAddresses := s.Peer.Host.Peerstore().Addrs(replica)
// 		peerAddress := peerAddresses[0]
// 		ipAddress, err := util.ExtractIPAddr(peerAddress.String())
// 		if err != nil {
// 			logger.Warnf("Failed to extract IP address for peer: %s", replica)
// 			continue
// 		}
// 		peerInfo, err := s.Peer.GetPeerPortsFromDB(replica)
// 		if err != nil {
// 			logger.Warnf("failed to get peer ports for peer: %s", replica)
// 		}

// 		peerInfo, err = s.Peer.GetPeerPortsDirectly(replica)
// 		if err != nil {
// 			logger.Warnf("failed to get peer ports directly: %s", replica)
// 			continue
// 		}

// 		grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
// 		defer grpcCancel()
// 		target := fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort)
// 		conn, err := grpc.DialContext(grpcCtx, target, grpc.WithInsecure(), grpc.WithBlock())
// 		if err != nil {
// 			logger.Warnf("failed to connect to peer: %s", target)
// 			continue
// 		}
// 		conn.Close()

// 		leaderUrlResponse := &pb.LeaderUrlResponse{
// 			Success:        true,
// 			Name:           req.Name,
// 			GitRepoAddress: fmt.Sprintf("git://%s:%d/%s", ipAddress, peerInfo.GitDaemonPort, req.Name),
// 			GrpcAddress:    fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort),
// 		}

// 		return leaderUrlResponse, nil

// 	}

// 	return &pb.LeaderUrlResponse{
// 		Success:        false,
// 		Name:           req.Name,
// 		GitRepoAddress: "",
// 		GrpcAddress:    "",
// 	}, fmt.Errorf("failed to get git address for any insync relplica")
// }

func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, err
	}

	if repo.LeaderID == s.Peer.Host.ID() {
		address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
		return &pb.LeaderUrlResponse{
			Success:        true,
			Name:           req.Name,
			GitRepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
			GrpcAddress:    fmt.Sprintf("%s:%d", address, s.Peer.GrpcPort),
		}, nil
	}

	addresses, err := s.getPeerAdressesFromId(repo.LeaderID)

	grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer grpcCancel()

	conn, err := grpc.DialContext(grpcCtx, addresses.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		logger.Warnf("failed to connect to peer: %s", repo.LeaderID)

		electionResp, err := s.initiateElectionWithAnotherPeer(ctx, repo.PeerIDs, req.Name)

		if err != nil {
			return nil, err
		}

		newLeaderAddresses, err := s.getPeerAdressesFromId(peer.ID(electionResp.NewLeaderId))

		if err != nil {
			return nil, err
		}

		return &pb.LeaderUrlResponse{
			Success:        true,
			Name:           req.Name,
			GitRepoAddress: fmt.Sprintf("%s/%s", newLeaderAddresses.GitAddress, req.Name),
			GrpcAddress:    newLeaderAddresses.GrpcAddress,
		}, nil
	}

	return &pb.LeaderUrlResponse{
		Success:        true,
		Name:           req.Name,
		GitRepoAddress: fmt.Sprintf("%s/%s", addresses.GitAddress, req.Name),
		GrpcAddress:    addresses.GrpcAddress,
	}, nil
}

func (s *RepositoryService) initiateElectionWithAnotherPeer(ctx context.Context, peers []peer.ID, repoName string) (*pb.ElectionResponse, error) {
	for _, peerID := range peers {
		if peerID != s.Peer.Host.ID() {
			addresses, err := s.getPeerAdressesFromId(peerID)

			if err != nil {
				logger.Warnf("Failed to get peer address to initiate election")
				continue
			}

			grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
			defer grpcCancel()
			conn, err := grpc.DialContext(grpcCtx, addresses.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Warnf("failed to connect to peer: %s to start an election", peerID)
				// Need to initiate a new election
				continue
			}
			conn.Close()

			client := pb.NewElectionClient(conn)

			request := &pb.ElectionRequest{
				RepoName: repoName,
				NodeId:   "0",
			}

			resp, err := client.Election(ctx, request)
			if err != nil {
				logger.Warnf("Failed to initiate election with error: %v", err)
				continue
			}

			return resp, nil
		}
	}

	return nil, errors.New("no reachable peer found to initiate election")
}

func (s *RepositoryService) getPeerAdressesFromId(peerID peer.ID) (*p2p.PeerAddresses, error) {
	peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
	peerAddress := peerAddresses[0]
	ipAddress, err := util.ExtractIPAddr(peerAddress.String())
	if err != nil {
		return nil, err
	}
	peerInfo, err := s.Peer.GetPeerPortsFromDB(peerID)
	if err != nil {
		logger.Warnf("failed to get peer ports for peer: %s", peerID)
		peerInfo, err = s.Peer.GetPeerPortsDirectly(peerID)
		if err != nil {
			logger.Warnf("failed to get peer ports directly: %s", peerID)
			return nil, err
		}
	}

	return &p2p.PeerAddresses{
		GitAddress:  fmt.Sprintf("git://%s:%d", ipAddress, peerInfo.GitDaemonPort),
		GrpcAddress: fmt.Sprintf("%s:%s", peerAddress, peerInfo.GrpcPort),
	}, nil
}

func (s *RepositoryService) Pull(ctx context.Context, req *pb.RepoPullRequest) (*pb.RepoPullResponse, error) {
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(repo.InSyncReplicas); i++ {
		var replica = repo.InSyncReplicas[i]
		if replica == s.Peer.Host.ID() {
			address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
			return &pb.RepoPullResponse{
				Success:     true,
				RepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
			}, nil
		}
		peerAddresses := s.Peer.Host.Peerstore().Addrs(replica)
		peerAddress := peerAddresses[0]
		ipAddress, err := util.ExtractIPAddr(peerAddress.String())
		if err != nil {
			continue
		}
		peerInfo, err := s.Peer.GetPeerPortsFromDB(replica)
		if err != nil {
			logger.Warnf("failed to get peer ports for peer: %s", replica)
		}

		peerInfo, err = s.Peer.GetPeerPortsDirectly(replica)
		if err != nil {
			logger.Warnf("failed to get peer ports directly: %s", replica)
			continue
		}

		grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
		defer grpcCancel()
		target := fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort)
		conn, err := grpc.DialContext(grpcCtx, target, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.Warnf("failed to connect to peer: %s", target)
			continue
		}
		conn.Close()

		return &pb.RepoPullResponse{
			Success:     true,
			RepoAddress: fmt.Sprintf("git://%s:%d/%s", ipAddress, peerInfo.GitDaemonPort, req.Name),
		}, nil

	}
	return &pb.RepoPullResponse{
		Success:     false,
		RepoAddress: "",
	}, fmt.Errorf("failed to get git address for any insync relplica")
}
