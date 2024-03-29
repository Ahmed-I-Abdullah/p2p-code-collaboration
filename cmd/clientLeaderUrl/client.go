// This client is for testing
// It acts as the CLI client that we will add later on
package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/pb"
)

const (
	// change this hehe
	address     = "127.0.0.1:3000"
	defaultName = "Amino"
)

func main() {
	logger := log.New(os.Stdout, "client: ", log.LstdFlags|log.Lshortfile)

	logger.Printf("Connecting to gRPC server at %s...", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()
	logger.Println("Connected to gRPC server.")

	c := pb.NewRepositoryClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	logger.Printf("Getting git address for repo: %s", defaultName)
	r, err := c.GetLeaderUrl(ctx, &pb.LeaderUrlRequest{Name: defaultName})
	if err != nil {
		logger.Fatalf("Failed to get leader address: %v", err)
	}
	logger.Printf("Repository address retrieved successfully. Server Response: %v", r)
}
