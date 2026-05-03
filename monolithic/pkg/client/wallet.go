package client

import (
	"fmt"
	"log"

	wallet "ecommerce/pkg/pb/wallet"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var WalletClient wallet.WalletServiceClient

func InitWalletClient() {
	// connect to the service on port 50051 (or specific wallet port if different)
	// Wallet Service running on 50051 according to previous main.go
	cc, err := grpc.NewClient("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	// Initialize the client
	WalletClient = wallet.NewWalletServiceClient(cc)
	fmt.Println("Wallet Client connected on :8082")
}
