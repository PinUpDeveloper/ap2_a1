package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("[RateLimit] Redis error for IP %s: %v — allowing request", clientIP, err)
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		ttl, _ := rdb.TTL(ctx, key).Result()
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(maxRequests)-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		if count > int64(maxRequests) {
			log.Printf("[RateLimit] IP %s exceeded limit (%d/%d) — returning 429", clientIP, count, maxRequests)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"limit":       maxRequests,
				"retry_after": ttl.Seconds(),
			})
			return
		}

		c.Next()
	}
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
