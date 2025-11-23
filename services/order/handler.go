package order

import (
	"context"
	pb "grpc/protobuf/order"
)

type Handler struct {
	pb.UnimplementedOrderServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AddOrderHandler(ctx context.Context, req *pb.AddOrderRequest) (*pb.AddOrderResponse, error) {
	validationErrs := make(map[string]string)

	if req.BookId == "" {
		validationErrs["BookID"] = "book ID must be provided"
	}
	if req.UserName == "" {
		validationErrs["userName"] = "user name can't be empty"
	}
	if req.Count == 0 {
		validationErrs["count"] = "count must be greater than zero"
	}

	order := Order{
		BookID: req.BookId,
		Count:  req.Count,
	}

	id, err := h.svc.AddOrder(ctx, order)
	if err != nil {
		validationErrs["add order"] = "can not add order"
	}

	return &pb.AddOrderResponse{OrderId: id}, nil
}
