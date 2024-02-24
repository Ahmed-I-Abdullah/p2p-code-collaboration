package p2p

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ipfs/go-log/v2"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/cmd/flags"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
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

	// Set a function as stream handler. This function is called when a peer
	// initiates a connection and starts a stream with this peer.
	host.SetStreamHandler(protocol.ID("/chat/1.1/"), handleStream)

	
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
			stream, err := host.NewStream(ctx, peer.ID, protocol.ID("/chat/1.1/"))
			if err != nil {
				logger.Warning("Connection failed:", err)
				continue
			} else {
				rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	
				go writeData(rw)
				go readData(rw)
			}

		}
		
		select {}
	}

	return nil
}


func handleStream(stream network.Stream) {
	logger := log.Logger("p2p")
	logger.Info("Got a new stream!")
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	go readData(rw)
	go writeData(rw)
	
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if strings.TrimSpace(str) == "exit" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)


	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
		if (strings.TrimSpace(sendData) == "exit") {
			fmt.Println("reached here")
			return
		} else {
			fmt.Printf("%s is not an exit message", sendData)
		}
	}
}
