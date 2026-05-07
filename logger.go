package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/go-logr/logr"
)

// NewLogger creates a new logr.Logger with slog backend
func NewLogger(level string) logr.Logger {
	slogLevel := parseSlogLevel(level)
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewTextHandler(os.Stderr, opts)
	// handler := slog.NewJSONHandler(os.Stderr, opts)

	return logr.FromSlogHandler(handler)
}

// NewNopLogger creates a no-op logger for testing
func NewNopLogger() logr.Logger {
	return logr.Discard()
}

func parseSlogLevel(level string) slog.Level {
	//nolint:revive // identical-switch-branches: default fallback improves readability
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
