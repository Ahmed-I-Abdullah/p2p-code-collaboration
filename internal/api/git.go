// Package api provides the implementation of gRPC APIs for managing leader election and repository related functionalities
package api

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

// Logger for the API package
var logger = log.Logger("grpcService")

// RepositoryService implements the gRPC service for managing repositories
type RepositoryService struct {
	pb.UnimplementedRepositoryServer
	Peer                *p2p.Peer
	Git                 *gitops.Git
	PeerElectionService *ElectionService
	PushInProgress      map[string]bool
	mu                  sync.Mutex
}

// NewRepositoryService creates a new instance of RepositoryService
func NewRepositoryService(peer *p2p.Peer, git *gitops.Git, electionService *ElectionService) *RepositoryService {
	return &RepositoryService{
		Peer:                peer,
		Git:                 git,
		PeerElectionService: electionService,
		PushInProgress:      make(map[string]bool),
	}
}

// Init initializes the logging for the api package
func Init() {
	log.SetLogLevel("grpcService", "debug")
}

// AcquireLock handles the request to acquire a lock for a repository in order to start a push operation
func (s *RepositoryService) AcquireLock(ctx context.Context, req *pb.AcquireLockRequest) (*pb.AcquireLockResponse, error) {
	// Mutex lock to prevent concurrent access to the PushInProgress map
	s.mu.Lock()
	defer s.mu.Unlock()

	anotherPushInProgress, ok := s.PushInProgress[req.RepoName]

	// reject acquire request if another push is in progress
	if anotherPushInProgress {
		logger.Warnf("AcquireLock: Another push in progress for repository %s. Rejecting request.", req.RepoName)
		return &pb.AcquireLockResponse{
			Ok: false,
		}, nil
	}

	if !ok {
		logger.Debugf("No push in progress for repository %s. Accepting lock acquire request.", req.RepoName)
		s.PushInProgress[req.RepoName] = true
	}

	return &pb.AcquireLockResponse{
		Ok: true,
	}, nil
}

// ReleaseRepositoryLock releases the lock for a repository
func (s *RepositoryService) ReleaseRepositoryLock(repoName string) {
	s.mu.Lock()
	logger.Debugf("Push in progress map before deleting: %v", s.PushInProgress)
	logger.Debugf("Deleting key %s from InProgressPushed map", repoName)
	delete(s.PushInProgress, repoName)
	logger.Debugf("Push in progress map after deleting: %v", s.PushInProgress)
	s.mu.Unlock()
}

// Init handles the initialization of a repository
func (s *RepositoryService) Init(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	logger.Info("Received request to initialize a repository")

	if req.FromCli {
		return s.initFromCli(ctx, req, constansts.RepositoryReplicationFactor)
	}

	return s.initNotFromCli(ctx, req)
}

// initFromCli handles the initialization of a repository initiated from the command line
func (s *RepositoryService) initFromCli(ctx context.Context, req *pb.RepoInitRequest, replicationFactor int) (*pb.RepoInitResponse, error) {
	// Check if repository already exists in DHT
	existingRepository, err := s.Peer.DHT.GetValue(ctx, dhtutil.GetDHTPathFromRepoName(req.Name))
	if err != routing.ErrNotFound && err != nil {
		logger.Errorf("Failed to check if repository with key %s exists in DHT: %v", req.Name, err)
		return &pb.RepoInitResponse{
				Success: false,
				Message: "Failed to check if repository exists in DHT",
			},
			fmt.Errorf("Error fetching repository from DHT: %w", err)
	}

	// If repository already exists, return error
	if existingRepository != nil {
		return &pb.RepoInitResponse{
			Success: false,
			Message: "Repository already exits",
		}, errors.New("Repository already exits")
	}

	// Gather online peers to replicate the repository
	onlinePeers, err := s.gatherOnlinePeers(replicationFactor)
	if err != nil {
		return nil, err
	}

	// Process online peers to initiate repository creation
	return s.processOnlinePeers(ctx, req, onlinePeers, replicationFactor)
}

