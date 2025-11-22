package main

import (
	"context"
	"log"

	"grpc/pkg/grpcutils"
	"grpc/pkg/middleware"
	pb_bookstore "grpc/protobuf/bookstore"
	pb_review "grpc/protobuf/review"
	"grpc/services/review"
)

func main() {
	log.Println("Starting Review gRPC server...")

	// Connect to Bookstore service
	bookstoreConn := grpcutils.ConnectOrPanic(grpcutils.GRPCClientConfig{
		Address: "localhost:9090",
	})
	defer bookstoreConn.Close()

	bookstoreClient := pb_bookstore.NewBookstoreServiceClient(bookstoreConn)
	log.Println("Connected to Bookstore service")

	// Create review service
	svc := review.NewService(bookstoreClient)
	handler := review.NewHandler(svc)

	// Create gRPC server
	srv := grpcutils.NewServer("Review", false)

	srv.UseUnaryInterceptors(
		middleware.RequestID,
		middleware.Logger,
	)

	srv.RegisterService = func(ctx context.Context, grpcSrv *grpcutils.Server, closer *grpcutils.Closer) error {
		pb_review.RegisterReviewServiceServer(grpcSrv.ServerGRPC, handler)
		log.Println("ReviewService registered")
		return nil
	}

	srv.Run(context.Background(), ":9091")
}
