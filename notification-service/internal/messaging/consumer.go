package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ap2/notification-service/internal/worker"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type Consumer struct {
	conn               *amqp.Connection
	ch                 *amqp.Channel
	queueName          string
	notificationWorker *worker.NotificationWorker
}

func NewConsumer(url, queueName string, w *worker.NotificationWorker) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	if err := ch.Qos(1, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}
	return &Consumer{conn: conn, ch: ch, queueName: queueName, notificationWorker: w}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	deliveries, err := c.ch.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	log.Printf("notification service is listening queue=%s", c.queueName)
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-deliveries:
			if !ok {
				return nil
			}
			go c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] invalid message body: %v", err)
		_ = msg.Nack(false, false)
		return
	}

	_ = msg.Ack(false)

	c.notificationWorker.Process(ctx, worker.NotificationJob{
		EventID:       event.EventID,
		OrderID:       event.OrderID,
		Amount:        event.Amount,
		CustomerEmail: event.CustomerEmail,
		Status:        event.Status,
	})
}

func (c *Consumer) Close() error {
	if c == nil {
		return nil
	}
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