// initNotFromCli handles the initialization of a repository not initiated from the command line (local initialization)
func (s *RepositoryService) initNotFromCli(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	// return error if repo name is empty
	if req.Name == "" {
		return &pb.RepoInitResponse{
			Success: false,
			Message: "Repository name is empty",
		}, errors.New("Repository name is empty")
	}

	// Initialize the repository locally
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

// processOnlinePeers initializes the repository on online peers, handles successful and failed attempts, and stores repository details in DHT
func (s *RepositoryService) processOnlinePeers(ctx context.Context, req *pb.RepoInitRequest, onlinePeers []peer.ID, replicationFactor int) (*pb.RepoInitResponse, error) {
	// Attempt to initialize repository on online peers
	successfulPeers, failedPeers, peerError := s.initPeers(ctx, req, onlinePeers)

	// If no successful peers, return error
	if len(successfulPeers) == 0 {
		return &pb.RepoInitResponse{
				Success: false,
				Message: fmt.Sprintf("Could not create repo on all %v peers. Errors: %v", replicationFactor, multierr.Combine(peerError...)),
			},
			fmt.Errorf("Could not create repo on all %v peers. Errors: %w", replicationFactor, len(successfulPeers), multierr.Combine(peerError...))
	}

	// Evaluate repository peers based on replication factor and successful/failed peers
	repoPeers, err := s.evaluateRepoPeers(replicationFactor, successfulPeers, failedPeers)
	if err != nil {
		return nil, err
	}

	// Construct repository information
	repo := p2p.RepositoryPeers{
		PeerIDs:        repoPeers,
		InSyncReplicas: successfulPeers,
		Version:        1,
	}

	logger.Debugf("Storing repo in DHT: %v", repo)

	// Store repository details in DHT
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

// gatherOnlinePeers retrieves online peers based on replication factor
func (s *RepositoryService) gatherOnlinePeers(replicationFactor int) ([]peer.ID, error) {
	// Get all available peers
	allPeers := s.Peer.GetPeers()
	onlinePeers := make([]peer.ID, 0, replicationFactor)

	// Always initialize repository on current peer
	onlinePeers = append(onlinePeers, s.Peer.Host.ID())

	// Iterate through peers, excluding bootstrap node and current peer
	for _, p := range allPeers {
		if p != s.Peer.Host.ID() && !s.Peer.IsBootstrapNode(p) {
			logger.Infof("Current peer ID is %s", p.Pretty())
			onlinePeers = append(onlinePeers, p)
		}
	}

	// Check if enough online peers found
	if len(onlinePeers) < replicationFactor {
		return nil, fmt.Errorf("only found %d online peers, expected %d", len(onlinePeers), replicationFactor)
	}

	return onlinePeers, nil
}

// evaluateRepoPeers evaluates repository peers based on replication factor and successful/failed peers
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

// initPeers initiates repository creation process on online peers and returns successful, failed peers, and any errors encountered
func (s *RepositoryService) initPeers(ctx context.Context, req *pb.RepoInitRequest, onlinePeers []peer.ID) ([]peer.ID, []peer.ID, []error) {
	var errs []error
	var failedPeers, successfulPeers []peer.ID

	// Iterate through online peers and attempt repository creation
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

// signalCreateNewRepository initiates repository creation process on a specific peer
// It first attempts to create the repository on the specified peer, and if unsuccessful, it retries recursively with new peer ports
func (s *RepositoryService) signalCreateNewRepository(ctx context.Context, p peer.ID, repoName string) (bool, error) {
	// Try creating a repository on peer by getting ports from DB and contacting them via gRPC
	return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, false)
}

// signalCreateNewRepositoryRecursive recursively tries to create a new repository on the specified peer
func (s *RepositoryService) signalCreateNewRepositoryRecursive(ctx context.Context, p peer.ID, repoName string, secondAttempt bool) (bool, error) {
	// If the target peer is the current host, initialize the repository locally
	if p == s.Peer.Host.ID() {
		logger.Info("Initializing repository on current host")
		_, err := s.Init(ctx, &pb.RepoInitRequest{Name: repoName, FromCli: false})
		if err != nil {
			return false, fmt.Errorf("failed to initialize repository on current host: %w", err)
		}
		return true, nil
	}

	// Retrieve peer addresses from the peerstore.
	peerAddresses := s.Peer.Host.Peerstore().Addrs(p)
	if len(peerAddresses) == 0 {
		return false, fmt.Errorf("no known addresses for peer %s", p)
	}
	peerAddress := peerAddresses[0]

	// Get peer ports either from the database or directly via stream handlers
	peerPorts, err := s.GetPeerPorts(ctx, p, secondAttempt)
	if err != nil {
		return false, fmt.Errorf("failed to get peer ports: %w", err)
	}

	// Extract IP address from the peer address
	ipAddress, err := util.ExtractIPAddr(peerAddress.String())
	if err != nil {
		return false, fmt.Errorf("failed to get peer IP Address: %w", err)
	}

	// Construct the target address for gRPC communication
	target := fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)
	logger.Infof("Connecting to peer %s at address %s for gRPC communication\n", p, target)

	// Dial the gRPC server of the peer
	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		// Retry once more if the connection failed, using the ports directly
		if secondAttempt {
			return false, fmt.Errorf("failed to connect to peer %s at address %s: %w", p, target, err)
		}
		logger.Warnf("Failed to connect to peer %s at address %s. Attempting once more by grabbing Peer ports directly!", p, target)
		return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, true)
	}
	defer conn.Close()

	logger.Infof("Connected to gRPC server of peer %s.", p)
	c := pb.NewRepositoryClient(conn)

	// Initialize the repository on the peer using gRPC
	_, err = c.Init(ctx, &pb.RepoInitRequest{Name: repoName, FromCli: false})
	if err != nil {
		return false, fmt.Errorf("failed to initialize repository: %w", err)
	}
	return true, nil
}

