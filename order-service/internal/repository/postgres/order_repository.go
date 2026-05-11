package postgres

import (
	"context"
	"errors"

	"ap2/order-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, nullableString(order.IdempotencyKey), order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at, updated_at
		FROM orders WHERE id = $1
	`, id)
	var order domain.Order
	if err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.IdempotencyKey,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at, updated_at
		FROM orders WHERE idempotency_key = $1
	`, key)
	var order domain.Order
	if err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.IdempotencyKey,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("order not found")
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *OrderRepository) GetRecent(ctx context.Context, limit int) ([]domain.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, item_name, amount, status, COALESCE(idempotency_key, ''), created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.ItemName,
			&order.Amount,
			&order.Status,
			&order.IdempotencyKey,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return orders, nil
}
