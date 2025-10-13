package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"todo-cli/cli"
	"todo-cli/trace"
)

// stripGlobalFlags pulls out -logtext|--logtext and -traceid/--traceid/--traceid=<val>
// so we can configure logging before passing the remaining args to the CLI.
func stripGlobalFlags(args []string) (clean []string, useText bool, traceID string) {
	clean = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]

		// Handle -logtext / --logtext (boolean, no value)
		if a == "-logtext" || a == "--logtext" {
			useText = true
			continue
		}

		// Handle --traceid=<value>
		if strings.HasPrefix(a, "--traceid=") {
			traceID = strings.TrimPrefix(a, "--traceid=")
			continue
		}

		// Handle -traceid <value> or --traceid <value>
		if a == "-traceid" || a == "--traceid" {
			if i+1 < len(args) {
				traceID = args[i+1]
				i++ // consume the value
			}
			continue
		}

		// Otherwise, keep the arg
		clean = append(clean, a)
	}
	return clean, useText, traceID
}

func main() {
	// Parse global logging/trace flags first.
	args, useText, providedTrace := stripGlobalFlags(os.Args[1:])

	// Create a base context that cancels on Ctrl+C (SIGINT).
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Add a TraceID to the context (use provided when given).
	var ctx context.Context
	var traceID string
	if providedTrace != "" {
		ctx, traceID = trace.NewWithID(sigCtx, providedTrace)
	} else {
		ctx, traceID = trace.New(sigCtx)
	}

	// Configure a global logger and stamp the TraceID onto it.
	var handler slog.Handler
	if useText {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler).With(slog.String("trace_id", traceID))
	slog.SetDefault(logger)

	// Run the CLI in a goroutine so the process can stay alive.
	errCh := make(chan error, 1)
	go func() {
		errCh <- cli.New().Run(ctx, args)
	}()

	var runErr error
	// Collect the CLI result (if any), but DO NOT exit yet.
	select {
	case runErr = <-errCh:
		if runErr != nil {
			slog.ErrorContext(ctx, "cli run failed", "error", runErr)
			// Also echo a human-friendly line:
			fmt.Fprintln(os.Stderr, runErr)
		} else {
			slog.InfoContext(ctx, "cli run completed")
		}
	default:
		// CLI still running; that's fine. We'll still only exit on Ctrl+C below.
	}

	// Inform the user and wait for Ctrl+C.
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to exit...")
	<-sigCtx.Done() // block until SIGINT

	// Graceful shutdown point.
	if runErr != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
