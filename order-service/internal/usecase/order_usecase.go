package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"ap2/order-service/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount      = errors.New("amount must be greater than 0")
	ErrOrderNotFound      = errors.New("order not found")
	ErrCancelNotAllowed   = errors.New("only pending orders can be cancelled")
	ErrPaymentUnavailable = errors.New("payment service unavailable")
)

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

type OrderUsecase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
	cache         domain.OrderCache
	cacheTTL      time.Duration
}

func NewOrderUsecase(repo domain.OrderRepository, paymentClient domain.PaymentClient, cache domain.OrderCache, cacheTTL time.Duration) *OrderUsecase {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	return &OrderUsecase{repo: repo, paymentClient: paymentClient, cache: cache, cacheTTL: cacheTTL}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if strings.TrimSpace(input.IdempotencyKey) != "" {
		existing, err := u.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.NewString(),
		CustomerID:     strings.TrimSpace(input.CustomerID),
		ItemName:       strings.TrimSpace(input.ItemName),
		Amount:         input.Amount,
		Status:         domain.OrderStatusPending,
		IdempotencyKey: strings.TrimSpace(input.IdempotencyKey),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	paymentResp, err := u.paymentClient.Authorize(ctx, domain.PaymentRequest{OrderID: order.ID, Amount: order.Amount})
	if err != nil {
		_ = u.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed)
		u.invalidateCache(ctx, order.ID)
		order.Status = domain.OrderStatusFailed
		order.UpdatedAt = time.Now().UTC()
		return nil, fmt.Errorf("%w: %v", ErrPaymentUnavailable, err)
	}

	if paymentResp.Status == "Authorized" {
		order.Status = domain.OrderStatusPaid
	} else {
		order.Status = domain.OrderStatusFailed
	}

	if err := u.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}
	order.UpdatedAt = time.Now().UTC()

	if err := u.cache.Set(ctx, order.ID, order, u.cacheTTL); err != nil {
		log.Printf("[cache] failed to set order %s: %v", order.ID, err)
	}

	return order, nil
}

func (u *OrderUsecase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	// 1. Check cache
	if cached, err := u.cache.Get(ctx, id); err == nil && cached != nil {
		log.Printf("[cache] HIT for order %s", id)
		return cached, nil
	} else if err != nil {
		log.Printf("[cache] GET error for order %s: %v", id, err)
	}

	// 2. Cache miss — query database
	log.Printf("[cache] MISS for order %s — querying DB", id)
	order, err := u.repo.GetByID(ctx, id)
	if err != nil || order == nil {
		return nil, ErrOrderNotFound
	}

	// 3. Populate cache
	if err := u.cache.Set(ctx, id, order, u.cacheTTL); err != nil {
		log.Printf("[cache] failed to set order %s: %v", id, err)
	}

	return order, nil
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil || order == nil {
		return nil, ErrOrderNotFound
	}
	if !order.CanBeCancelled() {
		return nil, ErrCancelNotAllowed
	}
	if err := u.repo.UpdateStatus(ctx, id, domain.OrderStatusCancelled); err != nil {
		return nil, err
	}
	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now().UTC()

	u.invalidateCache(ctx, id)

	return order, nil
}

func (u *OrderUsecase) GetRecentOrders(ctx context.Context, limit int) ([]domain.Order, error) {
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}
	return u.repo.GetRecent(ctx, limit)
}

func (u *OrderUsecase) invalidateCache(ctx context.Context, id string) {
	if err := u.cache.Delete(ctx, id); err != nil {
		log.Printf("[cache] failed to invalidate order %s: %v", id, err)
	} else {
		log.Printf("[cache] invalidated order %s", id)
	}
}
