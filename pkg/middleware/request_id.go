package middleware

import (
	"context"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func RequestID(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	requestID := uuid.New().String()

	log.Printf("Request ID: %s", requestID)

	ctx = context.WithValue(ctx, "request_id", requestID)

	md := metadata.Pairs("x-request-id", requestID)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return handler(ctx, req)
}
