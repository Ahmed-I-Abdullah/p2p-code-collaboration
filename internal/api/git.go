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

var logger = log.Logger("grpcService")

type RepositoryService struct {
	pb.UnimplementedRepositoryServer
	Peer                *p2p.Peer
	Git                 *gitops.Git
	PeerElectionService *ElectionService
}

func NewRepositoryService(peer *p2p.Peer, git *gitops.Git, electionService *ElectionService) *RepositoryService {
	return &RepositoryService{
		Peer:                peer,
		Git:                 git,
		PeerElectionService: electionService,
	}
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

// signalCreateNewRepository initiates repository creation process on a specific peer
func (s *RepositoryService) signalCreateNewRepository(ctx context.Context, p peer.ID, repoName string) (bool, error) {
	// Try creating a repository on peer by getting ports from DB and contacting them via gRPC
	return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, false)
}

// signalCreateNewRepositoryRecursive recursively tries to create a new repository on the specified peer
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

	peerPorts, err := s.GetPeerPorts(ctx, p, secondAttempt)
	if err != nil {
		return false, fmt.Errorf("failed to get peer ports: %w", err)
	}

	ipAddress, err := util.ExtractIPAddr(peerAddress.String())
	if err != nil {
		return false, fmt.Errorf("failed to get peer IP Address: %w", err)
	}

	target := fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)
	logger.Infof("Connecting to peer %s at address %s for gRPC communication\n", p, target)

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if secondAttempt {
			return false, fmt.Errorf("failed to connect to peer %s at address %s: %w", p, target, err)
		}
		logger.Warnf("Failed to connect to peer %s at address %s. Attempting once more by grabbing Peer ports directly!", p, target)
		return s.signalCreateNewRepositoryRecursive(ctx, p, repoName, true)
	}
	defer conn.Close()

	logger.Infof("Connected to gRPC server of peer %s.", p)
	c := pb.NewRepositoryClient(conn)

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

// NotifyPushCompletion pushes the new repository changes to all the other peers and updates the ISR list with the ids of the successful peers
func (s *RepositoryService) NotifyPushCompletion(ctx context.Context, req *pb.NotifyPushCompletionRequest) (*pb.NotifyPushCompletionResponse, error) {
	logger.Infof("Starting NotifyPushCompletion for repository: %s", req.Name)

	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		logger.Errorf("Failed to get repository %s from DHT: %v", req.Name, err)
		return nil, err
	}

	// Always add the leader in the ISR list
	repo.InSyncReplicas = []peer.ID{s.Peer.Host.ID()}
	repo.Version++

	successfulPeersCh := make(chan peer.ID)
	var wg sync.WaitGroup

	// waitgroup with the number of peers to track goroutines completions
	wg.Add(len(repo.PeerIDs))

	for _, peerID := range repo.PeerIDs {
		if peerID == s.Peer.Host.ID() {
			wg.Done()
			continue
		}

		go func(peerID peer.ID) {
			defer wg.Done()
			// if context is done, exit the goroutine
			select {
			case <-ctx.Done():
				logger.Warnf("Context cancelled or timed-out for peer: %s", peerID)
				return
			default:
			}

			if err := s.pushToPeer(ctx, req.Name, peerID); err == nil {
				logger.Debugf("Successfully pushed changes to peer: %s", peerID)
				successfulPeersCh <- peerID
			} else {
				logger.Warnf("Failed to push changes to peer: %s. Error: %v", peerID, err)
			}
		}(peerID)
	}

	go func() {
		// waiting for all the goroutines to finish
		wg.Wait()
		// closing the channel as no more data will be pushed once everyone exits
		close(successfulPeersCh)
	}()

	for peerID := range successfulPeersCh {
		logger.Debugf("Successfully received successful peer: %s", peerID)
		repo.InSyncReplicas = append(repo.InSyncReplicas, peerID)
	}

	logger.Debugf("Finished processing successful peers")

	if err := dhtutil.StoreRepoInDHT(ctx, s.Peer, req.Name, *repo); err != nil {
		logger.Errorf("Failed to store repository details %s in DHT: %v", req.Name, err)
		return &pb.NotifyPushCompletionResponse{
			Success: false,
			Message: "Failed to store repository details in DHT",
		}, err
	}

	logger.Infof("Successfully stored repository details %s in DHT. ISR: %v", req.Name, repo.InSyncReplicas)

	return &pb.NotifyPushCompletionResponse{
		Success: true,
		Message: fmt.Sprintf("%d peers have successfully notified about the push change. ISR: %v", len(repo.InSyncReplicas), repo.InSyncReplicas),
	}, nil
}

func (s *RepositoryService) pushToPeer(ctx context.Context, repoName string, peerID peer.ID) error {
	logger.Debugf("Starting push to peer: %s", peerID)

	peerAddress, err := util.GetPeerAdressesFromID(peerID, s.Peer)
	if err != nil {
		logger.Warnf("Failed to get peer addresses for peer: %s", peerID)
		return err
	}

	targetGitAddress := fmt.Sprintf("%s/%s", peerAddress.GitAddress, repoName)
	if err := s.Git.PushToPeer(repoName, targetGitAddress); err != nil {
		logger.Warnf("Failed to push changes to peer: %s. Error: %v", peerID, err)
		return err
	}

	logger.Debugf("Successfully pushed changes to peer: %s", peerID)

	return nil
}

