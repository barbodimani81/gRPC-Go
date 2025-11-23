package main

import (
	"context"
	"grpc/pkg/grpcutils"
	"grpc/pkg/middleware"
	pb "grpc/protobuf/order"
	"grpc/services/order"
	"log"
)

func main() {
	log.Println("Starting Order gRPC server...")

	// Connect to Bookstore service
	bookstoreConn := grpcutils.ConnectOrPanic(grpcutils.GRPCClientConfig{
		Address: "localhost:9090",
	})
	defer bookstoreConn.Close()

	svc := order.NewService()
	hndlr := order.NewHandler(svc)

	server := grpcutils.NewServer("Order", false)

	server.UseUnaryInterceptors(
		middleware.RequestID,
		middleware.Logger,
	)

	server.RegisterService = func(ctx context.Context, grpcSrv *grpcutils.Server, closer *grpcutils.Closer) error {
		pb.RegisterOrderServiceServer(grpcSrv.ServerGRPC, hndlr)
		log.Println("BookstoreService registered")
		return nil
	}

	server.Run(context.Background(), ":9093")
}
