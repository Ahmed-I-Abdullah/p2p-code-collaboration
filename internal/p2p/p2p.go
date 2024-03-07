package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

var logger = log.Logger("p2p")

type Peer struct {
	Host             host.Host
	DHT              *dht.IpfsDHT
	RoutingDiscovery *drouting.RoutingDiscovery
	GrpcPort         int
	GitDaemonPort    int
	sync.RWMutex
}

func Initialize(config flags.Config) (*Peer, error) {
	log.SetLogLevel("p2p", "info")
	ctx := context.Background()

	host, err := libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...))
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	logger.Infof("Host created. ID: %s", host.ID())
	logger.Infof("Listening on addresses: %v", host.Addrs())

	kademliaDHT, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if config.IsBootstrap {
		logger.Info("This node is a bootstrap node.")
		select {}
	}

	host.SetStreamHandler("/p2p/peerinfo/1.0.0", func(s network.Stream) {
		handlePeerInfoStream(s, config.GrpcPort, config.GitDaemonPort)
	})

	logger.Info("Bootstrapping the DHT")
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}
	logger.Info("DHT bootstrapped successfully.")

	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			logger.Errorf("failed to parse peer address %s: %v", peerAddr, err)
			continue
		}
		if err := host.Connect(ctx, *peerinfo); err != nil {
			logger.Warningf("failed to connect to bootstrap node %s: %v", peerinfo.ID, err)
		} else {
			logger.Infof("Connected to bootstrap node: %s", peerinfo.ID)
		}
	}

	logger.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	logger.Info("Successfully announced!")

	connectToPeers(ctx, routingDiscovery, host, config.RendezvousString)

	peer := &Peer{
		Host:             host,
		DHT:              kademliaDHT,
		RoutingDiscovery: routingDiscovery,
		GrpcPort:         config.GrpcPort,
		GitDaemonPort:    config.GitDaemonPort,
	}

	return peer, nil
}

func handlePeerInfoStream(s network.Stream, grpcPort, gitDaemonPort int) {
	logger.Infof("Handling peer info stream for remote peer: %s", s.Conn().RemotePeer().Pretty())

	info := &PeerInfo{
		GrpcPort:      grpcPort,
		GitDaemonPort: gitDaemonPort,
	}

	infoBytes, err := json.Marshal(info)
	if err != nil {
		logger.Fatalf("Error marshaling info: %v", err)
		return
	}
	logger.Infof("Peer info marshaled successfully")

	_, err = s.Write(infoBytes)
	if err != nil {
		logger.Fatalf("Error writing to stream: %v", err)
		return
	}
	logger.Infof("Peer info sent successfully")
	s.Close()
}

func connectToPeers(ctx context.Context, routingDiscovery *drouting.RoutingDiscovery, host host.Host, rendezvous string) {
	go func() {
		logger.Info("Searching for other peers...")
		for {
			peerChan, err := routingDiscovery.FindPeers(ctx, rendezvous)
			if err != nil {
				logger.Errorf("failed to find peers: %v", err)
				continue
			}

			for foundPeer := range peerChan {
				// Skip the peer if it is the host itself.
				if foundPeer.ID == host.ID() {
					continue
				}

				err := host.Connect(ctx, foundPeer)
				if err != nil {
					logger.Errorf("failed to connect to peer %s: %v", foundPeer.ID, err)
				} else {
					logger.Infof("Connected to peer: %s", foundPeer.ID)
				}
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func (p *Peer) GetPeerPorts(peerID peer.ID) (*PeerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger.Infof("Opening stream to peer: %s", peerID.Pretty())
	s, err := p.Host.NewStream(ctx, peerID, "/p2p/peerinfo/1.0.0")
	if err != nil {
		logger.Errorf("Error opening stream: %v", err)
		return nil, fmt.Errorf("Failed to open stream to peer: %w", err)
	}
	logger.Infof("Stream to peer %s opened successfully", peerID.Pretty())

	if ctx.Err() != nil {
		logger.Errorf("Context deadline exceeded while waiting for peer response")
		return nil, ctx.Err()
	}

	logger.Infof("Reading from stream...")
	infoBytes, err := ioutil.ReadAll(s)
	if err != nil {
		logger.Errorf("Error reading from stream: %v", err)
		return nil, fmt.Errorf("Failed to read stream from peer: %w", err)
	}
	logger.Infof("Read from stream successful")

	logger.Infof("Info bytes read: %s", infoBytes)

	var info PeerInfo
	logger.Infof("Unmarshaling peer info...")
	err = json.Unmarshal(infoBytes, &info)
	if err != nil {
		logger.Errorf("Error unmarshaling peer info: %v", err)
		return nil, fmt.Errorf("Error unmarshaling peer info: %w", err)
	}
	logger.Infof("Unmarshaling peer info successful")

	logger.Infof("Peer %s is listening on gRPC port %s and git Daemon port %s", peerID.Pretty(), info.GrpcPort, info.GitDaemonPort)

	return &info, nil
}

func (p *Peer) IsPeerUp(peerId peer.ID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.Host.Network().DialPeer(ctx, peerId)
	return err == nil
}

func (p *Peer) GetPeers() []peer.ID {
	return p.Host.Peerstore().Peers()
}
