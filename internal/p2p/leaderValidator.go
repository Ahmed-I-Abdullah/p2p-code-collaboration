package p2p

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

type LeaderIDValidator struct{}

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

func (LeaderIDValidator) Select(key string, values [][]byte) (int, error) {
	if len(values) > 0 {
		return 0, nil
	}
	return -1, errors.New("no valid leader ID found")
}
