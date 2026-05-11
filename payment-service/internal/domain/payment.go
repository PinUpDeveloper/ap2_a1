package domain

import "time"

const (
	PaymentStatusAuthorized = "Authorized"
	PaymentStatusDeclined   = "Declined"
	PaymentLimit            = int64(100000)
)

type Payment struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	TransactionID string    `json:"transaction_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
