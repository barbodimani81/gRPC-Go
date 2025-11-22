package grpcutils

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClientConfig struct {
	Address string
}

func Connect(cfg GRPCClientConfig) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	return conn, nil
}

func ConnectOrPanic(cfg GRPCClientConfig) *grpc.ClientConn {
	conn, err := Connect(cfg)
	if err != nil {
		panic(err.Error())
	}
	return conn
}
