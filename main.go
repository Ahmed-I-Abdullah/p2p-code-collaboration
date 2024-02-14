package main

import (
    "github.com/ipfs/go-log/v2"
    "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/cmd/flags"
    "github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pkg/p2p"
)

func main() {    
    logger := log.Logger("main")
    log.SetLogLevel("main", "info")

    logger.Info("Parsing input flags")

    config, err := flags.ParseFlags()
    if err != nil {
        logger.Errorf("Error parsing flags: %v", err)
        return
    }

    err = p2p.Initialize(config)

    if err != nil {
        logger.Errorf("Error initializing P2P network: %v", err)
        return
    }

    select {}
}
