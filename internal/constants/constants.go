package constansts

import "github.com/libp2p/go-libp2p/core/protocol"

var PeerPrivateKeyPath = "./.priv_key"

var PeerPortsProtocol protocol.ID = "/p2p/peerinfo/1.0.0"

var DefaultReposDirectory = "./repos"

var DHTRepoPrefix protocol.ID = "/repo"
var DHTLeaderPrefix protocol.ID = "/leader"
