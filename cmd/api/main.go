// cmd/api/main.go
// Main entry point for the Todo API server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo-app/api_app"
	"todo-app/trace"
)

// main is the entry point for the Todo API server.
func main() {
	// Logging (mirrors CLI style): JSON by default, text when LOGTEXT=1.
	var handler slog.Handler

	// Choose log handler based on environment variable.
	if os.Getenv("LOGTEXT") == "1" {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	// Create a logger with a default TraceID for main.
	logger := slog.New(handler).With(slog.String("trace_id", trace.GenerateID()))
	slog.SetDefault(logger)

	// Build server from env and run.
	s, addr := api_app.FromEnv()

	slog.Info("todo api starting", "addr", addr)
	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Run server in background.
	done := make(chan struct{})
	go func() {
		if err := s.Run(ctx, addr); err != nil {
			slog.Error("server exited with error", "error", err)
		}
		close(done)
	}()

	<-done
	time.Sleep(50 * time.Millisecond) // small drain period for logs
}
