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

func GetPeerAdressesFromID(peerID peer.ID, peer *p2p.Peer) (*p2p.PeerAddresses, error) {
	if peerID == peer.Host.ID() {
		address, _ := ExtractIPAddr(peer.Host.Addrs()[0].String())
		return &p2p.PeerAddresses{
			ID:          peerID,
			GitAddress:  fmt.Sprintf("git://%s:%d", address, peer.GitDaemonPort),
			GrpcAddress: fmt.Sprintf("%s:%d", address, peer.GrpcPort),
		}, nil
	}

	peerAddresses := peer.Host.Peerstore().Addrs(peerID)
	if len(peerAddresses) == 0 {
		return nil, fmt.Errorf("failed to get peer addresses for peer with ID %s from peer store", peerID.String())
	}

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
		ID:          peerID,
		GitAddress:  fmt.Sprintf("git://%s:%d", ipAddress, peerInfo.GitDaemonPort),
		GrpcAddress: fmt.Sprintf("%s:%d", ipAddress, peerInfo.GrpcPort),
	}, nil
}