func (s *RepositoryService) GetLeaderUrl(ctx context.Context, req *pb.LeaderUrlRequest) (*pb.LeaderUrlResponse, error) {
	storedLeaderID, err := dhtutil.GetLeaderFromDHT(context.Background(), s.Peer, req.Name)
	if err == nil {
		addresses, err := s.checkPeerAlive(ctx, storedLeaderID)
		// If current leader alive, just return it
		if err == nil {
			logger.Debugf("Leader from DHT with ID %s is alive! Returning response", storedLeaderID.String())
			return createLeaderUrlResponse(addresses, req.Name), nil
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
	newLeaderAddresses, err := util.GetPeerAdressesFromID(leaderID, s.Peer)

	if err != nil {
		return nil, err
	}

	return createLeaderUrlResponse(newLeaderAddresses, req.Name), nil
}

func createLeaderUrlResponse(addresses *p2p.PeerAddresses, repoName string) *pb.LeaderUrlResponse {
	return &pb.LeaderUrlResponse{
		Success:        true,
		Name:           repoName,
		GitRepoAddress: fmt.Sprintf("%s/%s", addresses.GitAddress, repoName),
		GrpcAddress:    addresses.GrpcAddress,
	}
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
			resp, err := s.getLeaderFromPeer(ctx, peerID, request)

			if err != nil {
				logger.Warnf("getLeaderFromRepoPeers: failed to get leader from peer with ID %s. Error: %v", peerID.String(), err)
				continue
			}

			return resp, nil

		}
	}

	// If no other peers alive. Call my local. This just makes sure that the dht gets updated etc.
	return s.PeerElectionService.GetCurrentLeader(ctx, request)
}

func (s *RepositoryService) getLeaderFromPeer(ctx context.Context, peerID peer.ID, req *pb.CurrentLeaderRequest) (*pb.CurrentLeaderResponse, error) {
	addresses, err := util.GetPeerAdressesFromID(peerID, s.Peer)
	if err != nil {
		logger.Warnf("Failed to get peer address to get current leader. Error: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addresses.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Warnf("Failed to connect to peer: %s to get current leader. Error: %v", peerID, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewElectionClient(conn)
	resp, err := client.GetCurrentLeader(ctx, req)
	if err != nil {
		logger.Warnf("Failed to get current leader with error: %v", err)
		return nil, err
	}
	return resp, nil
}

func (s *RepositoryService) Pull(ctx context.Context, req *pb.RepoPullRequest) (*pb.RepoPullResponse, error) {
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo in DHT: %v", err)
	}

	results := make(chan *pb.RepoPullResponse)
	errors := make(chan error)
	for _, replica := range repo.InSyncReplicas {
		go func(pid peer.ID) {
			if pid == s.Peer.Host.ID() {
				address, err := util.ExtractIPAddr(s.Peer.Host.Addrs()[0].String())
				if err != nil {
					errors <- fmt.Errorf("failed to extract IP address: %v", err)
					return
				}
				results <- &pb.RepoPullResponse{
					Success:     true,
					RepoAddress: fmt.Sprintf("git://%s:%d/%s", address, s.Peer.GitDaemonPort, req.Name),
				}
			} else {
				resp, err := s.getPullResponseFromPeer(ctx, pid, req.Name)
				if err != nil {
					errors <- err
				} else {
					results <- resp
				}
			}
		}(replica)
	}

	numReplicas := len(repo.InSyncReplicas)
	completed := 0
	var errList []error
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-results:
			return result, nil
		case err := <-errors:
			errList = append(errList, err)
			completed++
			// Only return an error if all peers failed to provide a response
			if completed == numReplicas {
				return nil, fmt.Errorf("all operations failed: %v", errList)
			}
		}
	}
}

func (s *RepositoryService) getPullResponseFromPeer(ctx context.Context, pid peer.ID, repoName string) (*pb.RepoPullResponse, error) {
	peerAddress, err := util.GetPeerAdressesFromID(pid, s.Peer)
	if err != nil {
		logger.Warnf("Failed to get peer address for pulling. Error: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Warnf("Failed to connect to peer: %s for pulling. Error: %v", pid, err)
		return nil, err
	}
	defer conn.Close()

	return &pb.RepoPullResponse{
		Success:     true,
		RepoAddress: fmt.Sprintf("git://%s:%d/%s", peerAddress.GitAddress, s.Peer.GitDaemonPort, repoName),
	}, nil
}

func (s *RepositoryService) checkPeerAlive(ctx context.Context, peerID peer.ID) (*p2p.PeerAddresses, error) {
	peerAddress, err := util.GetPeerAdressesFromID(peerID, s.Peer)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewElectionClient(conn)

	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	pingReq := &pb.PingRequest{}
	_, err = client.Ping(pingCtx, pingReq)
	if err != nil {
		return nil, err
	}

	return peerAddress, nil
}
