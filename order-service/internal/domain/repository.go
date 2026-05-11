package domain

import "context"

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetRecent(ctx context.Context, limit int) ([]Order, error)
}
