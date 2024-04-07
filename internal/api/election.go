package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	util "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils/dhtutil"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

type ElectionService struct {
	pb.UnimplementedElectionServer
	Peer              *p2p.Peer
	electionInProcess map[string]bool
	mu                sync.Mutex
}

func (s *ElectionService) GetCurrentLeader(ctx context.Context, req *pb.CurrentLeaderRequest) (*pb.CurrentLeaderResponse, error) {
	s.mu.Lock()
	electionInProgress, ok := s.electionInProcess[req.RepoName]
	s.mu.Unlock()

	if electionInProgress {
		logger.Debugf("Election in progress for repository %s", req.RepoName)
		return nil, fmt.Errorf("Cannot get leader. Election in progress for repository %s", req.RepoName)
	}

	if !ok {
		logger.Debugf("Starting a new election for repository %s", req.RepoName)
		// Start a new election if no election data exists for the given repository
		electionResult := make(chan string, 1)
		go s.startElection(req.RepoName, electionResult)

		select {
		case newLeader := <-electionResult:
			logger.Debugf("Received election result for repository %s. Leader: %s", req.RepoName, newLeader)
			return &pb.CurrentLeaderResponse{LeaderId: newLeader}, nil
		case <-time.After(30 * time.Second):
			logger.Debugf("Leader election failed for repository %s", req.RepoName)
			return nil, fmt.Errorf("Leader election failed for repository %s", req.RepoName)
		}
	}

	leader, err := s.checkLeaderAlive(ctx, req.RepoName)
	if err != nil {
		logger.Debugf("Starting a new election for repository %s as current leader is found to be not alive", req.RepoName)
		// Start a new election if current leader is found to be not alive
		electionResult := make(chan string, 1)
		go s.startElection(req.RepoName, electionResult)

		select {
		case leader = <-electionResult:
			logger.Debugf("Received election result for repository %s. Leader: %s", req.RepoName, leader)
			return &pb.CurrentLeaderResponse{LeaderId: leader}, nil
		case <-time.After(30 * time.Second):
			logger.Debugf("Leader election failed for repository %s", req.RepoName)
			return nil, fmt.Errorf("Leader election failed for repository %s", req.RepoName)
		}
	}

	return &pb.CurrentLeaderResponse{LeaderId: leader}, nil
}

func (s *ElectionService) Election(ctx context.Context, req *pb.ElectionRequest) (*pb.ElectionResponse, error) {
	// Check for ongoing election
	s.mu.Lock()
	if _, ok := s.electionInProcess[req.RepoName]; ok {
		s.mu.Unlock()
		logger.Debugf("Election in progress for repository %s", req.RepoName)
		return nil, fmt.Errorf("election in progress")
	}

	// If I have a lower id then don't participate in the election
	if s.Peer.Host.ID().String() < req.NodeId {
		logger.Debugf("Skipping election as node ID %s is lower than %s", s.Peer.Host.ID().String(), req.NodeId)
		return &pb.ElectionResponse{Type: pb.ElectionType_OTHER.String()}, nil
	}

	// Set flag for ongoing election
	s.electionInProcess[req.RepoName] = true
	s.mu.Unlock()

	// Spawning a new goroutine for election process
	electionResult := make(chan string, 1)
	go s.startElection(req.RepoName, electionResult)

	select {
	case leader := <-electionResult:
		logger.Debugf("Received election result for repository %s. Leader: %s", req.RepoName, leader)
		return &pb.ElectionResponse{NewLeaderId: leader}, nil
	case <-time.After(30 * time.Second):
		logger.Debugf("Election timeout for repository %s", req.RepoName)
		return nil, fmt.Errorf("timeout, election took too long")
	}
}

