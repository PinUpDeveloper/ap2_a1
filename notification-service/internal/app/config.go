package app

import "os"

type Config struct {
	RabbitMQURL   string
	QueueName     string
	RedisAddr     string
	RedisPassword string
	ProviderMode  string // REAL or SIMULATED
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
}

func LoadConfig() Config {
	return Config{
		RabbitMQURL:   getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:     getEnv("QUEUE_NAME", "payment.completed"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		ProviderMode:  getEnv("PROVIDER_MODE", "SIMULATED"),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:      getEnv("SMTP_FROM", "noreply@example.com"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
