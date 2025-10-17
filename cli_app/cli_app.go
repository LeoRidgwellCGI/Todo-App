package cli_app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	// Domain / persistence package
	"todo-app/todo"
)

//
// app/cli_app.go (package cli_app)
// ------------------------
// This package owns user-facing command/flag handling. It DOES NOT do direct
// business logic or I/O; instead it coordinates with the `todo` package.
// Key behaviors:
//  - Accepts flags (-list, -add, -status, -update, -newdesc, -delete, -out).
//  - Forces all file I/O to live under ./out by normalizing -out.
//  - Uses context-aware logging and returns errors up to main().
//

// App is a thin container for CLI configuration.
type CLI_App struct{}

// New constructs an App. Useful for future dependency injection.
func New() *CLI_App { return &CLI_App{} }

// usage prints human-readable help and includes documentation for global flags.
func usage() {
	fmt.Fprintf(os.Stderr, `Todo-App

Manage to-do items: list, add, update descriptions, or delete by ID.

Usage:
  go run . -list [-out out/todos.json]
  go run . -add "<description>" [-status <not started|started|completed>] [-out out/todos.json]
  go run . -update <id> -newdesc "<new description>" [-out out/todos.json]
  go run . -delete <id> [-out out/todos.json]

Notes:
  * All output is written under ./out/.
    If you pass a different -out value, it will be normalized to ./out/<basename>.
  * The process exits only on Ctrl+C (SIGINT).

Global flags (parsed before others in main):
  -logtext              Use plain text logs instead of JSON
  -traceid <value>      Provide an external TraceID (overrides auto-generated)
  --traceid=<value>     Alternate form
`)
}

// PrintList prints a simple fixed table to stdout.
// We rely on tabwriter to align columns regardless of content width.
// NOTE: stdout is for user-facing output; logs go to stderr via slog.
func PrintList(list []todo.Item) {
	// Create a writer that aligns columns based on tab stops.
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	// Header line (columns are separated by tabs; tabwriter turns tabs into padding).
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")

	// Body rows
	for _, t := range list {
		// Time is formatted as RFC3339 for easy machine readability and consistency.
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}

	// Flush to ensure content is rendered even if buffers are not full.
	_ = w.Flush()
}

// normalizeOutPath ensures the data file path is always under ./out/.
// If user provides something like "/tmp/foo.json" or "something/bar.json",
// we rewrite it to "out/<basename>" to keep all outputs local to the repo.
func normalizeOutPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		p = "todos.json"
	}
	clean := filepath.Clean(p)
	// Convert to slash so splitting is consistent on all OSes.
	rel := strings.TrimLeft(filepath.ToSlash(clean), "/")
	firstSeg := rel
	if idx := strings.IndexByte(rel, '/'); idx >= 0 {
		firstSeg = rel[:idx]
	}
	if firstSeg == "out" {
		return clean
	}
	return filepath.Join("out", filepath.Base(clean))
}

// Run executes the CLI command flow using the provided context and args.
// Returns an error for any failure (parsing, I/O, validation), which main() logs.
func (a *CLI_App) Run(ctx context.Context, args []string) error {
	// Define the CLI flagset
	fs := flag.NewFlagSet("todo-app", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	listOnly := fs.Bool("list", false, "display current list and exit")
	desc := fs.String("add", "", "description for the to-do item to add")
	status := fs.String("status", string(todo.StatusNotStarted), "status for the new to-do (not started|started|completed)")
	updateID := fs.Int("update", 0, "ID of the to-do to update (description only)")
	newDesc := fs.String("newdesc", "", "new description for the to-do when using -update")
	out := fs.String("out", "out/todos.json", "path to the JSON file to read/write (forced under ./out)")
	deleteID := fs.Int("delete", 0, "ID of the to-do to delete")

	// Override default usage printer
	fs.Usage = usage

	// Parse args
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		slog.ErrorContext(ctx, "flag parsing failed", "error", err)
		return err
	}

	// Safely read flag values once to avoid repeated pointer dereferencing.
	listMode := false
	if listOnly != nil {
		listMode = *listOnly
	}
	descVal := ""
	if desc != nil {
		descVal = strings.TrimSpace(*desc)
	}
	statusVal := todo.StatusNotStarted
	if status != nil {
		statusVal = todo.Status(*status)
	}
	updateIDVal := 0
	if updateID != nil {
		updateIDVal = *updateID
	}
	newDescVal := ""
	if newDesc != nil {
		newDescVal = strings.TrimSpace(*newDesc)
	}
	deleteIDVal := 0
	if deleteID != nil {
		deleteIDVal = *deleteID
	}
	outVal := "out/todos.json"
	if out != nil {
		outVal = *out
	}

	// Map the chosen output file to live under ./out/
	outPath := normalizeOutPath(outVal)

	// Load existing items before applying any mutations.
	list, err := todo.Load(ctx, outPath)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load todos", "error", err, "path", outPath)
		return err
	}

	// Command routing

	// Command routing — mutually exclusive modes for simplicity.
	switch {
	case listMode:
		PrintList(list)
		return nil
	case descVal != "":
		var it todo.Item
		var err error
		list, it, err = todo.Add(list, descVal, statusVal)
		if err != nil {
			slog.ErrorContext(ctx, "add failed", "error", err)
			return err
		}
		_ = it
		PrintList(list)
		return todo.Save(ctx, list, outPath)
	case updateIDVal > 0 && newDescVal != "":
		var err error
		list, err = todo.UpdateDescription(list, updateIDVal, newDescVal)
		if err != nil {
			slog.ErrorContext(ctx, "update failed", "error", err)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)
	case deleteIDVal > 0:
		var err error
		list, err = todo.Delete(list, deleteIDVal)
		if err != nil {
			slog.ErrorContext(ctx, "delete failed", "error", err)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)
	default:
		usage()
		fmt.Println("\nExamples:")
		fmt.Println("  go run . -list")
		fmt.Println("  go run . -add \"Buy milk\" -status started")
		fmt.Println("  go run . -update 3 -newdesc \"Buy oat milk\"")
		fmt.Println("  go run . -delete 2")
		return nil
	}

}
