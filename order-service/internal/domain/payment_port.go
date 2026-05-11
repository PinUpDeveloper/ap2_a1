package domain

import "context"

type PaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type PaymentResponse struct {
	ID            string `json:"id,omitempty"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

type PaymentClient interface {
	Authorize(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
}
