package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"ap2/notification-service/internal/provider"

	"github.com/redis/go-redis/v9"
)

const (
	idempotencyKeyPrefix = "notification:processed:"
	idempotencyTTL       = 24 * time.Hour

	maxRetries  = 5
	baseBackoff = 2 * time.Second // doubling: 2s, 4s, 8s, 16s, 32s
)

// NotificationJob holds all data about a payment event to notify on.
type NotificationJob struct {
	EventID       string
	OrderID       string
	Amount        int64
	CustomerEmail string
	Status        string
}

// NotificationWorker processes jobs asynchronously with idempotency and retry.
type NotificationWorker struct {
	sender      provider.EmailSender
	redisClient *redis.Client
}

func NewNotificationWorker(sender provider.EmailSender, redisClient *redis.Client) *NotificationWorker {
	return &NotificationWorker{sender: sender, redisClient: redisClient}
}

// Process handles a single notification job.
// It checks Redis for duplicate processing (idempotency) and retries with exponential backoff.
func (w *NotificationWorker) Process(ctx context.Context, job NotificationJob) {
	idempotencyKey := idempotencyKeyPrefix + job.EventID

	// --- Idempotency check ---
	result, err := w.redisClient.Get(ctx, idempotencyKey).Result()
	if err == nil {
		log.Printf("[Worker] SKIPPED duplicate event_id=%s (status=%s)", job.EventID, result)
		return
	}
	if err != redis.Nil {
		log.Printf("[Worker] Redis error checking idempotency for event_id=%s: %v", job.EventID, err)
		// Continue anyway — better to risk a duplicate than silently drop the notification
	}

	// --- Build email message ---
	msg := provider.EmailMessage{
		To:      job.CustomerEmail,
		Subject: fmt.Sprintf("Your order #%s has been %s", job.OrderID, job.Status),
		Body: fmt.Sprintf(
			"Hello,\n\nYour order #%s has been processed.\nStatus: %s\nAmount: $%.2f\n\nThank you!",
			job.OrderID, job.Status, float64(job.Amount)/100,
		),
	}

	// --- Retry with exponential backoff ---
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastErr = w.sender.Send(msg)
		if lastErr == nil {
			break
		}

		if attempt < maxRetries {
			backoff := baseBackoff * time.Duration(1<<(attempt-1)) // 2s, 4s, 8s, 16s, 32s
			log.Printf("[Worker] attempt %d/%d failed for event_id=%s: %v — retrying in %s",
				attempt, maxRetries, job.EventID, lastErr, backoff)
			select {
			case <-ctx.Done():
				log.Printf("[Worker] context cancelled, aborting retries for event_id=%s", job.EventID)
				return
			case <-time.After(backoff):
			}
		}
	}

	// --- Record result in Redis for idempotency ---
	if lastErr != nil {
		log.Printf("[Worker] FAILED to send notification for event_id=%s after %d attempts: %v",
			job.EventID, maxRetries, lastErr)
		// Store failure status so we don't retry indefinitely across restarts
		_ = w.redisClient.Set(ctx, idempotencyKey, "failed", idempotencyTTL).Err()
		return
	}

	log.Printf("[Worker] SUCCESS notification sent for event_id=%s order_id=%s", job.EventID, job.OrderID)
	// Store success status to prevent duplicate sends on message re-delivery
	if err := w.redisClient.Set(ctx, idempotencyKey, "sent", idempotencyTTL).Err(); err != nil {
		log.Printf("[Worker] WARNING: failed to record idempotency key for event_id=%s: %v", job.EventID, err)
	}
}
