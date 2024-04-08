// Package constants provides various constants used within the peer-to-peer code collaboration system
package constants

import "github.com/libp2p/go-libp2p/core/protocol"

// PeerPrivateKeyPath represents the default path to store the private key of the peer
var PeerPrivateKeyPath = "./.priv_key"

// PeerPortsProtocol represents the protocol used for identifying peer ports information in the network
var PeerPortsProtocol protocol.ID = "/p2p/peerinfo/1.0.0"

// DefaultReposDirectory represents the default directory path for storing repositories
var DefaultReposDirectory = "./repos"

// DHTRepoPrefix represents the prefix used in DHT (Distributed Hash Table) for repository-related entries
var DHTRepoPrefix protocol.ID = "/repo"

// DHTLeaderPrefix represents the prefix used in DHT (Distributed Hash Table) for leader-related entries
var DHTLeaderPrefix protocol.ID = "/leader"

// RepositoryReplicationFactor represents the number of replicas for each repository in the network when initializing a new repo
var RepositoryReplicationFactor = 4
