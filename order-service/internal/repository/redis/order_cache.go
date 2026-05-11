package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ap2/order-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redis.Client
}

func NewOrderCache(client *redis.Client) *OrderCache {
	return &OrderCache{client: client}
}

func cacheKey(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func (c *OrderCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	val, err := c.client.Get(ctx, cacheKey(id)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var order domain.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, fmt.Errorf("unmarshal order: %w", err)
	}
	return &order, nil
}

func (c *OrderCache) Set(ctx context.Context, id string, order *domain.Order, ttl time.Duration) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}
	return c.client.Set(ctx, cacheKey(id), data, ttl).Err()
}

func (c *OrderCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, cacheKey(id)).Err()
}
