package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	// Local packages
	"todo-app/cli"
	"todo-app/trace"
)

//
// main.go (package main)
// ----------------------
// This is the executable entrypoint for Todo-App.
// Responsibilities:
//  1) Parse *global* flags that affect logging style and the TraceID before any subcommand runs.
//  2) Create a context that cancels on SIGINT (Ctrl+C).
//  3) Generate or accept an external TraceID and attach it to all logs.
//  4) Run the CLI logic (cli.App) while keeping the process alive until Ctrl+C.
//

// stripGlobalFlags extracts global flags that must be handled before the CLI flagset.
// Recognized flags:
//   - -logtext / --logtext     -> switch from JSON logs to human-readable text logs
//   - -traceid <val> / --traceid <val> / --traceid=<val>  -> provide an external trace id
//
// The function returns (remainingArgs, useTextLogger, providedTraceID).
func stripGlobalFlags(args []string) (clean []string, useText bool, traceID string) {
	clean = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]

		// Boolean logging style flag
		if a == "-logtext" || a == "--logtext" {
			useText = true
			continue
		}

		// Long form with equals: --traceid=VALUE
		if strings.HasPrefix(a, "--traceid=") {
			traceID = strings.TrimPrefix(a, "--traceid=")
			continue
		}

		// Space-separated forms: -traceid VALUE or --traceid VALUE
		if a == "-traceid" || a == "--traceid" {
			if i+1 < len(args) {
				traceID = args[i+1]
				i++ // consume value
			}
			continue
		}

		// Unhandled: keep arg
		clean = append(clean, a)
	}
	return clean, useText, traceID
}

func main() {
	// 1) Parse global flags from os.Args. We do this *before* CLI flag parsing
	// so we can configure logging (JSON vs text) and establish a TraceID early.
	args, useText, providedTrace := stripGlobalFlags(os.Args[1:])

	// 2) Create a signal-aware context that is canceled on SIGINT (Ctrl+C).
	//    This lets downstream code optionally react to cancellation when needed.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 3) Create a context that also carries a TraceID for end-to-end logging.
	//    If the user provided one via -traceid/--traceid, we use it; otherwise we generate one.
	var ctx context.Context
	var traceID string
	if providedTrace != "" {
		ctx, traceID = trace.NewWithID(sigCtx, providedTrace)
	} else {
		ctx, traceID = trace.New(sigCtx)
	}

	// Configure slog globally. We attach the trace_id so all logs include it.
	var handler slog.Handler
	if useText {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler).With(slog.String("trace_id", traceID))
	slog.SetDefault(logger)

	// 4) Run the CLI in a goroutine so we can keep the process alive until Ctrl+C.
	errCh := make(chan error, 1)
	go func() {
		errCh <- cli.New().Run(ctx, args)
	}()

	// We non-blockingly read the CLI result. Whether it fails or succeeds,
	// the process does not exit until user presses Ctrl+C.
	var runErr error
	select {
	case runErr = <-errCh:
		if runErr != nil {
			slog.ErrorContext(ctx, "cli run failed", "error", runErr)
			// Also print a human-friendly message
			fmt.Fprintln(os.Stderr, runErr)
		} else {
			slog.InfoContext(ctx, "cli run completed")
		}
	default:
		// CLI still running (fine) — we'll still wait for Ctrl+C below.
	}

	// Inform user and then wait for Ctrl+C to exit.
	fmt.Fprintln(os.Stderr, "Todo-App is running. Press Ctrl+C to exit...")
	<-sigCtx.Done()

	// Graceful exit code: 1 if CLI returned an error, else 0.
	if runErr != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
