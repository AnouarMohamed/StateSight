package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Common struct {
	ServiceName         string
	LogLevel            string
	DatabaseURL         string
	RedisURL            string
	GitHubWebhookSecret string
}

type API struct {
	Common
	HTTPPort          int
	ReadHeaderTimeout time.Duration
}

type Worker struct {
	Common
	PollTimeout time.Duration
}

func LoadAPI() (API, error) {
	common := loadCommon("statesight-api")
	port, err := intFromEnv("API_PORT", 8080)
	if err != nil {
		return API{}, err
	}
	return API{
		Common:            common,
		HTTPPort:          port,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func LoadWorker() (Worker, error) {
	common := loadCommon("statesight-worker")
	return Worker{
		Common:      common,
		PollTimeout: 5 * time.Second,
	}, nil
}

func loadCommon(defaultService string) Common {
	return Common{
		ServiceName:         stringFromEnv("SERVICE_NAME", defaultService),
		LogLevel:            stringFromEnv("LOG_LEVEL", "info"),
		DatabaseURL:         stringFromEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/statesight?sslmode=disable"),
		RedisURL:            stringFromEnv("REDIS_URL", "redis://localhost:6379/0"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
	}
}

func stringFromEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func intFromEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
