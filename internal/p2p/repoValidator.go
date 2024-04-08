// Package p2p provides functionality for peer-to-peer communication
package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RepoValidator provides methods for validating and selecting repository peer data
type RepoValidator struct{}

// Validate checks if the provided byte slice represents a valid RepositoryPeers object
// It returns an error if the validation fails
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

// Select identifies the index of the RepositoryPeers object with the highest version among multiple values
// It returns the index of the object with the highest version and an error if no valid objects are found
func (RepoValidator) Select(key string, values [][]byte) (int, error) {
	var highestVersionIndex int
	highestVersion := -1

	for i, value := range values {
		var repoPeers RepositoryPeers
		if err := json.Unmarshal(value, &repoPeers); err != nil {
			return -1, fmt.Errorf("data is not a valid RepositoryPeers object: %w", err)
		}

		if repoPeers.Version > highestVersion {
			highestVersion = repoPeers.Version
			highestVersionIndex = i
		}
	}

	if highestVersion == -1 {
		return -1, errors.New("no valid RepositoryPeers objects found")
	}

	return highestVersionIndex, nil
}
