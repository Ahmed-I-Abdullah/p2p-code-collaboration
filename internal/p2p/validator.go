package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
)

type RepoValidator struct{}

func (RepoValidator) Validate(key string, value []byte) error {
	var repoPeers RepositoryPeers
	if err := json.Unmarshal(value, &repoPeers); err != nil {
		return fmt.Errorf("data is not a valid RepositoryPeers object: %w", err)
	}

	if len(repoPeers.PeerIDs) == 0 {
		return errors.New("at least one PeerID is required")
	}

	return nil
}

func (RepoValidator) Select(key string, value [][]byte) (int, error) { return 0, nil }
