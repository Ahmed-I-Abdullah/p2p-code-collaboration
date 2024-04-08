// Code from https://github.com/libp2p/go-libp2p/blob/master/examples/chat-with-rendezvous/flags.go
// package flags provides the command line flags to configure the peer
package flags

import (
	"flag"
	"strings"

	constansts "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/constants"
	maddr "github.com/multiformats/go-multiaddr"
)

type AddrList []maddr.Multiaddr

// String returns the string representation of the AddrList
func (al *AddrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

// Set parses the given string value into a multiaddress and adds it to the AddrList
func (al *AddrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

// StringsToAddrs converts a slice of address strings to a slice of multiaddresses
func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

// Config represents the configuration options for the chat application.
type Config struct {
	RendezvousString string   // Unique string to identify nodes
	ReposDirectory   string   // Directory to serve git repos from
	PrivateKeyFile   string   // File to store the peer's private key
	BootstrapPeers   AddrList // List of bootstrap peers' multiaddresses
	ListenAddresses  AddrList // List of addresses to listen on
	IsBootstrap      bool     // Whether the node is a bootstrap node
	GrpcPort         int      // The Grpc server port
	GitDaemonPort    int      // The port for the git daemon
}

// ParseFlags parses command-line flags and returns the configuration
func ParseFlags() (Config, error) {
	config := Config{}
	flag.StringVar(&config.RendezvousString, "rendezvous", "meet me here", "Unique string to identify group of nodes.")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.BoolVar(&config.IsBootstrap, "is_bootstrap", false, "Whether the node is a bootstrap node.")
	flag.IntVar(&config.GrpcPort, "grpcport", 0, "The Grpc server port")
	flag.IntVar(&config.GitDaemonPort, "gitport", 0, "The port for the git daemon")
	flag.StringVar(&config.ReposDirectory, "repos_dir", "./repos", "Directory to serve git repos from")
	flag.StringVar(&config.PrivateKeyFile, "priv_key", constansts.PeerPrivateKeyPath, "File to store the peer's private key")
	flag.Parse()

	return config, nil
}
