package review

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	pbbookstore "grpc/protobuf/bookstore"
)

var (
	ErrReviewNotFound = errors.New("review not found")
	ErrInvalidRating  = errors.New("rating must be between 1 and 5")
	ErrBookNotFound   = errors.New("book not found")
)

type Review struct {
	ID        string
	BookID    string
	UserName  string
	Rating    int32
	Comment   string
	CreatedAt time.Time
}

type Service struct {
	mu              sync.RWMutex
	reviews         map[string]Review
	reviewsByBook   map[string][]string // book_id -> review_ids
	bookstoreClient pbbookstore.BookstoreServiceClient
}

func NewService(bookstoreClient pbbookstore.BookstoreServiceClient) *Service {
	return &Service{
		reviews:         make(map[string]Review),
		reviewsByBook:   make(map[string][]string),
		bookstoreClient: bookstoreClient,
	}
}

func (s *Service) AddReview(ctx context.Context, review Review) (string, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return "", ErrInvalidRating
	}

	// Verify book exists by calling Bookstore service
	_, err := s.bookstoreClient.GetBook(ctx, &pbbookstore.GetBookRequest{
		Id: review.BookID,
	})
	if err != nil {
		return "", ErrBookNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	review.ID = uuid.New().String()
	review.CreatedAt = time.Now()

	s.reviews[review.ID] = review
	s.reviewsByBook[review.BookID] = append(s.reviewsByBook[review.BookID], review.ID)

	return review.ID, nil
}

func (s *Service) GetReviews(ctx context.Context, bookID string) ([]Review, float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reviewIDs, exists := s.reviewsByBook[bookID]
	if !exists {
		return []Review{}, 0, nil
	}

	reviews := make([]Review, 0, len(reviewIDs))
	var totalRating float64

	for _, id := range reviewIDs {
		if review, exists := s.reviews[id]; exists {
			reviews = append(reviews, review)
			totalRating += float64(review.Rating)
		}
	}

	avgRating := 0.0
	if len(reviews) > 0 {
		avgRating = totalRating / float64(len(reviews))
	}

	return reviews, avgRating, nil
}

func (s *Service) GetBookWithReviews(ctx context.Context, bookID string) (*pbbookstore.Book, []Review, float64, error) {
	// Call Bookstore service to get book details
	bookResp, err := s.bookstoreClient.GetBook(ctx, &pbbookstore.GetBookRequest{
		Id: bookID,
	})
	if err != nil {
		return nil, nil, 0, ErrBookNotFound
	}

	reviews, avgRating, err := s.GetReviews(ctx, bookID)
	if err != nil {
		return nil, nil, 0, err
	}

	return bookResp.Book, reviews, avgRating, nil
}
