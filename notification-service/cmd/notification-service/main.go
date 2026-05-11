package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"ap2/notification-service/internal/app"
	"ap2/notification-service/internal/messaging"
	"ap2/notification-service/internal/provider"
	"ap2/notification-service/internal/worker"

	goredis "github.com/redis/go-redis/v9"
)

func main() {
	cfg := app.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Redis ---
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("notification service connected to redis at", cfg.RedisAddr)

	// --- Email provider: choose via PROVIDER_MODE env var ---
	var emailSender provider.EmailSender
	switch cfg.ProviderMode {
	case "REAL":
		log.Println("using REAL SMTP email provider")
		emailSender = provider.NewSMTPEmailSender(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom,
		)
	default:
		log.Println("using SIMULATED email provider (set PROVIDER_MODE=REAL for real SMTP)")
		emailSender = provider.NewSimulatedEmailSender()
	}

	// --- Worker ---
	notifWorker := worker.NewNotificationWorker(emailSender, rdb)

	// --- RabbitMQ consumer ---
	consumer, err := messaging.NewConsumer(cfg.RabbitMQURL, cfg.QueueName, notifWorker)
	if err != nil {
		log.Fatalf("failed to start notification consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("notification consumer stopped with error: %v", err)
	}
	log.Println("notification service stopped gracefully")
}
