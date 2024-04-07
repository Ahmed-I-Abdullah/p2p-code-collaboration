package util

import (
	"fmt"
	"strings"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func ExtractIPAddr(address string) (string, error) {
	parts := strings.Split(address, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected format")
	}

	return parts[2], nil
}

func GetPeerAdressesFromId(peerID peer.ID, peer *p2p.Peer) (*p2p.PeerAddresses, error) {
	peerAddresses := peer.Host.Peerstore().Addrs(peerID)
	peerAddress := peerAddresses[0]
	ipAddress, err := ExtractIPAddr(peerAddress.String())
	if err != nil {
		return nil, err
	}
	peerInfo, err := peer.GetPeerPortsFromDB(peerID)
	if err != nil {
		peerInfo, err = peer.GetPeerPortsDirectly(peerID)
		if err != nil {
			return nil, err
		}
	}

	return &p2p.PeerAddresses{
		GitAddress:  fmt.Sprintf("git://%s:%d", ipAddress, peerInfo.GitDaemonPort),
		GrpcAddress: fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort),
	}, nil
}
