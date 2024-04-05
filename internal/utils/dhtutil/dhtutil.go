package dhtutil

import (
	"context"
	"encoding/json"
	"fmt"

	constants "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
)

func StoreRepoInDHT(ctx context.Context, peer *p2p.Peer, key string, repo p2p.RepositoryPeers) error {
	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return fmt.Errorf("failed to serialize repository peers data: %w", err)
	}

	if err := peer.DHT.PutValue(ctx, GetDHTPathFromRepoName(key), repoBytes); err != nil {
		return fmt.Errorf("failed to store repository peers data in DHT: %w", err)
	}
	return nil
}

func GetRepoInDHT(ctx context.Context, peer *p2p.Peer, key string) (*p2p.RepositoryPeers, error) {
	repoBytes, err := peer.DHT.GetValue(ctx, GetDHTPathFromRepoName(key))
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize repository peers data: %w", err)
	}
	var repo p2p.RepositoryPeers
	err = json.Unmarshal(repoBytes, &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve repository peers data in DHT: %w", err)
	}
	return &repo, nil
}

func GetDHTPathFromRepoName(repoName string) string {
	return fmt.Sprintf("%s/%s", constants.DHTRepoPrefix, repoName)
}
