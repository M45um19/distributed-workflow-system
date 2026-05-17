package config

import (
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewAuthgRPCConnection(address string) *grpc.ClientConn {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Auth gRPC service dial failed: %v", err)
	}
	return conn
}