// GetPeerPorts retrieves peer ports either from the embedded database (BadgerDB) or directly via stream handlers
func (s *RepositoryService) GetPeerPorts(ctx context.Context, p peer.ID, secondAttempt bool) (*p2p.PeerInfo, error) {
	var peerPorts *p2p.PeerInfo
	var err error
	if !secondAttempt {
		peerPorts, err = s.Peer.GetPeerPortsFromDB(p)
	} else {
		peerPorts, err = s.Peer.GetPeerPortsDirectly(p)
	}
	return peerPorts, err
}

// NotifyPushCompletion notifies about the completion of a push operation to the repository
func (s *RepositoryService) NotifyPushCompletion(ctx context.Context, req *pb.NotifyPushCompletionRequest) (*pb.NotifyPushCompletionResponse, error) {
	// Defer the release of the repository lock.
	defer s.ReleaseRepositoryLock(req.Name)

	// Retrieve repository information from the DHT.
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, err
	}

	// Clear the inSyncReplicas
	repo.InSyncReplicas = make([]peer.ID, 0)
	// Add the current peer (Leader) ID to inSyncReplicas
	repo.InSyncReplicas = append(repo.InSyncReplicas, s.Peer.Host.ID())
	// Update the version
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

		// Get peer ports either from the database or directly via stream handlers.
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

		// If push is successful, add peer to successful peers slice
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

// GetLeaderUrl retrieves the URL of the leader for a given repository
// It first attempts to retrieve the leader from the Distributed Hash Table (DHT)
// If the leader is alive, it returns the leader's URL. Otherwise, it tries to
// get the leader from repository peers (by initiaitng a new leader election) and returns the URL of the new leader
func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
	// Attempt to get leader ID from DHT
	storedLeaderID, err := dhtutil.GetLeaderFromDHT(ctx, s.Peer, req.Name)
	if err == nil {
		// Check if the current leader is alive
		addresses, err := s.checkPeerAlive(ctx, storedLeaderID)
		if err == nil {
			// If leader is alive, return its URL
			logger.Debugf("Leader from DHT with ID %s is alive! Returning response", storedLeaderID.String())
			return &pb.LeaderUrlResponse{
				Success:        true,
				Name:           req.Name,
				GitRepoAddress: fmt.Sprintf("%s/%s", addresses.GitAddress, req.Name),
				GrpcAddress:    addresses.GrpcAddress,
			}, nil
		}
	}

	// If leader from DHT is not alive, attempt to get leader from repository peers by initiaitng a new leader election process
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

	// Get addresses of the new leader
	newLeaderAddresses, err := s.GetPeerAdressesFromID(leaderID)
	if err != nil {
		return nil, err
	}

	// Return the URL of the new leader
	return &pb.LeaderUrlResponse{
		Success:        true,
		Name:           req.Name,
		GitRepoAddress: fmt.Sprintf("%s/%s", newLeaderAddresses.GitAddress, req.Name),
		GrpcAddress:    newLeaderAddresses.GrpcAddress,
	}, nil
}

