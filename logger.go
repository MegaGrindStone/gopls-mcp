package main

import (
	"log/slog"
	"os"
	"strings"
)

// initLogger initializes the slog logger with configuration from environment variables.
func initLogger() *slog.Logger {
	// Get log level from environment, default to INFO
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "INFO"
	}

	var level slog.Level
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Get log format from environment, default to text
	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "text"
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
		// Add source information in debug mode
		AddSource: level == slog.LevelDebug,
	}

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
