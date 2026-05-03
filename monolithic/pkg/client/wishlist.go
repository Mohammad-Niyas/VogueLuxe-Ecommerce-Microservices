package client

import (
    "fmt"
    "log"

    "ecommerce/wishlist-service/pkg/pb" 

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var WishlistClient pb.WishlistServiceClient

func InitWishlistClient() {
    // connect to the service on port 50051
    cc, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Could not connect: %v", err)
    }

    // Initialize the client
    WishlistClient = pb.NewWishlistServiceClient(cc)
    fmt.Println("Wishlist Client connected on :50051")
}