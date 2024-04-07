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

func StartServer(ctx context.Context, peer *p2p.Peer, git *gitops.Git) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", peer.GrpcPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()

	// Register git gRPC server and election gRPC server
	electionService := NewElectionService(peer)
	pb.RegisterElectionServer(s, electionService)
	pb.RegisterRepositoryServer(s, NewRepositoryService(peer, git, electionService))

	return s.Serve(lis)
}
