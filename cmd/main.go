package main

import (
	"context"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	"github.com/ipfs/go-log/v2"
)

func main() {
	logger := log.Logger("main")
	log.SetLogLevel("main", "info")

	logger.Info("Parsing input flags")
	ctx := context.Background()

	config, err := flags.ParseFlags()
	if err != nil {
		logger.Errorf("Error parsing flags: %v", err)
		return
	}

	if config.GrpcPort == 0 && !config.IsBootstrap {
		logger.Fatalf("Please provide a Grpc server port using the grpcport flag")
		return
	}

	if config.GitDaemonPort == 0 && !config.IsBootstrap {
		logger.Fatalf("Please provide a Git daemon port using the gitport flag")
		return
	}

	peer, err := p2p.Initialize(config)
	if err != nil {
		logger.Fatalf("failed to initialize P2P: %v", err)
	}

	err = api.StartServer(ctx, peer)
	if err != nil {
		logger.Fatalf("failed to start grpc server: %v", err)
	}

	select {}
}
