package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ap2/payment-service/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "payment.completed"

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	if _, err := ch.QueueDeclare(
		QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}

	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) PublishPaymentCompleted(ctx context.Context, event domain.PaymentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}

	confirm, err := p.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		"",        // default exchange
		QueueName, // routing key = queue name
		true,      // mandatory
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.EventID,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish payment event: %w", err)
	}
	if confirm == nil || !confirm.Wait() {
		return fmt.Errorf("payment event was not confirmed by broker")
	}
	return nil
}

func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
