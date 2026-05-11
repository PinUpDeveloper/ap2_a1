package postgres

import (
	"context"

	"ap2/payment-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct{ db *pgxpool.Pool }

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository { return &PaymentRepository{db: db} }

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments (id, order_id, transaction_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, payment.ID, payment.OrderID, payment.TransactionID, payment.Amount, payment.Status, payment.CreatedAt)
	return err
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, order_id, transaction_id, amount, status, created_at
		FROM payments WHERE order_id = $1
	`, orderID)
	var payment domain.Payment
	if err := row.Scan(&payment.ID, &payment.OrderID, &payment.TransactionID, &payment.Amount, &payment.Status, &payment.CreatedAt); err != nil {
		return nil, err
	}
	return &payment, nil
}
