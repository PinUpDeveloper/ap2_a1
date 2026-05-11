package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"ap2/payment-service/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount = errors.New("amount must be greater than 0")
	ErrOrderIDEmpty  = errors.New("order_id is required")
	ErrNotFound      = errors.New("payment not found")
)

type CreatePaymentInput struct {
	OrderID       string
	Amount        int64
	CustomerEmail string
}

type PaymentUsecase struct {
	repo      domain.PaymentRepository
	publisher domain.PaymentEventPublisher
}

func NewPaymentUsecase(repo domain.PaymentRepository, publisher domain.PaymentEventPublisher) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, publisher: publisher}
}

func (u *PaymentUsecase) CreatePayment(ctx context.Context, input CreatePaymentInput) (*domain.Payment, error) {
	if strings.TrimSpace(input.OrderID) == "" {
		return nil, ErrOrderIDEmpty
	}
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	existing, err := u.repo.GetByOrderID(ctx, input.OrderID)
	if err == nil && existing != nil {
		return existing, nil
	}

	status := domain.PaymentStatusAuthorized
	if input.Amount > domain.PaymentLimit {
		status = domain.PaymentStatusDeclined
	}

	payment := &domain.Payment{
		ID:            uuid.NewString(),
		OrderID:       strings.TrimSpace(input.OrderID),
		TransactionID: uuid.NewString(),
		Amount:        input.Amount,
		Status:        status,
		CreatedAt:     time.Now().UTC(),
	}
	if err := u.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	if status == domain.PaymentStatusAuthorized && u.publisher != nil {
		email := strings.TrimSpace(input.CustomerEmail)
		if email == "" {
			email = "user@example.com"
		}
		if err := u.publisher.PublishPaymentCompleted(ctx, domain.PaymentEvent{
			EventID:       payment.ID,
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: email,
			Status:        payment.Status,
		}); err != nil {
			return nil, err
		}
	}
	return payment, nil
}

func (u *PaymentUsecase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	payment, err := u.repo.GetByOrderID(ctx, orderID)
	if err != nil || payment == nil {
		return nil, ErrNotFound
	}
	return payment, nil
}
