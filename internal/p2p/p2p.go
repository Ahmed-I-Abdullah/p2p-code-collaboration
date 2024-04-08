// Package p2p provides functionality for peer-to-peer communication and operations
package p2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"

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

// Peer represents a peer in the P2P network.
type Peer struct {
	Host             host.Host                  // The libp2p host
	DHT              *dht.IpfsDHT               // The DHT (Distributed Hash Table) for peer discovery
	RoutingDiscovery *drouting.RoutingDiscovery // The routing discovery service
	GrpcPort         int                        // The gRPC server port
	GitDaemonPort    int                        // The Git Daemon port
	BootstrapPeers   flags.AddrList             // List of bootstrap peers' multiaddresses
}

// Initialize initializes a new peer with the provided configuration
func Initialize(config flags.Config) (*Peer, error) {
	log.SetLogLevel("p2p", "info")
	ctx := context.Background()

	priv, err := RetrievePrivateKey(config.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	// Initialize a new libp2p host
	host, err := libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...), libp2p.Identity(priv))
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	logger.Infof("Host created. ID: %s", host.ID())
	logger.Infof("Listening on addresses: %v", host.Addrs())

	// Register default dht validators and custom validators for repo and leader namespace
	validator := record.NamespacedValidator{
		"pk":     record.PublicKeyValidator{},
		"ipns":   ipns.Validator{},
		"repo":   RepoValidator{},
		"leader": LeaderIDValidator{},
	}

	dhtOptions := []dht.Option{
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix(constants.DHTRepoPrefix),
		dht.ProtocolPrefix(constants.DHTLeaderPrefix),
		dht.Validator(validator),
	}

	kademliaDHT, err := dht.New(ctx, host, dhtOptions...)

	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// If the peer is a bootstrap node, no need to carryout discovery or the next steps
	if config.IsBootstrap {
		logger.Info("This node is a bootstrap node.")
		select {}
	}

	// Add stream handler for peer ports protocol (protocol handling retrieval of peer ports)
	host.SetStreamHandler(constants.PeerPortsProtocol, func(s network.Stream) {
		handlePeerInfoStream(s, config.GrpcPort, config.GitDaemonPort)
	})

	// Bootstrap the DHT
	logger.Info("Bootstrapping the DHT")
	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}
	logger.Info("DHT bootstrapped successfully.")

	// Connect to all boostrap peers
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

	// Announce peer's presence via the DHT
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

	// Discover peers and connect to them
	peer.connectToPeers(ctx, routingDiscovery, host, config.RendezvousString)

	return peer, nil
}

// handlePeerInfoStream handles the peer info stream for remote peers
// It receives a network stream (s), GRPC port, and Git daemon port
// marshals the peer info into JSON format, and writes it to the stream
func handlePeerInfoStream(s network.Stream, grpcPort, gitDaemonPort int) {
	logger.Infof("Handling peer info stream for remote peer: %s", s.Conn().RemotePeer().Pretty())

	// Create a PeerInfo struct with GRPC port and Git daemon port
	info := &PeerInfo{
		GrpcPort:      grpcPort,
		GitDaemonPort: gitDaemonPort,
	}

	// Marshal the PeerInfo struct into JSON format
	infoBytes, err := json.Marshal(info)
	if err != nil {
		logger.Fatalf("Error marshaling info: %v", err)
		return
	}
	logger.Infof("Peer info marshaled successfully")

	// Write the marshaled info bytes to the stream
	_, err = s.Write(infoBytes)
	if err != nil {
		logger.Fatalf("Error writing to stream: %v", err)
		return
	}
	logger.Infof("Peer info sent successfully")
	s.Close()
}

// connectToPeers establishes connections with other peers using the routing discovery
// It continuously searches for new peers in the network and attempts to connect to them
// It uses the provided context, routingDiscovery, host, and rendezvous string for the peer discovery process
func (p *Peer) connectToPeers(ctx context.Context, routingDiscovery *drouting.RoutingDiscovery, host host.Host, rendezvous string) {
	go func() {
		logger.Info("Searching for other peers...")
		for {
			// Find peers using routing discovery
			peerChan, err := routingDiscovery.FindPeers(ctx, rendezvous)
			if err != nil {
				logger.Errorf("failed to find peers: %v", err)
				continue
			}

			for foundPeer := range peerChan {
				// Skip the peer if it is the host itself, so the peer doesnot dial itself
				if foundPeer.ID == host.ID() {
					continue
				}

				// Attempt to connect to the found peer
				err := host.Connect(ctx, foundPeer)
				if err != nil {
					logger.Errorf("failed to connect to peer %s: %v", foundPeer.ID, err)
				} else {
					logger.Debugf("Connected to peer: %s", foundPeer.ID)

					// Check if ports information for the peer exists in the database
					ports, err := database.Get([]byte(foundPeer.ID))
					if errors.Is(err, badger.ErrKeyNotFound) {
						// If ports information not found, retrieve it directly from the peer and store it in the database
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
						logger.Errorf("could not retrieve ports for %v from db. Err -> %v", foundPeer.ID, err)
					} else {
						logger.Debugf("entry for peerID %v already exists. Here is the data: %s", foundPeer.ID, ports)
					}
				}
			}

			// Wait for a ten seconds before searching for peers again
			time.Sleep(10 * time.Second)
		}
	}()
}

