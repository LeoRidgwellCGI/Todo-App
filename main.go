package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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
			// if value missing, fall through to default (will generate)
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

	// Create a base context with a TraceID (use provided when given).
	base := context.Background()
	var ctx context.Context
	var traceID string
	if providedTrace != "" {
		ctx, traceID = trace.NewWithID(base, providedTrace)
	} else {
		ctx, traceID = trace.New(base)
	}

	// Configure a global logger and stamp the TraceID onto it
	// so ALL logs include {"trace_id": "<id>"}.
	var handler slog.Handler
	if useText {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler).With(slog.String("trace_id", traceID))
	slog.SetDefault(logger)

	// Run the CLI with context so downstream logs also include the TraceID.
	if err := cli.New().Run(ctx, args); err != nil {
		slog.ErrorContext(ctx, "cli run failed", "error", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
