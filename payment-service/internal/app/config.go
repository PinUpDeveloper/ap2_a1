package app

import "os"

type Config struct {
	Port        string
	GRPCPort    string
	DBDsn       string
	RabbitMQURL string
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("APP_PORT", "8081"),
		GRPCPort:    getEnv("GRPC_PORT", "50051"),
		DBDsn:       getEnv("DB_DSN", "host=localhost port=5433 user=postgres password=1234 dbname=payment_db sslmode=disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