// RetrievePrivateKey retrieves or generates an Ed25519 private key from the specified file path
// If the private key file does not exist, a new private key is generated and saved to the file
// Otherwise, the private key is loaded from the file
func RetrievePrivateKey(privKeyPath string) (crypto.PrivKey, error) {
	// Check if the private key file exists
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		// Generate a new Ed25519 private key
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

		// Write the private key bytes to the file
		err = ioutil.WriteFile(privKeyPath, bytes, os.ModePerm)
		if err != nil {
			logger.Error("Failed to write private key to file: ", err)
			return nil, err
		}

		logger.Info("Generated and saved a new private key")
		return priv, nil
	} else {
		// Read the private key bytes from the file
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

// getPeerPortsRaw opens a stream to a peer and retrieves raw peer info bytes
// It returns the raw info bytes or an error if stream handling fails
func (p *Peer) getPeerPortsRaw(peerID peer.ID) ([]byte, error) {
	// Create a context with a one second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger.Infof("Opening stream to peer: %s", peerID.Pretty())

	// Open a stream to the peer with the peer ports protocol
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

// GetPeerPortsDirectly gets peer ports directly from a peer by opening a stream and retrieving raw info bytes
// It then saves the raw info bytes to the database and returns the peer info or an error if any operation fails
func (p *Peer) GetPeerPortsDirectly(peerID peer.ID) (*PeerInfo, error) {
	// Get raw peer ports
	infoBytes, err := p.getPeerPortsRaw(peerID)
	if err != nil {
		return nil, err
	}

	// Try storing the ports data in BadgerDB
	err = database.Put([]byte(peerID), infoBytes)
	if err != nil {
		logger.Warnf("Could not save peer ports with id %v to the database")
	}
	var info PeerInfo

	// Unmarshal bytes into peerInfo struct
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

// GetPeerPortsFromDB retrieves peer ports from the database based on the peer ID
// If the ports are not found in the database, it retrieves them directly using GetPeerPortsDirectly
// It returns the peer info or an error if any operation fails
func (p *Peer) GetPeerPortsFromDB(peerID peer.ID) (*PeerInfo, error) {
	val, err := database.Get([]byte(peerID))

	// If the ports are not stored in the database, get them directly
	if errors.Is(err, badger.ErrKeyNotFound) {
		return p.GetPeerPortsDirectly(peerID)
	}

	if err != nil {
		logger.Errorf("could not get key %v : %v", peerID, err)
		return nil, err
	}

	var info PeerInfo
	// Unmarshal bytes into a PeerInfo struct
	err = json.Unmarshal(val, &info)
	if err != nil {
		logger.Errorf("Error unmarshaling peer info: %v", err)
		return nil, fmt.Errorf("Error unmarshaling peer info: %w", err)
	}
	return &info, nil
}

// IsPeerUp checks if a peer with the given peer ID is reachable and active
// It returns true if the peer is up and reachable, otherwise false
func (p *Peer) IsPeerUp(peerId peer.ID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.Host.Network().DialPeer(ctx, peerId)
	return err == nil
}

// GetPeers returns a list of peer IDs known by the peer's peerstore.
func (p *Peer) GetPeers() []peer.ID {
	return p.Host.Peerstore().Peers()
}

// IsBootstrapNode checks if the given peer ID corresponds to a bootstrap peer
// It returns true if the peer ID matches any of the bootstrap peers, otherwise false
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

// storeInDHT stores data in the DHT with the specified key
// It returns an error if the operation fails
func (p *Peer) storeInDHT(ctx context.Context, key string, data []byte) error {
	logger.Infof("Storing data in DHT with key %s", key)

	err := p.DHT.PutValue(ctx, key, data)

	if err != nil {
		logger.Errorf("Failed to store value with key %s in DHT", key)
	}

	return err
}
