package grpc

import (
	"errors"
	"time"

	"ap2/order-service/internal/usecase"
	ap2proto "ap2/contracts/gen/go/ap2proto"

	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderServer struct {
	ap2proto.UnimplementedOrderServiceServer
	usecase *usecase.OrderUsecase
}

func NewOrderServer(usecase *usecase.OrderUsecase) *OrderServer {
	return &OrderServer{usecase: usecase}
}

func (s *OrderServer) SubscribeToOrderUpdates(req *ap2proto.OrderRequest, stream ap2proto.OrderService_SubscribeToOrderUpdatesServer) error {
	if req.GetOrderId() == "" {
		return grpcstatus.Error(codes.InvalidArgument, "order_id is required")
	}

	ctx := stream.Context()
	order, err := s.usecase.GetOrder(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			return grpcstatus.Error(codes.NotFound, err.Error())
		}
		return grpcstatus.Error(codes.Internal, err.Error())
	}

	lastStatus := order.Status
	if err := stream.Send(&ap2proto.OrderStatusUpdate{
		OrderId:   order.ID,
		Status:    order.Status,
		UpdatedAt: timestamppb.New(time.Now().UTC()),
	}); err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			current, err := s.usecase.GetOrder(ctx, req.GetOrderId())
			if err != nil {
				return grpcstatus.Error(codes.Internal, err.Error())
			}
			if current.Status != lastStatus {
				lastStatus = current.Status
				if err := stream.Send(&ap2proto.OrderStatusUpdate{
					OrderId:   current.ID,
					Status:    current.Status,
					UpdatedAt: timestamppb.New(time.Now().UTC()),
				}); err != nil {
					return err
				}
			}
		}
	}
}
