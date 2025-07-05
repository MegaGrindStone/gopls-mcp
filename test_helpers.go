package main

import (
	"log/slog"
	"os"
)

// newTestLogger creates a logger for tests with reduced verbosity.
func newTestLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelWarn, // Only show warnings and errors in tests
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
