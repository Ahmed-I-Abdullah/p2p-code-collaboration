package api

import (
	"context"
	"fmt"
	"net"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
	"google.golang.org/grpc"
)

// StartServer starts a gRPC server for handling peer-to-peer communication
// The server provides services for managing elections and repositories
//
// Context ctx is used for setting timeouts and detecting context cancellation for graceful shutdown of the server
// Parameter peer represents the local peer instance through which the server operates
// Parameter git represents the Git operations interface for managing repositories
//
// It returns an error if the server fails to start or encounters any issues during operation
func StartServer(ctx context.Context, peer *p2p.Peer, git *gitops.Git) error {
	// Listen for incoming connections on the specified port of the peer.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", peer.GrpcPort))
	if err != nil {
		return err
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register the election service with the gRPC server
	electionService := NewElectionService(peer)
	pb.RegisterElectionServer(s, electionService)

	// Register the repository service with the gRPC server
	repositoryService := NewRepositoryService(peer, git, electionService)
	pb.RegisterRepositoryServer(s, repositoryService)

	// Start serving incoming requests on the listener
	return s.Serve(lis)
}
