// Package p2p provides functionality for peer-to-peer communication and operations
package p2p

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

// LeaderIDValidator provides methods for validating and selecting leader peer IDs
type LeaderIDValidator struct{}

// Validate validates the leader peer ID
// It checks if the provided value is a valid JSON-encoded peer.ID string
func (LeaderIDValidator) Validate(key string, value []byte) error {
	var leaderID string
	if err := json.Unmarshal(value, &leaderID); err != nil {
		return fmt.Errorf("data is not a valid peer.ID string: %w", err)
	}

	if _, err := peer.Decode(leaderID); err != nil {
		return fmt.Errorf("invalid peer.ID format: %w", err)
	}

	return nil
}

// Select selects the leader peer ID from the provided list of values
// It returns the index of the first value or an error if no valid leader ID is found
func (LeaderIDValidator) Select(key string, values [][]byte) (int, error) {
	if len(values) > 0 {
		return 0, nil
	}
	return -1, errors.New("no valid leader ID found")
}
