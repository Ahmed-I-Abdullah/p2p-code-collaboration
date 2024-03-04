// Code from https://github.com/libp2p/go-libp2p/blob/master/examples/chat-with-rendezvous/flags.go

package flags

import (
	"flag"
	"strings"

	maddr "github.com/multiformats/go-multiaddr"
)

type AddrList []maddr.Multiaddr

func (al *AddrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *AddrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

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

type Config struct {
	RendezvousString string
	ReposDirectory   string
	BootstrapPeers   AddrList
	ListenAddresses  AddrList
	IsBootstrap      bool
	GrpcPort         int
	GitDaemonPort    int
}

func ParseFlags() (Config, error) {
	config := Config{}
	flag.StringVar(&config.RendezvousString, "rendezvous", "meet me here", "Unique string to identify group of nodes.")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.BoolVar(&config.IsBootstrap, "is_bootstrap", false, "Whether the node is a bootstrap node.")
	flag.IntVar(&config.GrpcPort, "grpcport", 0, "The Grpc server port")
	flag.IntVar(&config.GitDaemonPort, "gitport", 0, "The port for the git daemon")
	flag.StringVar(&config.ReposDirectory, "repos_dir", "./repos", "Directory to serve git repos from")
	flag.Parse()

	return config, nil
}
