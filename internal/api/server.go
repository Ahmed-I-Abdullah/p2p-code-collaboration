package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api/pb"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc"
)

type RepositoryService struct {
	pb.UnimplementedRepositoryServer
	Peer *p2p.Peer
}

func (s *RepositoryService) Init(ctx context.Context, req *pb.RepoInitRequest) (*pb.RepoInitResponse, error) {
	// How many peers we are going to store the repository on
	replicationFactor := 2

	logger := log.Logger("grpcService")
	log.SetLogLevel("grpcService", "info")

	logger.Info("Received request to initialize a repository")

	// If the request is directly from the cli, choose random peers, etc
	if req.FromCli {
		var onlinePeers []peer.ID
		allPeers := s.Peer.GetPeers()
		for i := 0; len(onlinePeers) < replicationFactor && i < len(allPeers); i++ {

			// TODO: Do not connect to boostrap
			if allPeers[i] == s.Peer.Host.ID() {
				continue
			}
			logger.Infof("Current peer ID is %s", allPeers[i].Pretty())
			onlinePeers = append(onlinePeers, allPeers[i])
		}

		if len(onlinePeers) != replicationFactor {
			return nil, fmt.Errorf("only found %d online peers", len(onlinePeers))
		}

		for _, p := range onlinePeers {
			peerAddresses := s.Peer.Host.Peerstore().Addrs(p)

			if len(peerAddresses) == 0 {
				return nil, fmt.Errorf("no known addresses for peer %s", p)
			}
			peerAddress := peerAddresses[0]

			peerPorts, err := s.Peer.GetPeerPorts(p)

			if err != nil {
				// Need to handle error if peers didnt create repo..
				logger.Errorf("Failed to get peer ports: %v", err)
				continue
			}

			ipAddress, err := extractIPAddr(peerAddress.String())
			if err != nil {
				logger.Errorf("Failed to get peer IP Adress: %v", err)
				continue
			}

			target := fmt.Sprintf("%s:%d", ipAddress, peerPorts.GrpcPort)

			logger.Infof("Connecting to peer %s at address %s for gRPC communication\n", p, target)
			conn, err := grpc.Dial(target, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Errorf("Failed to connect: %v", err)
			}
			defer conn.Close()

			logger.Infof("Connected to gRPC server of peer %s.", p)

			c := pb.NewRepositoryClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			_, err = c.Init(ctx, &pb.RepoInitRequest{Name: req.Name, FromCli: false})

			if err != nil {
				// Need tom handle error if peers didnt create repo..
				logger.Errorf("Failed to initialize repository: %v", err)
			}
		}

		return &pb.RepoInitResponse{
			Success: true,
			Message: "Repo created",
		}, nil

	} else {
		if req.Name == "" {
			return &pb.RepoInitResponse{
				Success: false,
				Message: "Repository name is empty",
			}, nil
		}

		err := createBareRepository(req.Name)

		if err != nil {
			return &pb.RepoInitResponse{
				Success: false,
				Message: "Failed to create file: " + err.Error(),
			}, nil
		}

		return &pb.RepoInitResponse{
			Success: true,
			Message: "File created successfully",
		}, nil

	}
}

func createBareRepository(repositoryName string) error {

	// For now we just create a file in the current directory
	// TODO:
	// 1- Get repository storage path from config
	// 2- Run git command to create bare repositroy there
	f, err := os.Create(repositoryName)
	if err != nil {
		return err
	}

	defer f.Close()

	return nil
}

func extractIPAddr(address string) (string, error) {
	parts := strings.Split(address, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected format")
	}

	return parts[2], nil
}

func StartServer(ctx context.Context, peer *p2p.Peer) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", peer.GrpcPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()

	pb.RegisterRepositoryServer(s, &RepositoryService{Peer: peer})

	return s.Serve(lis)
}
