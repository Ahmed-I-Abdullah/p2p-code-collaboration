// Package p2p implements functionalities for peer-to-peer communication
package p2p

import "github.com/libp2p/go-libp2p/core/peer"

// PeerInfo represents information about a peer including its ports for gRPC and Git daemon.
type PeerInfo struct {
	GrpcPort      int `json:"grpc_port"`       // Port for gRPC communication
	GitDaemonPort int `json:"git_daemon_port"` // Port for Git daemon
}

// PeerAddresses represents the network addresses of a peer, including its peer ID, gRPC address, and Git address
type PeerAddresses struct {
	ID          peer.ID // Unique identifier for the peer
	GrpcAddress string  // Network address for gRPC communication
	GitAddress  string  // Network address for Git communication
}

// RepositoryPeers represents a set of peers associated with a repository
// including peer IDs, in-sync replicas, and version information
type RepositoryPeers struct {
	PeerIDs        []peer.ID `json:"peerIDs"`        // IDs of the peers storing the repository
	InSyncReplicas []peer.ID `json:"inSyncReplicas"` // IDs of peers that are in sync with leader
	Version        int       `json:"version"`        // Version of the repository
}
