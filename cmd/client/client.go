// This client is for testing
// It acts as the CLI client that we will add later on
package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/Ahmed-I-Abdullah/p2p-code-collaboration/internal/api/pb"
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

	logger.Printf("Initializing repository with name: %s", defaultName)
	r, err := c.Init(ctx, &pb.RepoInitRequest{Name: defaultName, FromCli: true})
	if err != nil {
		logger.Fatalf("Failed to initialize repository: %v", err)
	}
	logger.Printf("Repository initialized successfully. Server Response: %v", r.Message)
}
