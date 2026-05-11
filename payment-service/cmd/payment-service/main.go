package main

import (
	"context"
	"log"
	"net"
	"time"

	ap2proto "ap2/contracts/gen/go/ap2proto"
	"ap2/payment-service/internal/app"
	"ap2/payment-service/internal/messaging/rabbitmq"
	"ap2/payment-service/internal/repository/postgres"
	grpctransport "ap2/payment-service/internal/transport/grpc"
	httptransport "ap2/payment-service/internal/transport/http"
	"ap2/payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	cfg := app.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DBDsn)
	if err != nil {
		log.Fatalf("failed to connect to payment database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping payment database: %v", err)
	}

	repository := postgres.NewPaymentRepository(db)
	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer publisher.Close()
	paymentUsecase := usecase.NewPaymentUsecase(repository, publisher)
	httpHandler := httptransport.NewPaymentHandler(paymentUsecase)
	grpcHandler := grpctransport.NewPaymentServer(paymentUsecase)

	go func() {
		router := gin.Default()
		httpHandler.Register(router)
		log.Printf("payment REST service listening on :%s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on payment grpc port: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	ap2proto.RegisterPaymentServiceServer(grpcServer, grpcHandler)

	log.Printf("payment gRPC service listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("grpc method=%s duration=%s err=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}
