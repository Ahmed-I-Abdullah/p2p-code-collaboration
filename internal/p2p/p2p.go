package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"io/ioutil"
	"os"
	"time"

	constants "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/database"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
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
	log.SetLogLevel("p2p", "debug")
	ctx := context.Background()

	priv, err := RetrievePrivateKey(config.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...), libp2p.Identity(priv))
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
					logger.Infof("Connected to peer: %s", foundPeer.ID)

					ports, err := database.Get([]byte(foundPeer.ID))

					if errors.Is(err, badger.ErrKeyNotFound) {
						infoBytes, err := p.getPeerPortsRaw(foundPeer.ID)
						if err != nil {
							logger.Errorf("failed to get ports for peer %v : %v", foundPeer.ID, err)
						} else {
							err = database.Put([]byte(foundPeer.ID), infoBytes)
							if err != nil {
								logger.Errorf("failed to write ports to db for peer %v : %v", foundPeer.ID, err)
							}
							logger.Infof("port info for peer %v has been saved", foundPeer.ID)
						}
					} else if err != nil {
						logger.Errorf("could not retrive ports for %v from db. Err -> %v", foundPeer.ID, err)
					} else {
						logger.Debugf(" entry for peerID %v already exists. Here is the data: %s", foundPeer.ID, ports)
					}
				}
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func RetrievePrivateKey(privKeyPath string) (crypto.PrivKey, error) {
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			logger.Error("Failed for generate private key")
			return nil, err
		}

		bytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			logger.Error("Failed to marshal private key: ", err)
			return nil, err
		}

		err = ioutil.WriteFile(privKeyPath, bytes, os.ModePerm)
		if err != nil {
			logger.Error("Failed to write private key to file: ", err)
			return nil, err
		}

		logger.Info("Generated and saved a new private key")
		return priv, nil

	} else {
		bytes, err := ioutil.ReadFile(privKeyPath)
		if err != nil {
			logger.Error("Failed to read private key from file: ", err)
			return nil, err
		}

		priv, err := crypto.UnmarshalPrivateKey(bytes)
		if err != nil {
			logger.Error("Failed to unmarshal private key: ", err)
			return nil, err
		}

		logger.Info("Loaded private key from file")
		return priv, nil
	}
}

func (p *Peer) getPeerPortsRaw(peerID peer.ID) ([]byte, error) {
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

	return infoBytes, nil
}

// GetPeerPortsDirectly gets peer ports directly from peer and saves it to the database
func (p *Peer) GetPeerPortsDirectly(peerID peer.ID) (*PeerInfo, error) {
	infoBytes, err := p.getPeerPortsRaw(peerID)
	if err != nil {
		return nil, err
	}
	err = database.Put([]byte(peerID), infoBytes)
	if err != nil {
		logger.Warnf("Could not save peer ports with id %v to the database")
	}
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
	val, err := database.Get([]byte(peerID))
	if errors.Is(err, badger.ErrKeyNotFound) {
		return p.GetPeerPortsDirectly(peerID)
	}
	if err != nil {
		logger.Errorf("could not get key %v : %v", peerID, err)
		return nil, err
	}
	var info PeerInfo
	err = json.Unmarshal(val, &info)
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
