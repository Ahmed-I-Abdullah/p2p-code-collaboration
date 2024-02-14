package p2p

import (
	"context"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/cmd/flags"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)


func Initialize(config flags.Config) error {
	logger := log.Logger("p2p")
	log.SetLogLevel("p2p", "info")

	
	host, err := libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...))
	if err != nil {
		return err
	}
	logger.Infof("Host created. We are: %s", host.ID())
	logger.Infof("Listening on addresses: %v", host.Addrs())

	
	ctx := context.Background()
	kademliaDHT, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		return err
	}

	if config.IsBootstrap {
		logger.Info("This node is a bootstrap node.")
		select {}
	}

	
	logger.Info("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return err
	}
	logger.Info("DHT bootstrapped successfully.")

	
	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		if err := host.Connect(ctx, *peerinfo); err != nil {
			logger.Warning(err)
		} else {
			logger.Infof("Connected to bootstrap node: %s", peerinfo.ID)
		}
	}

	
	logger.Info("Announcing ourselves...")
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)
	logger.Info("Successfully announced!")

	
	for {
		logger.Info("Searching for other peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
		if err != nil {
			return err
		}

		for peer := range peerChan {
			if peer.ID == host.ID() {
				continue
			}
			logger.Infof("Found peer: %s", peer.ID)
		}
		
		time.Sleep(3 * time.Second)
	}

	return nil
}


func handleStream(stream network.Stream) {
	logger := log.Logger("p2p")
	logger.Info("Got a new stream!")
	
}
