package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/flags"
	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/gitops"
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

	git := gitops.New(config.ReposDirectory)

	// Initialize gRPC server
	go func() {
		if err := api.StartServer(ctx, peer, git); err != nil {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
	logger.Info("Started gRPC server")

	// Create directory where repos will be stored
	if _, err := os.Stat(config.ReposDirectory); os.IsNotExist(err) {
		err := os.Mkdir(config.ReposDirectory, 0755)
		if err != nil {
			logger.Fatalf("Failed to create repos directory: %v", err)
		}
	}

	// Enable all Git Daemon services
	daemon.Enable(daemon.UploadPackService)
	daemon.Enable(daemon.UploadArchiveService)
	daemon.Enable(daemon.ReceivePackService)

	daemon.DefaultServer.BasePath = config.ReposDirectory
	daemon.DefaultServer.Verbose = true
	daemon.DefaultServer.ExportAll = true
	daemon.DefaultServer.StrictPaths = false

	daemonAddress := fmt.Sprintf(":%d", config.GitDaemonPort)
	if err := daemon.ListenAndServe(daemonAddress); err != nil {
		logger.Fatalf("Failed to start git daemon: %v", err)
	}

	logger.Info("Started Git Daemon Server")

	select {}
}
