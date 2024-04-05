package api

import (
	"context"
	"sort"

	"fmt"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	util "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/utils/dhtutil"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

type ElectionService struct {
	pb.UnimplementedElectionServer
	Peer *p2p.Peer
}

type PeerInformation struct {
	ID      peer.ID
	Address string
}

type ElectionState struct {
	Leader     string
	InElection bool
}

var repoElectionState map[string]ElectionState

func init() {
	repoElectionState = make(map[string]ElectionState)
}

// InitiateElection implements the bully algorithm
func (s *ElectionService) Election(ctx context.Context, req *pb.ElectionRequest) (*pb.ElectionResponse, error) {
	repoName := req.RepoName
	repo, err := dhtutil.GetRepoInDHT(ctx, s.Peer, req.RepoName)

	if err != nil {
		return nil, err
	}

	peerInformations := make([]PeerInformation, len(repo.PeerIDs))

	for i := range repo.PeerIDs {
		peerID := repo.PeerIDs[i]
		peerInfo := PeerInformation{}
		peerInfo.ID = peerID
		peerAddresses := s.Peer.Host.Peerstore().Addrs(peerID)
		peerAddress := peerAddresses[0]
		ipAddress, err := util.ExtractIPAddr(peerAddress.String())

		if err != nil {
			logger.Warnf("Failed to get IP Address of peer %s in initiate election function", peerID)
			continue
		}

		peerPorts, err := s.Peer.GetPeerPortsFromDB(peerID)

		if err != nil {
			logger.Warnf("Failed to get peer ports for peer: %s", peerID)
			peerPorts, err = s.Peer.GetPeerPortsDirectly(peerID)
			if err != nil {
				logger.Warnf("Failed to get peer ports directly: %s", peerID)
				continue
			}
		}

		peerInfo.Address = fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)
		peerInformations[i] = peerInfo
	}

	// Sort peers in descending order of ID
	sort.Slice(peerInformations, func(i, j int) bool {
		return peerInformations[i].ID > peerInformations[j].ID
	})

	// I am the boss, will announce leader to everyone
	if peerInformations[0].ID == s.Peer.Host.ID() {
		for _, peerInfo := range peerInformations {
			go func(peerAddr string) {
				_, err := s.sendLeaderAnnouncement(ctx, peerInfo, repoName, s.Peer.Host.ID().String())

				if err != nil {
					logger.Errorf("Failed to announce leader to peer %s: %v", peerAddr, err)
				}
			}(peerInfo.Address)
		}

		// Update map to reflect leader
		repoElectionState[repoName] = ElectionState{
			Leader:     s.Peer.Host.ID().String(),
			InElection: false,
		}

	}

	// Check if I am the bully
	if isNodeBully(s.Peer.Host.ID().String(), req.GetNodeId()) {
		fmt.Println("I am the bully")
		bully := s.Peer.Host.ID().String()

		// Update the map to indicate participation in the election
		state := ElectionState{
			Leader:     "",
			InElection: true,
		}
		if _, ok := repoElectionState[repoName]; !ok {
			repoElectionState[repoName] = state
		} else {
			existingState := repoElectionState[repoName]
			existingState.InElection = true
			repoElectionState[repoName] = existingState
		}

		for _, peerInfo := range peerInformations {
			if peerInfo.ID.String() > bully {
				// Launch RPC call in a goroutine
				go func(peerAddr string) {
					resp, err := s.initiateElectionRPC(ctx, peerInfo, bully, repoName)

					if err != nil {
						logger.Errorf("Failed to initiate election on peer %s: %v", peerAddr, err)
						// return
					}

					if resp.Type == pb.ElectionType_BULLY.String() {
						// Bigger guys are alive, I am no longer participating
						existingState := repoElectionState[repoName]
						existingState.InElection = false
						repoElectionState[repoName] = existingState
					}
				}(peerInfo.Address)
			}
		}

		// TODO:
		// Need a way to wait for leader annoncement and if didnt receive reinitate the election
	}

	return &pb.ElectionResponse{Type: pb.ElectionType_OTHER.String()}, nil
}

func (s *ElectionService) LeaderAnnouncement(ctx context.Context, req *pb.LeaderAnnouncementRequest) (*pb.LeaderAnnouncementResponse, error) {
	repoName := req.RepoName
	state := ElectionState{
		Leader:     req.LeaderId,
		InElection: false,
	}

	repoElectionState[repoName] = state

	return &pb.LeaderAnnouncementResponse{}, nil
}

func (s *ElectionService) initiateElectionRPC(ctx context.Context, peerInformation PeerInformation, bully string, repoName string) (*pb.ElectionResponse, error) {
	conn, err := grpc.Dial(peerInformation.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewElectionClient(conn)

	request := &pb.ElectionRequest{
		RepoName: repoName,
		NodeId:   bully,
	}

	resp, err := client.Election(ctx, request)
	if err != nil {
		return resp, err
	}

	if resp.Type == pb.ElectionType_BULLY.String() {
		logger.Infof("Candidate %s, declared bully", &peerInformation.ID)
	}

	return resp, nil
}

func (s *ElectionService) sendLeaderAnnouncement(ctx context.Context, peerInfo PeerInformation, repoName string, leaderID string) (*pb.LeaderAnnouncementResponse, error) {
	conn, err := grpc.Dial(peerInfo.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewElectionClient(conn)

	request := &pb.LeaderAnnouncementRequest{
		RepoName: repoName,
		LeaderId: leaderID,
	}

	return client.LeaderAnnouncement(ctx, request)
}

func isNodeBully(currentNodeId string, incomingNodeId string) bool {
	return incomingNodeId < currentNodeId
}
