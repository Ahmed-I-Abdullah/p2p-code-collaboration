package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	constansts "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
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
	Peer                *p2p.Peer
	Git                 *gitops.Git
	PeerElectionService *ElectionService
}

func Init() {
	log.SetLogLevel("grpcService", "debug")
}

func (s *RepositoryService) Init(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	logger.Info("Received request to initialize a repository")

	if req.FromCli {
		return s.initFromCli(ctx, req, constansts.RepositoryReplicationFactor)
	}

	return s.initNotFromCli(ctx, req)
}

func (s *RepositoryService) initFromCli(ctx context.Context, req *pb.RepoInitRequest, replicationFactor int) (*pb.RepoInitResponse, error) {
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

	onlinePeers, err := s.gatherOnlinePeers(replicationFactor)
	if err != nil {
		return nil, err
	}

	return s.processOnlinePeers(ctx, req, onlinePeers, replicationFactor)
}

func (s *RepositoryService) initNotFromCli(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	if req.Name == "" {
		return &pb.RepoInitResponse{
			Success: false,
			Message: "Repository name is empty",
		}, errors.New("Repository name is empty")
	}

	if err := s.Git.InitBare(req.Name); err != nil {
		return &pb.RepoInitResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create local bare repository: %v", err),
		}, nil
	}

	return &pb.RepoInitResponse{
		Success: true,
		Message: "Repository created successfully",
	}, nil
}

