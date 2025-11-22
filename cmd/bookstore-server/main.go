package main

import (
	"context"
	"log"

	"grpc/pkg/grpcutils"
	"grpc/pkg/middleware"
	pb "grpc/protobuf/bookstore"
	"grpc/services/bookstore"
)

func main() {
	log.Println("Starting Bookstore gRPC server...")

	svc := bookstore.NewService()
	handler := bookstore.NewHandler(svc)

	srv := grpcutils.NewServer("Bookstore", false)

	// Add interceptors
	srv.UseUnaryInterceptors(
		middleware.RequestID,
		middleware.Logger,
	)

	srv.RegisterService = func(ctx context.Context, grpcSrv *grpcutils.Server, closer *grpcutils.Closer) error {
		pb.RegisterBookstoreServiceServer(grpcSrv.ServerGRPC, handler)
		log.Println("BookstoreService registered")
		return nil
	}

	srv.Run(context.Background(), ":9090")
}
