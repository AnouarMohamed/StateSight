package observability

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger returns a JSON logger configured for service workloads.
func NewLogger(level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