func (s *RepositoryService) processOnlinePeers(ctx context.Context, req *pb.RepoInitRequest, onlinePeers []peer.ID, replicationFactor int) (*pb.RepoInitResponse, error) {
	successfulPeers, failedPeers, peerError := s.initPeers(ctx, req, onlinePeers)

	if len(successfulPeers) == 0 {
		return &pb.RepoInitResponse{
				Success: false,
				Message: fmt.Sprintf("Could not create repo on all %v peers. Errors: %v", replicationFactor, multierr.Combine(peerError...)),
			},
			fmt.Errorf("Could not create repo on all %v peers. Errors: %w", replicationFactor, len(successfulPeers), multierr.Combine(peerError...))
	}

	repoPeers, err := s.evaluateRepoPeers(replicationFactor, successfulPeers, failedPeers)
	if err != nil {
		return nil, err
	}

	repo := p2p.RepositoryPeers{
		PeerIDs:        repoPeers,
		InSyncReplicas: successfulPeers,
		Version:        1,
	}

	logger.Debugf("Storing repo in DHT: %v", repo)

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

func (s *RepositoryService) gatherOnlinePeers(replicationFactor int) ([]peer.ID, error) {
	allPeers := s.Peer.GetPeers()
	onlinePeers := make([]peer.ID, 0, replicationFactor)

	onlinePeers = append(onlinePeers, s.Peer.Host.ID())

	for _, p := range allPeers {
		if p != s.Peer.Host.ID() && !s.Peer.IsBootstrapNode(p) {
			logger.Infof("Current peer ID is %s", p.Pretty())
			onlinePeers = append(onlinePeers, p)
		}
	}

	if len(onlinePeers) < replicationFactor {
		return nil, fmt.Errorf("only found %d online peers, expected %d", len(onlinePeers), replicationFactor)
	}

	return onlinePeers, nil
}

func (s *RepositoryService) evaluateRepoPeers(replicationFactor int, successfulPeers, failedPeers []peer.ID) ([]peer.ID, error) {
	repoPeers := successfulPeers
	if len(successfulPeers) < replicationFactor {
		remainingPeers := replicationFactor - len(successfulPeers)
		if len(failedPeers) <= remainingPeers {
			return nil, errors.New("not enough peers to create repository")
		}
		repoPeers = append(successfulPeers, failedPeers[:remainingPeers]...)
	}

	return repoPeers, nil
}

func (s *RepositoryService) initPeers(ctx context.Context, req *pb.RepoInitRequest, onlinePeers []peer.ID) ([]peer.ID, []peer.ID, []error) {
	var errs []error
	var failedPeers, successfulPeers []peer.ID

	for _, p := range onlinePeers {
		success, err := s.signalCreateNewRepository(ctx, p, req.Name)

		if success {
			successfulPeers = append(successfulPeers, p)
		} else {
			failedPeers = append(failedPeers, p)
			errs = append(errs, fmt.Errorf("failed processing peer %v: %w", p, err))
		}
	}

	return successfulPeers, failedPeers, errs
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
		err = s.Git.PushToPeer(req.Name, targetGitAddress)
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
func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
	storedLeaderID, err := dhtutil.GetLeaderFromDHT(context.Background(), s.Peer, req.Name)
	if err == nil {
		addresses, err := s.checkPeerAlive(ctx, storedLeaderID)
		// If current leader alive, just return it
		if err == nil {
			return &pb.LeaderUrlResponse{
				Success:        true,
				Name:           req.Name,
				GitRepoAddress: addresses.GitAddress,
				GrpcAddress:    addresses.GrpcAddress,
			}, nil
		}
	}
	leaderResp, err := s.getLeaderFromRepoPeers(ctx, req.Name, storedLeaderID)
	logger.Debugf("Got leader response for repository: %v", leaderResp)
	if err != nil {
		logger.Errorf("Failed to get leader from repo peer. Error: %v", err)
		return nil, err
	}
	leaderID, err := peer.Decode(leaderResp.LeaderId)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Leader Id after casting: %s", leaderID)
	newLeaderAddresses, err := s.getPeerAdressesFromId(leaderID)
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
func (s *RepositoryService) getLeaderFromRepoPeers(ctx context.Context, repoName string, failedLeader peer.ID) (*pb.CurrentLeaderResponse, error) {
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, repoName)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Retrieved repo information from DHT to get leader. Info: %v", repo)
	request := &pb.CurrentLeaderRequest{
		RepoName: repoName,
	}
	for _, peerID := range repo.PeerIDs {
		if peerID != s.Peer.Host.ID() && peerID != failedLeader {
			addresses, err := s.getPeerAdressesFromId(peerID)
			if err != nil {
				logger.Warnf("Failed to get peer address to get current leader. Error: %v", err)
				continue
			}
			grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer grpcCancel()
			conn, err := grpc.DialContext(grpcCtx, addresses.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Warnf("failed to connect to peer: %s to get current leader. Error: %v", peerID, err)
				// Need to initiate a new election
				continue
			}
			defer conn.Close()
			client := pb.NewElectionClient(conn)
			resp, err := client.GetCurrentLeader(ctx, request)
			if err != nil {
				logger.Warnf("Failed to get current leader with error: %v", err)
				continue
			}
			return resp, nil
		}
	}
	// If no other peers alive. Call my local. This just makes sure that the db get updated etc.
	return s.PeerElectionService.GetCurrentLeader(ctx, request)
}
func (s *RepositoryService) getPeerAdressesFromId(peerID peer.ID) (*p2p.PeerAddresses, error) {
	if peerID == s.Peer.Host.ID() {
		address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
		return &p2p.PeerAddresses{
			ID:          peerID,
			GitAddress:  fmt.Sprintf("git://%s:%d", address, s.Peer.GitDaemonPort),
			GrpcAddress: fmt.Sprintf("%s:%d", address, s.Peer.GrpcPort),
		}, nil
	}
	peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
	if len(peerAddresses) == 0 {
		return nil, fmt.Errorf("Failed to get peer addresses for peer with ID %s from peer store", peerID.String())
	}
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
		ID:          peerID,
		GitAddress:  fmt.Sprintf("git://%s:%d", ipAddress, peerInfo.GitDaemonPort),
		GrpcAddress: fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort),
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
func (s *RepositoryService) checkPeerAlive(ctx context.Context, peerID peer.ID) (*p2p.PeerAddresses, error) {
	peerAddress, err := util.GetPeerAdressesFromId(peerID, s.Peer)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return peerAddress, nil
}
