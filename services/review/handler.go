package review

import (
	"context"
	"errors"

	"grpc/pkg/grpcutils"
	pb "grpc/protobuf/review"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedReviewServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AddReview(ctx context.Context, req *pb.AddReviewRequest) (*pb.AddReviewResponse, error) {
	validationErrors := make(map[string]string)

	if req.BookId == "" {
		validationErrors["book_id"] = "book ID is required"
	}
	if req.UserName == "" {
		validationErrors["user_name"] = "user name is required"
	}
	if req.Rating < 1 || req.Rating > 5 {
		validationErrors["rating"] = "rating must be between 1 and 5"
	}

	if len(validationErrors) > 0 {
		return nil, grpcutils.NewBadRequestError(validationErrors)
	}

	review := Review{
		BookID:   req.BookId,
		UserName: req.UserName,
		Rating:   req.Rating,
		Comment:  req.Comment,
	}

	id, err := h.svc.AddReview(ctx, review)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Error(codes.Internal, "failed to add review")
	}

	return &pb.AddReviewResponse{ReviewId: id}, nil
}

func (h *Handler) GetReviews(ctx context.Context, req *pb.GetReviewsRequest) (*pb.GetReviewsResponse, error) {
	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book ID is required")
	}

	reviews, avgRating, err := h.svc.GetReviews(ctx, req.BookId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get reviews")
	}

	return &pb.GetReviewsResponse{
		Reviews:       reviewsToProto(reviews),
		AverageRating: avgRating,
	}, nil
}

func (h *Handler) GetBookWithReviews(ctx context.Context, req *pb.GetBookWithReviewsRequest) (*pb.GetBookWithReviewsResponse, error) {
	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book ID is required")
	}

	book, reviews, avgRating, err := h.svc.GetBookWithReviews(ctx, req.BookId)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Error(codes.Internal, "failed to get book with reviews")
	}

	return &pb.GetBookWithReviewsResponse{
		Book: &pb.BookInfo{
			Id:     book.Id,
			Title:  book.Title,
			Author: book.Author,
			Price:  book.Price,
		},
		Reviews:       reviewsToProto(reviews),
		AverageRating: avgRating,
	}, nil
}

func reviewToProto(review Review) *pb.Review {
	return &pb.Review{
		Id:        review.ID,
		BookId:    review.BookID,
		UserName:  review.UserName,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt.Unix(),
	}
}

func reviewsToProto(reviews []Review) []*pb.Review {
	result := make([]*pb.Review, len(reviews))
	for i, review := range reviews {
		result[i] = reviewToProto(review)
	}
	return result
}
