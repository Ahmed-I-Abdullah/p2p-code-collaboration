package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"io/ioutil"
	"time"

	constants "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/database"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
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
	BootstrapPeers   flags.AddrList
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

	validator := record.NamespacedValidator{
		"pk":   record.PublicKeyValidator{},
		"ipns": ipns.Validator{},
		"repo": RepoValidator{},
	}

	dhtOptions := []dht.Option{
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix(constants.DHTRepoPrefix),
		dht.Validator(validator),
	}

	kademliaDHT, err := dht.New(ctx, host, dhtOptions...)

	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if config.IsBootstrap {
		logger.Info("This node is a bootstrap node.")
		select {}
	}

	host.SetStreamHandler(constants.PeerPortsProtocol, func(s network.Stream) {
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

	peer := &Peer{
		Host:             host,
		DHT:              kademliaDHT,
		RoutingDiscovery: routingDiscovery,
		GrpcPort:         config.GrpcPort,
		GitDaemonPort:    config.GitDaemonPort,
		BootstrapPeers:   config.BootstrapPeers,
	}

	peer.connectToPeers(ctx, routingDiscovery, host, config.RendezvousString)

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

// func connectToPeers(ctx context.Context, routingDiscovery *drouting.RoutingDiscovery, host host.Host, rendezvous string) {
// 	go func() {
// 		logger.Info("Searching for other peers...")

// 		peerChan, err := routingDiscovery.FindPeers(ctx, rendezvous)
// 		if err != nil {
// 			logger.Errorf("failed to find peers: %v", err)
// 			return
// 		}

// 		for foundPeer := range peerChan {
// 			// Skip the peer if it is the host itself.
// 			if foundPeer.ID == host.ID() {
// 				continue
// 			}

// 			err := host.Connect(ctx, foundPeer)
// 			if err != nil {
// 				logger.Errorf("failed to connect to peer %s: %v", foundPeer.ID, err)
// 			} else {
// 				logger.Infof("Connected to peer: %s", foundPeer.ID)
// 			}
// 		}
// 	}()
// }

func (p *Peer) connectToPeers(ctx context.Context, routingDiscovery *drouting.RoutingDiscovery, host host.Host, rendezvous string) {

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

					err := database.DBCon.Update(func(txn *badger.Txn) error {
						item, err := txn.Get([]byte(foundPeer.ID))
						if err == nil {
							item.Value(func(val []byte) error {
								logger.Debugf(" entry for peerID %v already exists. Here is the data: %s", foundPeer.ID, val)
								return nil
							})
							return nil
						}
						if !errors.Is(err, badger.ErrKeyNotFound) {
							logger.Errorf(" could not get database information: %v", err)
							return err
						}
						peerInfo, err := p.GetPeerPorts(foundPeer.ID)
						value, err := json.Marshal(peerInfo)
						if err != nil {
							logger.Fatalf("failed to marshal peer info")
						}
						updateErr := txn.Set([]byte(foundPeer.ID), value)
						return updateErr
					})
					if err != nil {
						logger.Error(err)
					}
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
	s, err := p.Host.NewStream(ctx, peerID, constants.PeerPortsProtocol)
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

func (p *Peer) GetPeerPortsFromDB(peerID peer.ID) (*PeerInfo, error) {
	var valCpy []byte
	err := database.DBCon.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(peerID))
		if err != nil {
			return err
		}
		valCpy, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Errorf("could not get key %v : %v", peerID, err)
		return nil, err
	}
	var info PeerInfo
	err = json.Unmarshal(valCpy, &info)
	if err != nil {
		logger.Errorf("Error unmarshaling peer info: %v", err)
		return nil, fmt.Errorf("Error unmarshaling peer info: %w", err)
	}
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

func (p *Peer) IsBootstrapNode(peerID peer.ID) bool {
	for _, addr := range p.BootstrapPeers {
		addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}

		if addrInfo.ID == peerID {
			return true
		}
	}

	// If no match is found in the loop, the peer is not a bootstrap peer
	return false
}

func (p *Peer) storeInDHT(ctx context.Context, key string, data []byte) error {
	logger.Infof("Storing data in DHT with key %s", key)

	err := p.DHT.PutValue(ctx, key, data)

	if err != nil {
		logger.Errorf("Failed to store value with key %s in DHT", key)
	}

	return err
}
