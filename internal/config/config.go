package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Common struct {
	ServiceName string
	LogLevel    string
	DatabaseURL string
	RedisURL    string
}

type API struct {
	Common
	GitHubWebhookSecret string
	AuthRequired        bool
	OIDCIssuerURL       string
	OIDCAudience        string
	OIDCAllowInsecure   bool
	HTTPPort            int
	ReadHeaderTimeout   time.Duration
}

type Worker struct {
	Common
	GitBinary               string
	GitCacheDir             string
	KubectlBinary           string
	AllowSyntheticLiveState bool
	PollTimeout             time.Duration
}

func LoadAPI() (API, error) {
	common := loadCommon("statesight-api")
	port, err := intFromEnv("API_PORT", 8080)
	if err != nil {
		return API{}, err
	}
	cfg := API{
		Common:              common,
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		AuthRequired:        boolFromEnv("AUTH_REQUIRED", false),
		OIDCIssuerURL:       strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL")),
		OIDCAudience:        strings.TrimSpace(os.Getenv("OIDC_AUDIENCE")),
		OIDCAllowInsecure:   boolFromEnv("OIDC_ALLOW_INSECURE_ISSUER", false),
		HTTPPort:            port,
		ReadHeaderTimeout:   5 * time.Second,
	}
	if err := validateAPIAuth(cfg); err != nil {
		return API{}, err
	}
	return cfg, nil
}

func LoadWorker() (Worker, error) {
	common := loadCommon("statesight-worker")
	return Worker{
		Common:                  common,
		GitBinary:               stringFromEnv("GIT_BIN", "git"),
		GitCacheDir:             stringFromEnv("GIT_CACHE_DIR", ".statesight/git-cache"),
		KubectlBinary:           stringFromEnv("KUBECTL_BIN", "kubectl"),
		AllowSyntheticLiveState: boolFromEnv("ALLOW_SYNTHETIC_LIVE_STATE", false),
		PollTimeout:             5 * time.Second,
	}, nil
}

func loadCommon(defaultService string) Common {
	return Common{
		ServiceName: stringFromEnv("SERVICE_NAME", defaultService),
		LogLevel:    stringFromEnv("LOG_LEVEL", "info"),
		DatabaseURL: stringFromEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/statesight?sslmode=disable"),
		RedisURL:    stringFromEnv("REDIS_URL", "redis://localhost:6379/0"),
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

func boolFromEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func validateAPIAuth(cfg API) error {
	if !cfg.AuthRequired {
		return nil
	}
	if cfg.OIDCIssuerURL == "" {
		return fmt.Errorf("OIDC_ISSUER_URL is required when AUTH_REQUIRED=true")
	}
	if cfg.OIDCAudience == "" {
		return fmt.Errorf("OIDC_AUDIENCE is required when AUTH_REQUIRED=true")
	}

	issuer, err := url.Parse(cfg.OIDCIssuerURL)
	if err != nil || issuer.Host == "" || (issuer.Scheme != "https" && issuer.Scheme != "http") {
		return fmt.Errorf("OIDC_ISSUER_URL must be a valid HTTP(S) issuer URL")
	}
	if issuer.Scheme != "https" && !cfg.OIDCAllowInsecure {
		return fmt.Errorf("OIDC_ISSUER_URL must use HTTPS unless OIDC_ALLOW_INSECURE_ISSUER=true")
	}
	return nil
}
