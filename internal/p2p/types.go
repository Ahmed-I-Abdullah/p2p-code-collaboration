package p2p

type PeerInfo struct {
	GrpcPort      int `json:"grpc_port"`
	GitDaemonPort int `json:"git_daemon_port"`
}
