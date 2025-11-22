package bookstore

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrBookNotFound = errors.New("book not found")
	ErrBookExists   = errors.New("book already exists")
	ErrInvalidInput = errors.New("invalid input")
)

type Book struct {
	ID       string
	Title    string
	Author   string
	ISBN     string
	Price    float64
	Quantity int32
	Category string
}

type Service struct {
	mu    sync.RWMutex
	books map[string]Book
}

func NewService() *Service {
	return &Service{
		books: make(map[string]Book),
	}
}

func (s *Service) GetBook(ctx context.Context, id string) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, exists := s.books[id]
	if !exists {
		return Book{}, ErrBookNotFound
	}
	return book, nil
}

func (s *Service) ListBooks(ctx context.Context, page, pageSize int32, category string) ([]Book, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []Book
	for _, book := range s.books {
		if category == "" || book.Category == category {
			filtered = append(filtered, book)
		}
	}

	total := len(filtered)
	start := int((page - 1) * pageSize)
	end := int(page * pageSize)

	if start >= total {
		return []Book{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (s *Service) AddBook(ctx context.Context, book Book) (string, error) {
	if book.Title == "" || book.Author == "" {
		return "", ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if ISBN already exists
	for _, existing := range s.books {
		if existing.ISBN == book.ISBN {
			return "", ErrBookExists
		}
	}

	book.ID = uuid.New().String()
	s.books[book.ID] = book

	return book.ID, nil
}

func (s *Service) UpdateBook(ctx context.Context, id string, updates Book) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book, exists := s.books[id]
	if !exists {
		return Book{}, ErrBookNotFound
	}

	// Update fields
	if updates.Title != "" {
		book.Title = updates.Title
	}
	if updates.Author != "" {
		book.Author = updates.Author
	}
	if updates.Price > 0 {
		book.Price = updates.Price
	}
	if updates.Quantity >= 0 {
		book.Quantity = updates.Quantity
	}

	s.books[id] = book
	return book, nil
}

func (s *Service) DeleteBook(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.books[id]; !exists {
		return ErrBookNotFound
	}

	delete(s.books, id)
	return nil
}
