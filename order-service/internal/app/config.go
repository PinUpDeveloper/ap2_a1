package app

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   string
	GRPCPort               string
	DBDsn                  string
	PaymentGRPCAddr        string
	HTTPTimeoutSeconds     int
	RedisAddr              string
	RedisPassword          string
	CacheTTLSeconds        int
	RateLimitRequests      int
	RateLimitWindowSeconds int
}

func LoadConfig() Config {
	return Config{
		Port:                   getEnv("APP_PORT", "8080"),
		GRPCPort:               getEnv("GRPC_PORT", "50052"),
		DBDsn:                  getEnv("DB_DSN", "host=localhost port=5433 user=postgres password=1234 dbname=order_db sslmode=disable"),
		PaymentGRPCAddr:        getEnv("PAYMENT_GRPC_ADDR", "localhost:50051"),
		HTTPTimeoutSeconds:     getEnvAsInt("HTTP_TIMEOUT_SECONDS", 2),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		CacheTTLSeconds:        getEnvAsInt("CACHE_TTL_SECONDS", 300),
		RateLimitRequests:      getEnvAsInt("RATE_LIMIT_REQUESTS", 10),
		RateLimitWindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}

func (c Config) HTTPTimeout() time.Duration { return time.Duration(c.HTTPTimeoutSeconds) * time.Second }
func (c Config) CacheTTL() time.Duration    { return time.Duration(c.CacheTTLSeconds) * time.Second }
func (c Config) RateLimitWindow() time.Duration {
	return time.Duration(c.RateLimitWindowSeconds) * time.Second
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvAsInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
