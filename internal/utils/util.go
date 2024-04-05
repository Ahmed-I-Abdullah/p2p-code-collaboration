package util

import (
	"fmt"
	"strings"
)

func ExtractIPAddr(address string) (string, error) {
	parts := strings.Split(address, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected format")
	}

	return parts[2], nil
}
