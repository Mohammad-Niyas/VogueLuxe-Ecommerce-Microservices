package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"ecommerce/wallet-service/config"
	"ecommerce/wallet-service/controllers"
	"ecommerce/wallet-service/middleware"
	"ecommerce/wallet-service/pkg/pb"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Println("Wallet Service (gRPC) running on port " + port)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor,
			middleware.JWTInterceptor,
		),
	)

	walletServer := &controllers.WalletServer{}

	pb.RegisterWalletServiceServer(grpcServer, walletServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
