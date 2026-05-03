package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	log.Printf("Method: %s", info.FullMethod)

	resp, err := handler(ctx, req)

	log.Printf("Method: %s | Duration: %v | Error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}
