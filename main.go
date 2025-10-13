package main

import (
	"fmt"
	"log/slog"
	"os"

	"todo-cli/cli"
)

// stripLoggingFlags scans args for -logtext/--logtext and removes it.
// Returns cleaned args and whether text logging is requested.
func stripLoggingFlags(args []string) ([]string, bool) {
	clean := make([]string, 0, len(args))
	useText := false
	for _, a := range args {
		switch a {
		case "-logtext", "--logtext":
			useText = true
		default:
			clean = append(clean, a)
		}
	}
	return clean, useText
}

func main() {
	// Parse for global logging flag before CLI execution.
	args, useText := stripLoggingFlags(os.Args[1:])

	// Configure global structured logger.
	var handler slog.Handler
	if useText {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))

	// Run CLI app.
	if err := cli.New().Run(args); err != nil {
		slog.Error("cli run failed", "error", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
