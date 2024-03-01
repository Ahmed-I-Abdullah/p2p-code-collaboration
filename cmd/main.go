package main

import (
	"context"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/p2p"
	daemon "github.com/aymanbagabas/go-git-daemon"
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

	// Initialize peers, DHT and p2p protocol
	peer, err := p2p.Initialize(config)
	if err != nil {
		logger.Fatalf("Failed to initialize P2P: %v", err)
	} else {
		logger.Info("Initialized Peer")
	}

	// Initialize gRPC server
	go func() {
		if err := api.StartServer(ctx, peer); err != nil {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	logger.Info("Started gRPC server")

	daemon.Enable(daemon.UploadPackService)
	daemon.Enable(daemon.UploadArchiveService)
	daemon.Enable(daemon.ReceivePackService)

	daemon.DefaultServer.BasePath = "./repos"
	daemon.DefaultServer.Verbose = true
	daemon.DefaultServer.ExportAll = true
	daemon.DefaultServer.StrictPaths = false

	// Start server on the default port :9418
	if err := daemon.ListenAndServe(":9418"); err != nil {
		logger.Fatal(err)
	}

	logger.Info("Started Git Daemon Server")

	select {}
}
