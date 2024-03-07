package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api/pb"
	constansts "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
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
		// Initial process remains the same
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
		var errs []error

		for len(successfulPeers) < replicationFactor && len(onlinePeers) > 0 {
			p := onlinePeers[0] // choose a peer
			success, err := s.signalCreateNewRepository(ctx, p, req.Name)

			if success {
				successfulPeers = append(successfulPeers, p)
			} else {
				errs = append(errs, fmt.Errorf("failed processing peer %v: %w", p, err))
			}

			// Regardless of success or failure, we remove the peer from the slice
			onlinePeers = onlinePeers[1:]
		}

		if len(successfulPeers) < replicationFactor {
			return &pb.RepoInitResponse{
				Success: false,
				Message: fmt.Sprintf("Could not create repo on all %v peers, created on %v peers. Errors: %v", replicationFactor, len(successfulPeers), multierr.Combine(errs...)),
			}, fmt.Errorf("Could not create repo on all %v peers, created on %v peers. Errors: %w", replicationFactor, len(successfulPeers), multierr.Combine(errs...))
		}

		repo := p2p.RepositoryPeers{
			PeerIDs:        successfulPeers,
			InSyncReplicas: nil, // TODO: Fix this later
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

	peerPorts, err := s.Peer.GetPeerPortsFromDB(p)
	if err != nil {
		return false, fmt.Errorf("failed to get peer ports: %w", err)
	}

	ipAddress, err := extractIPAddr(peerAddress.String())
	if err != nil {
		return false, fmt.Errorf("failed to get peer IP Adress: %w", err)
	}

	target := fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)

	logger.Infof("Connecting to peer %s at address %s for gRPC communication\n", p, target)
	conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return false, fmt.Errorf("failed to connect: %w", err)
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
