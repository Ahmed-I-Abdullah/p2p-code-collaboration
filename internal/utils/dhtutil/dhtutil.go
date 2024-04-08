// package dhtutil provides utility functions for interacting with the Distributed Hash Table (DHT)
package dhtutil

import (
	"context"
	"encoding/json"
	"fmt"

	constants "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

// StoreRepoInDHT stores repository peer information in the DHT under the specified key (repo ID)
// It serializes the repository peers data and stores it in the DHT using the provided peer instance
func StoreRepoInDHT(ctx context.Context, peer *p2p.Peer, key string, repo p2p.RepositoryPeers) error {
	// Serialize repository peers data
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	// Store serialized data in the DHT
	return StoreDataInDHT(ctx, peer, GetDHTPathFromRepoName(key), repoBytes)
}

// GetRepoInDHT retrieves repository peer information from the DHT using the specified key (repo ID)
// It deserializes the data retrieved from the DHT into a RepositoryPeers instance
func GetRepoInDHT(ctx context.Context, peer *p2p.Peer, key string) (*p2p.RepositoryPeers, error) {
	// Retrieve data from the DHT
	bytes, err := GetDataFromDHT(ctx, peer, GetDHTPathFromRepoName(key))
	if err != nil {
		return nil, err
	}

	// Deserialize data into RepositoryPeers instance
	var repo p2p.RepositoryPeers
	err = json.Unmarshal(bytes, &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize repo %s data from DHT: %w", key, err)
	}
	return &repo, nil
}

// StoreLeaderInDHT stores the ID of the leader peer for a repository in the DHT under the specified key (repo ID)
// It serializes the leader peer ID and stores it in the DHT using the provided peer instance.
func StoreLeaderInDHT(ctx context.Context, peer *p2p.Peer, key string, leaderID peer.ID) error {
	// Serialize leader peer ID
	bytes, err := json.Marshal(leaderID)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	// Store serialized data in the DHT
	return StoreDataInDHT(ctx, peer, GetDHTLeaderPathFromRepoName(key), bytes)
}

// GetLeaderFromDHT retrieves the ID of the leader peer for a repository from the DHT using the specified key (repo ID)
// It deserializes the data retrieved from the DHT into a peer.ID instance.
func GetLeaderFromDHT(ctx context.Context, inPeer *p2p.Peer, key string) (peer.ID, error) {
	// Retrieve data from the DHT
	bytes, err := GetDataFromDHT(ctx, inPeer, GetDHTLeaderPathFromRepoName(key))
	if err != nil {
		return "", err
	}

	// Deserialize data into peer.ID instance
	var leaderID peer.ID
	err = json.Unmarshal(bytes, &leaderID)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize repo %s data from DHT: %w", key, err)
	}
	return leaderID, nil
}

// StoreDataInDHT stores the provided data in the DHT under the specified key using the provided peer instance
func StoreDataInDHT(ctx context.Context, peer *p2p.Peer, dhtKey string, data []byte) error {
	if err := peer.DHT.PutValue(ctx, dhtKey, data); err != nil {
		return fmt.Errorf("failed to store repository peers data in DHT: %w", err)
	}
	return nil
}

// GetDataFromDHT retrieves data from the DHT using the specified key and the provided peer instance
func GetDataFromDHT(ctx context.Context, peer *p2p.Peer, dhtKey string) ([]byte, error) {
	// Retrieve data from the DHT
	dataBytes, err := peer.DHT.GetValue(ctx, dhtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data with key %s from dht: %w", dhtKey, err)
	}

	return dataBytes, nil
}

// GetDHTPathFromRepoName generates the DHT key path for storing repository data based on the repository name
func GetDHTPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constants.DHTRepoPrefix, repoName)
}

// GetDHTLeaderPathFromRepoName generates the DHT key path for storing leader peer ID based on the repository name
func GetDHTLeaderPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constants.DHTLeaderPrefix, repoName)
}