// getLeaderFromRepoPeers retrieves the leader from repository peers (initiates an election)
func (s *RepositoryService) getLeaderFromRepoPeers(ctx context.Context, repoName string, failedLeader peer.ID) (*pb.CurrentLeaderResponse, error) {
	// Get repository information from DHT
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, repoName)
	if err != nil {
		return nil, err
	}

	request := &pb.CurrentLeaderRequest{
		RepoName: repoName,
	}

	for _, peerID := range repo.PeerIDs {
		// Skip current peer and failed leader
		if peerID != s.Peer.Host.ID() && peerID != failedLeader {
			addresses, err := s.GetPeerAdressesFromID(peerID)
			if err != nil {
				logger.Warnf("Failed to get peer address to get current leader. Error: %v", err)
				continue
			}

			// Dial to peer's gRPC address
			grpcCtx, grpcCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer grpcCancel()
			conn, err := grpc.DialContext(grpcCtx, addresses.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Warnf("failed to connect to peer: %s to get current leader. Error: %v", peerID, err)
				// Need to initiate a new election
				continue
			}
			defer conn.Close()

			// Call gRPC method to get current leader
			client := pb.NewElectionClient(conn)
			resp, err := client.GetCurrentLeader(ctx, request)
			if err != nil {
				logger.Warnf("Failed to get current leader with error: %v", err)
				continue
			}
			return resp, nil
		}
	}

	// If no other peers alive, call the local peer to update the database
	return s.PeerElectionService.GetCurrentLeader(ctx, request)
}

// GetPeerAdressesFromID retrieves peer addresses from a given peer ID
func (s *RepositoryService) GetPeerAdressesFromID(peerID peer.ID) (*p2p.PeerAddresses, error) {
	// If peer is the local peer
	if peerID == s.Peer.Host.ID() {
		address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
		return &p2p.PeerAddresses{
			ID:          peerID,
			GitAddress:  fmt.Sprintf("git://%s:%d", address, s.Peer.GitDaemonPort),
			GrpcAddress: fmt.Sprintf("%s:%d", address, s.Peer.GrpcPort),
		}, nil
	}

	// If peer is remote
	peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
	if len(peerAddresses) == 0 {
		return nil, fmt.Errorf("Failed to get peer addresses for peer with ID %s from peer store", peerID.String())
	}
	peerAddress := peerAddresses[0]
	ipAddress, err := util.ExtractIPAddr(peerAddress.String())
	if err != nil {
		return nil, err
	}

	// Get peer ports from the database
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

// Pull retrieves the URL of the repository for pulling changes
func (s *RepositoryService) Pull(ctx context.Context, req *pb.RepoPullRequest) (*pb.RepoPullResponse, error) {
	// Get repository information from DHT
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, err
	}

	// loop over in-sync replicas
	for i := 0; i < len(repo.InSyncReplicas); i++ {
		var replica = repo.InSyncReplicas[i]
		if replica == s.Peer.Host.ID() {
			address, _ := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
			return &pb.RepoPullResponse{
				Success:     true,
				RepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
			}, nil
		}

		// Get replica addresses
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

		// Attempt to dial peer on gRPC
		grpcCtx, grpcCancel := context.WithTimeout(context.Background(), time.Second*5)
		defer grpcCancel()
		target := fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort)
		conn, err := grpc.DialContext(grpcCtx, target, grpc.WithInsecure(), grpc.WithBlock())

		if err != nil {
			// If there was an error dialing peer, try the next peer in the ISR
			logger.Warnf("failed to connect to peer: %s", target)
			continue
		}

		conn.Close()
		return &pb.RepoPullResponse{
			Success:     true,
			RepoAddress: fmt.Sprintf("git://%s:%d/%s", ipAddress, peerInfo.GitDaemonPort, req.Name),
		}, nil
	}

	// If no in-sync replicas found return an error
	return &pb.RepoPullResponse{
		Success:     false,
		RepoAddress: "",
	}, fmt.Errorf("failed to get git address for any in-sync replica")
}

// checkPeerAlive checks if a peer is alive by pinging it
func (s *RepositoryService) checkPeerAlive(ctx context.Context, peerID peer.ID) (*p2p.PeerAddresses, error) {
	// Get peer addresses
	peerAddress, err := util.GetPeerAdressesFromID(peerID, s.Peer)
	if err != nil {
		return nil, err
	}

	// Dial to peer's gRPC address
	conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create gRPC client
	client := pb.NewElectionClient(conn)

	// Ping the peer
	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	pingReq := &pb.PingRequest{}
	_, err = client.Ping(pingCtx, pingReq)
	if err != nil {
		return nil, err
	}

	return peerAddress, nil
}
