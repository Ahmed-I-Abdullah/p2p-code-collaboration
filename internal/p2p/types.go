package p2p

import "github.com/libp2p/go-libp2p/core/peer"

type PeerInfo struct {
	GrpcPort      int `json:"grpc_port"`
	GitDaemonPort int `json:"git_daemon_port"`
}

type PeerAddresses struct {
	ID          peer.ID
	GrpcAddress string
	GitAddress  string
}

type RepositoryPeers struct {
	PeerIDs        []peer.ID `json:"peerIDs"`
	InSyncReplicas []peer.ID `json:"inSyncReplicas"`
	LeaderID       peer.ID   `json:"leaderID"`
	Version        int       `json:"version"`
}