func (s *ElectionService) startElection(repoName string, electionResult chan<- string) {
	defer func() {
		// Clean up flag after election
		s.mu.Lock()
		delete(s.electionInProcess, repoName)
		s.mu.Unlock()
	}()

	logger.Debugf("Starting election for repository %s", repoName)

	dhtRecord, err := dhtutil.GetRepoInDHT(context.Background(), s.Peer, repoName)
	if err != nil {
		logger.Debugf("Error getting DHT record for repository %s: %v", repoName, err)
		return
	}

	allPeerAddresses := make(map[string]*p2p.PeerAddresses)

	for _, p := range dhtRecord.PeerIDs {
		if p != s.Peer.Host.ID() {
			addresses, err := util.GetPeerAdressesFromId(p, s.Peer)
			if err != nil {
				logger.Debugf("Failed to get addresses for peer with ID %s for leader election. Error: %v", p.String(), err)
			} else {
				logger.Debugf("Got addresses for peer with ID %s: %v", p.String(), addresses)
				allPeerAddresses[p.String()] = addresses
			}
		}
	}

	var highIdPeers []*p2p.PeerAddresses
	for _, p := range dhtRecord.InSyncReplicas {
		if p.String() > s.Peer.Host.ID().String() {
			if addresses, exists := allPeerAddresses[p.String()]; exists {
				highIdPeers = append(highIdPeers, addresses)
			}
		}
	}

	logger.Debugf("Found %d higher ID peers for repository %s", len(highIdPeers), repoName)

	if len(highIdPeers) == 0 {
		logger.Debugf("No peer with higher IDs. Announcing myself (%s) as a leader", s.Peer.Host.ID().String())
		// If no higher ID peers, I am the big boss
		electionResult <- s.Peer.Host.ID().Pretty()
		return
	}

	electionCh := make(chan string, len(highIdPeers))

	// Send election messages to peers with higher IDs.
	for _, p := range highIdPeers {
		go func(peerAddress *p2p.PeerAddresses) {
			logger.Debugf("Contacting peer with higher ID for election: %s", peerAddress.ID.String())
			s.contactPeerForElection(peerAddress, repoName, electionCh)
		}(p)
	}

	leader := ""
	for i := 0; i < len(highIdPeers); i++ {
		select {
		case peerId := <-electionCh:
			logger.Debugf("Received response from peer: %s", peerId)
			if peerId != "" {
				leader = peerId
			}
		case <-time.After(5 * time.Second): // Timeout if no message received within 5 seconds
			logger.Debugf("Election message timeout from peer %d", i)
			break
		}
	}

	if leader == "" {
		logger.Debugf("No leader elected. Current node (%s) is the leader", s.Peer.Host.ID().String())
		// If no leader is elected, then current node is the leader
		leader = s.Peer.Host.ID().String()
	} else {
		logger.Debugf("Elected leader: %s", leader)
	}

	// Store the new leader for the repository in DHT
	leaderID, err := peer.Decode(leader)

	if err != nil {
		logger.Debugf("Error decoding leader ID: %v", err)
		return
	}

	storedLeaderID, err := dhtutil.GetLeaderFromDHT(context.Background(), s.Peer, repoName)
	if err != nil {
		logger.Debugf("Error getting stored leader from DHT: %v", err)
		err = dhtutil.StoreLeaderInDHT(context.Background(), s.Peer, repoName, leaderID)
		if err != nil {
			logger.Debugf("Error storing leader in DHT: %v", err)
			return
		}
	}

	if err == nil {
		if storedLeaderID != leaderID {
			err = dhtutil.StoreLeaderInDHT(context.Background(), s.Peer, repoName, leaderID)
			if err != nil {
				logger.Debugf("Error storing leader in DHT: %v", err)
				return
			}
		}
	}

	logger.Debugf("Announcing leader (%s) to other peers", leader)
	s.announceLeader(repoName, leader, allPeerAddresses)
	electionResult <- leader
}

func (s *ElectionService) announceLeader(repoName string, leader string, peers map[string]*p2p.PeerAddresses) {
	for _, peerAddress := range peers {
		if peerAddress.ID.Pretty() != leader {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure())
			if err != nil {
				continue
			}
			client := pb.NewElectionClient(conn)
			client.LeaderAnnouncement(ctx, &pb.LeaderAnnouncementRequest{
				LeaderId: leader,
				RepoName: repoName,
			})
			conn.Close()
		}
	}
}

func (s *ElectionService) contactPeerForElection(addresses *p2p.PeerAddresses, repoName string, ch chan<- string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addresses.GrpcAddress, grpc.WithInsecure())

	if err != nil {
		ch <- ""
		return
	}
	defer conn.Close()

	client := pb.NewElectionClient(conn)
	resp, err := client.Election(ctx, &pb.ElectionRequest{
		RepoName: repoName,
		NodeId:   s.Peer.Host.ID().Pretty(),
	})

	if err != nil || resp.NewLeaderId == "" {
		ch <- ""
	} else {
		ch <- resp.NewLeaderId
	}
}

func (s *ElectionService) LeaderAnnouncement(ctx context.Context, req *pb.LeaderAnnouncementRequest) (*pb.LeaderAnnouncementResponse, error) {
	s.mu.Lock()
	s.electionInProcess[req.RepoName] = false
	s.mu.Unlock()

	return &pb.LeaderAnnouncementResponse{}, nil
}

func (s *ElectionService) checkLeaderAlive(ctx context.Context, repoName string) (string, error) {
	leaderID, err := dhtutil.GetLeaderFromDHT(ctx, s.Peer, repoName)
	if err != nil {
		// This error indicates that there is no leader saved in the database for the repo.
		return "", err
	}

	peerAddress, err := util.GetPeerAdressesFromId(leaderID, s.Peer)

	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, peerAddress.GrpcAddress, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return leaderID.String(), nil
}
