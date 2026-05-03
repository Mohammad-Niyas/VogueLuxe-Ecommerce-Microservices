package main

import (
	"ecommerce/wishlist-service/config"
	"ecommerce/wishlist-service/controllers"
	"ecommerce/wishlist-service/middleware"
	"ecommerce/wishlist-service/pkg/pb"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db := config.InitDB(dsn)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalln("Failed to listing:", err)
	}

	fmt.Println("Wishlist Service on :50051")

	s := controllers.Server{
		DB: db,
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor,
			middleware.JWTInterceptor,
		),
	)

	pb.RegisterWishlistServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
