package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	constansts "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
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

func (s *RepositoryService) Init(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	log.SetLogLevel("grpcService", "info")
	replicationFactor := 2
	logger.Info("Received request to initialize a repository")

	if req.FromCli {
		existingRepository, err := s.Peer.DHT.GetValue(ctx, getDHTPathFromRepoName(req.Name))

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
			Version:        1,
		}

		if err := s.storeRepoInDHT(ctx, req.Name, repo); err != nil {
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

	ipAddress, err := extractIPAddr(peerAddress.String())
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
	repo, err := s.getRepoInDHT(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	// Clear the inSyncReplicas
	repo.InSyncReplicas = make([]peer.ID, 0)

	// Add the current peer (Leader) ID to inSyncReplicas
	repo.InSyncReplicas = append(repo.InSyncReplicas, s.Peer.Host.ID())

	// Update the version if needed
	repo.Version++

	// Construct the git address
	address, _ := extractIPAddr(s.Peer.Host.Addrs()[0].String())
	gitAddress := fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name)

	// Slice to keep track of successful peers
	successfulPeers := make([]peer.ID, 0)

	// Iterate over peers in repo.PeerIDs and make pull requests
	for _, peerID := range repo.PeerIDs {
		// Fetch peer information
		peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
		peerAddress := peerAddresses[0]
		ipAddress, err := extractIPAddr(peerAddress.String())
		if err != nil {
			continue
		}
		peerInfo, err := s.Peer.GetPeerPortsFromDB(peerID)
		if err != nil {
			logger.Warnf("failed to get peer ports for peer: %s", peerID)
			continue
		}

		// Create gRPC client connection to the peer
		grpcCtx, grpcCancel := context.WithTimeout(ctx, time.Second*5)
		defer grpcCancel()
		target := fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort)
		conn, err := grpc.DialContext(grpcCtx, target, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.Warnf("failed to connect to peer: %s", target)
			continue
		}
		defer conn.Close()

		// Make gRPC call to RequestToPull
		client := pb.NewRepositoryClient(conn)
		pullReq := &pb.RequestToPullRequest{
			Name:       req.Name,
			GitAddress: gitAddress,
		}
		_, err = client.RequestToPull(ctx, pullReq)
		if err != nil {
			logger.Warnf("failed to pull from peer: %s", peerID)
			continue
		}

		// If pull request is successful, add peer to successful peers slice
		successfulPeers = append(successfulPeers, peerID)
	}

	// Add successful peers to in-sync list if there are any
	if len(successfulPeers) > 0 {
		repo.InSyncReplicas = append(repo.InSyncReplicas, successfulPeers...)
	} else {
		// If there are no successful peers, return failure response
		return &pb.NotifyPushCompletionResponse{
			Success: false,
			Message: "No peers were successfully notified about the push change",
		}, nil
	}

	if err := s.storeRepoInDHT(ctx, req.Name, *repo); err != nil {
		return &pb.NotifyPushCompletionResponse{
			Success: false,
			Message: "Failed to store repository details in DHT",
		}, err
	}
	return &pb.NotifyPushCompletionResponse{
		Success: true,
		Message: "Peers have successfully notified about the push change",
	}, nil
}

func (s *RepositoryService) RequestToPull(ctx context.Context, req *pb.RequestToPullRequest) (*pb.RequestToPullResponse, error) {
	args := []string{"cd", s.Git.ReposDir + "/" + req.Name}
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return &pb.RequestToPullResponse{
			Success: false,
		}, fmt.Errorf("Failed to change directory: %v", err)
	}

	args = []string{"pull", req.GitAddress}
	cmd = exec.Command("git", args...)
	cmd.Dir = s.Git.ReposDir + "/" + req.Name
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return &pb.RequestToPullResponse{
			Success: true,
		}, nil
	} else {
		return &pb.RequestToPullResponse{
			Success: false,
		}, fmt.Errorf("Failed to pull latest changes: %v", err)
	}

}

