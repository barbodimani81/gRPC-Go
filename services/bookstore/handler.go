package bookstore

import (
	"context"
	"errors"

	"grpc/pkg/grpcutils"
	pb "grpc/protobuf/bookstore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedBookstoreServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book id is required")
	}

	book, err := h.svc.GetBook(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Error(codes.Internal, "failed to get book")
	}

	return &pb.GetBookResponse{
		Book: bookToProto(book),
	}, nil
}

func (h *Handler) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	books, total, err := h.svc.ListBooks(ctx, page, pageSize, req.Category)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list books")
	}

	return &pb.ListBooksResponse{
		Books:      booksToProto(books),
		TotalCount: int32(total),
	}, nil
}

func (h *Handler) AddBook(ctx context.Context, req *pb.AddBookRequest) (*pb.AddBookResponse, error) {
	// Validation
	validationErrors := make(map[string]string)

	if req.Title == "" {
		validationErrors["title"] = "title is required"
	}
	if req.Author == "" {
		validationErrors["author"] = "author is required"
	}
	if req.Isbn == "" {
		validationErrors["isbn"] = "ISBN is required"
	}
	if req.Price <= 0 {
		validationErrors["price"] = "price must be greater than 0"
	}
	if req.Quantity < 0 {
		validationErrors["quantity"] = "quantity cannot be negative"
	}

	if len(validationErrors) > 0 {
		return nil, grpcutils.NewBadRequestError(validationErrors)
	}

	book := Book{
		Title:    req.Title,
		Author:   req.Author,
		ISBN:     req.Isbn,
		Price:    req.Price,
		Quantity: req.Quantity,
		Category: req.Category,
	}

	id, err := h.svc.AddBook(ctx, book)
	if err != nil {
		if errors.Is(err, ErrBookExists) {
			return nil, status.Error(codes.AlreadyExists, "book with this ISBN already exists")
		}
		return nil, status.Error(codes.Internal, "failed to add book")
	}

	return &pb.AddBookResponse{Id: id}, nil
}

func (h *Handler) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book id is required")
	}

	updates := Book{
		Title:    req.Title,
		Author:   req.Author,
		Price:    req.Price,
		Quantity: req.Quantity,
	}

	book, err := h.svc.UpdateBook(ctx, req.Id, updates)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Error(codes.Internal, "failed to update book")
	}

	return &pb.UpdateBookResponse{
		Book: bookToProto(book),
	}, nil
}

func (h *Handler) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book id is required")
	}

	err := h.svc.DeleteBook(ctx, req.Id)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete book")
	}

	return &pb.DeleteBookResponse{Success: true}, nil
}

// Helper functions
func bookToProto(book Book) *pb.Book {
	return &pb.Book{
		Id:       book.ID,
		Title:    book.Title,
		Author:   book.Author,
		Isbn:     book.ISBN,
		Price:    book.Price,
		Quantity: book.Quantity,
		Category: book.Category,
	}
}

func booksToProto(books []Book) []*pb.Book {
	result := make([]*pb.Book, len(books))
	for i, book := range books {
		result[i] = bookToProto(book)
	}
	return result
}
