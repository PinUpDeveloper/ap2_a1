package grpc

import (
	"context"

	ap2proto "ap2/contracts/gen/go/ap2proto"
	"ap2/payment-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	ap2proto.UnimplementedPaymentServiceServer
	usecase *usecase.PaymentUsecase
}

func NewPaymentServer(usecase *usecase.PaymentUsecase) *PaymentServer {
	return &PaymentServer{usecase: usecase}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *ap2proto.PaymentRequest) (*ap2proto.PaymentResponse, error) {
	payment, err := s.usecase.CreatePayment(ctx, usecase.CreatePaymentInput{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	})
	if err != nil {
		switch err {
		case usecase.ErrOrderIDEmpty, usecase.ErrInvalidAmount:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &ap2proto.PaymentResponse{
		Id:            payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     timestamppb.New(payment.CreatedAt),
	}, nil
}
