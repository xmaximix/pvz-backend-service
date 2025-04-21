package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort      string
	GRPCPort     string
	MetricsPort  string
	DatabaseURL  string
	JWTSecret    string
	DBMaxRetries int
	DBRetryDelay time.Duration
}

func Load() Config {
	godotenv.Load()
	return Config{
		AppPort:      getenv("APP_PORT", "8080"),
		GRPCPort:     getenv("GRPC_PORT", "3000"),
		MetricsPort:  getenv("METRICS_PORT", "9000"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://postgres:pass@db:5432/pvz?sslmode=disable"),
		JWTSecret:    getenv("JWT_SECRET", "secret"),
		DBMaxRetries: atoi(getenv("DB_MAX_RETRIES", "5")),
		DBRetryDelay: parseDuration(getenv("DB_RETRY_DELAY", "2s")),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
