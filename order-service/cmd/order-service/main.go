package main

import (
	"context"
	"log"
	"net"
	"time"

	ap2proto "ap2/contracts/gen/go/ap2proto"
	"ap2/order-service/internal/app"
	repogrpc "ap2/order-service/internal/repository/grpcclient"
	"ap2/order-service/internal/repository/postgres"
	rediscache "ap2/order-service/internal/repository/redis"
	grpctransport "ap2/order-service/internal/transport/grpc"
	httptransport "ap2/order-service/internal/transport/http"
	"ap2/order-service/internal/transport/http/middleware"
	"ap2/order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	cfg := app.LoadConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- Postgres ---
	db, err := pgxpool.New(ctx, cfg.DBDsn)
	if err != nil {
		log.Fatalf("failed to connect to order database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping order database: %v", err)
	}

	// --- Redis ---
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Printf("connected to redis at %s", cfg.RedisAddr)

	// --- Wire dependencies ---
	repository := postgres.NewOrderRepository(db)
	orderCache := rediscache.NewOrderCache(rdb)

	paymentClient, err := repogrpc.NewPaymentClient(cfg.PaymentGRPCAddr, cfg.HTTPTimeout())
	if err != nil {
		log.Fatalf("failed to connect to payment grpc server: %v", err)
	}
	defer paymentClient.Close()

	orderUsecase := usecase.NewOrderUsecase(repository, paymentClient, orderCache, cfg.CacheTTL())
	httpHandler := httptransport.NewOrderHandler(orderUsecase)
	grpcHandler := grpctransport.NewOrderServer(orderUsecase)

	go func() {
		router := gin.Default()
		router.Use(middleware.RateLimiter(rdb, cfg.RateLimitRequests, cfg.RateLimitWindow()))

		httpHandler.Register(router)
		log.Printf("order REST service listening on :%s (rate limit: %d req/%ds)",
			cfg.Port, cfg.RateLimitRequests, cfg.RateLimitWindowSeconds)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on order grpc port: %v", err)
	}

	grpcServer := grpc.NewServer()
	ap2proto.RegisterOrderServiceServer(grpcServer, grpcHandler)

	log.Printf("order gRPC service listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
