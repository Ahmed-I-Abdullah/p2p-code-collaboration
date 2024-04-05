package api

import (
	"context"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
)

type ElectionService struct {
	pb.UnimplementedElectionServer
	Peer *p2p.Peer
	Git  *gitops.Git
}

func (s *ElectionService) InitiateElection(ctx context.Context, req *pb.ElectionRequest) (*pb.RepoInitResponse, error) {
	if req.NodeId < s.Peer.Host.ID().String() {

	}
}
