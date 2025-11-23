package order

import (
	"context"
	"github.com/google/uuid"
	"sync"
)

type Order struct {
	ID     string
	BookID string
	Count  int32
	Status Status
}

type Status int

const (
	StatusPending Status = iota
	StatusConfirmed
	StatusShipped
)

type Service struct {
	mu     sync.RWMutex
	orders map[string]Order
}

func NewService() *Service {
	return &Service{
		orders: make(map[string]Order),
	}
}

func (s *Service) AddOrder(ctx context.Context, order Order) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order.ID = uuid.New().String()
	s.orders[order.ID] = order

	return order.ID, nil
}
