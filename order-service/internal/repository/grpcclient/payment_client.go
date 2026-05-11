package grpcclient

import (
	"context"
	"fmt"
	"time"

	"ap2/order-service/internal/domain"
	ap2proto "ap2/contracts/gen/go/ap2proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	conn    *grpc.ClientConn
	client  ap2proto.PaymentServiceClient
	timeout time.Duration
}

func NewPaymentClient(target string, timeout time.Duration) (*PaymentClient, error) {
	conn, err := grpc.Dial(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial payment grpc: %w", err)
	}

	return &PaymentClient{
		conn:    conn,
		client:  ap2proto.NewPaymentServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (p *PaymentClient) Close() error {
	if p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

func (p *PaymentClient) Authorize(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	resp, err := p.client.ProcessPayment(callCtx, &ap2proto.PaymentRequest{
		OrderId: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		ID:            resp.GetId(),
		OrderID:       resp.GetOrderId(),
		TransactionID: resp.GetTransactionId(),
		Amount:        resp.GetAmount(),
		Status:        resp.GetStatus(),
	}, nil
}