func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
	repo, err := s.getRepoInDHT(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	// Make a copy of InSyncReplicas
	sortedReplicas := make([]peer.ID, len(repo.InSyncReplicas))
	copy(sortedReplicas, repo.InSyncReplicas)

	// Sort the copy in descending order
	sort.Slice(sortedReplicas, func(i, j int) bool {
		return sortedReplicas[i] > sortedReplicas[j]
	})

	for i := 0; i < len(sortedReplicas); i++ {
		var replica = repo.InSyncReplicas[i]
		if replica == s.Peer.Host.ID() {
			address, _ := extractIPAddr(s.Peer.Host.Addrs()[0].String())
			return &pb.LeaderUrlResponse{
				Success:        true,
				Name:           req.Name,
				GitRepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
				GrpcAddress:    fmt.Sprintf("%s:%s", address, s.Peer.GrpcPort),
			}, nil
		}
		peerAddresses := s.Peer.Host.Peerstore().Addrs(replica)
		peerAddress := peerAddresses[0]
		ipAddress, err := extractIPAddr(peerAddress.String())
		if err != nil {
			continue
		}
		peerInfo, err := s.Peer.GetPeerPortsFromDB(replica)
		if err != nil {
			logger.Warnf("failed to get peer ports for peer: %s", replica)
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

		return &pb.LeaderUrlResponse{
			Success:        true,
			Name:           req.Name,
			GitRepoAddress: fmt.Sprintf("git://%s:%d/%s", ipAddress, peerInfo.GitDaemonPort, req.Name),
			GrpcAddress:    fmt.Sprintf("%s:%s", peerAddress, peerInfo.GrpcPort),
		}, nil

	}
	return &pb.LeaderUrlResponse{
		Success:        false,
		Name:           req.Name,
		GitRepoAddress: "",
		GrpcAddress:    "",
	}, fmt.Errorf("failed to get git address for any insync relplica")
}

func (s *RepositoryService) Pull(ctx context.Context, req *pb.RepoPullRequest) (*pb.RepoPullResponse, error) {
	repo, err := s.getRepoInDHT(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(repo.InSyncReplicas); i++ {
		var replica = repo.InSyncReplicas[i]
		if replica == s.Peer.Host.ID() {
			address, _ := extractIPAddr(s.Peer.Host.Addrs()[0].String())
			return &pb.RepoPullResponse{
				Success:     true,
				RepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
			}, nil
		}
		peerAddresses := s.Peer.Host.Peerstore().Addrs(replica)
		peerAddress := peerAddresses[0]
		ipAddress, err := extractIPAddr(peerAddress.String())
		if err != nil {
			continue
		}
		peerInfo, err := s.Peer.GetPeerPortsFromDB(replica)
		if err != nil {
			logger.Warnf("failed to get peer ports for peer: %s", replica)
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

func (s *RepositoryService) storeRepoInDHT(ctx context.Context, key string, repo p2p.RepositoryPeers) error {
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	if err := s.Peer.DHT.PutValue(ctx, getDHTPathFromRepoName(key), repoBytes); err != nil {
		return fmt.Errorf("failed to store repository peers data in DHT: %w", err)
	}
	return nil
}

func (s *RepositoryService) getRepoInDHT(ctx context.Context, key string) (*p2p.RepositoryPeers, error) {
	repoBytes, err := s.Peer.DHT.GetValue(ctx, getDHTPathFromRepoName(key))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize repository peers data: %w", err)
	}
	var repo p2p.RepositoryPeers
	err = json.Unmarshal(repoBytes, &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve repository peers data in DHT: %w", err)
	}
	return &repo, nil
}

func extractIPAddr(address string) (string, error) {
	parts := strings.Split(address, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected format")
	}

	return parts[2], nil
}

func getDHTPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constansts.DHTRepoPrefix, repoName)
}

func StartServer(ctx context.Context, peer *p2p.Peer, git *gitops.Git) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", peer.GrpcPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()

	pb.RegisterRepositoryServer(s, &RepositoryService{Peer: peer, Git: git})

	return s.Serve(lis)
}
