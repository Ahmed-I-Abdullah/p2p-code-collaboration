package dhtutil

import (
	"context"
	"encoding/json"
	"fmt"

	constants "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func StoreRepoInDHT(ctx context.Context, peer *p2p.Peer, key string, repo p2p.RepositoryPeers) error {
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	return StoreDataInDHT(ctx, peer, GetDHTPathFromRepoName(key), repoBytes)
}

func GetRepoInDHT(ctx context.Context, peer *p2p.Peer, key string) (*p2p.RepositoryPeers, error) {
	bytes, err := GetDataFromDHT(ctx, peer, GetDHTPathFromRepoName(key))

	if err != nil {
		return nil, err
	}

	var repo p2p.RepositoryPeers
	err = json.Unmarshal(bytes, &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize repo %s data from DHT: %w", key, err)
	}
	return &repo, nil
}

func StoreLeaderInDHT(ctx context.Context, peer *p2p.Peer, key string, leaderID peer.ID) error {
	bytes, err := json.Marshal(leaderID)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	return StoreDataInDHT(ctx, peer, GetDHTLeaderPathFromRepoName(key), bytes)
}

func GetLeaderFromDHT(ctx context.Context, inPeer *p2p.Peer, key string) (peer.ID, error) {
	bytes, err := GetDataFromDHT(ctx, inPeer, GetDHTLeaderPathFromRepoName(key))

	if err != nil {
		return "", err
	}

	var leaderID peer.ID

	err = json.Unmarshal(bytes, &leaderID)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize repo %s data from DHT: %w", key, err)
	}
	return leaderID, nil
}

func StoreDataInDHT(ctx context.Context, peer *p2p.Peer, dhtKey string, data []byte) error {
	if err := peer.DHT.PutValue(ctx, dhtKey, data); err != nil {
		return fmt.Errorf("failed to store repository peers data in DHT: %w", err)
	}
	return nil
}

func GetDataFromDHT(ctx context.Context, peer *p2p.Peer, dhtKey string) ([]byte, error) {
	dataBytes, err := peer.DHT.GetValue(ctx, dhtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize data with key %s from dht: %w", dhtKey, err)
	}

	return dataBytes, err
}

func GetDHTPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constants.DHTRepoPrefix, repoName)
}

func GetDHTLeaderPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constants.DHTLeaderPrefix, repoName)
}
