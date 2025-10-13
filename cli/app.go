package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	// Domain / persistence package
	"todo-app/todo"
)

//
// cli/app.go (package cli)
// ------------------------
// This package owns user-facing command/flag handling. It DOES NOT do direct
// business logic or I/O; instead it coordinates with the `todo` package.
// Key behaviors:
//  - Accepts flags (-list, -add, -status, -update, -newdesc, -delete, -out).
//  - Forces all file I/O to live under ./out by normalizing -out.
//  - Uses context-aware logging and returns errors up to main().
//

// App is a thin container for CLI configuration.
type App struct{}

// New constructs an App. Useful for future dependency injection.
func New() *App { return &App{} }

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
func (a *App) Run(ctx context.Context, args []string) error {
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

	// Map the chosen output file to live under ./out/
	outPath := normalizeOutPath(*out)

	// Load existing items before applying any mutations.
	list, err := todo.Load(ctx, outPath)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load todos", "error", err, "path", outPath)
		return err
	}

	// Command routing — mutually exclusive modes for simplicity.
	switch {
	case *listOnly:
		// Just print existing items in a table.
		PrintList(list)
		return nil

	case *deleteID > 0:
		// Delete by ID then save changes.
		list, err = todo.Delete(list, *deleteID)
		if err != nil {
			slog.ErrorContext(ctx, "delete failed", "error", err, "id", *deleteID)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)

	case *updateID > 0:
		// Update only the description for simplicity.
		if strings.TrimSpace(*newDesc) == "" {
			err := errors.New("-newdesc is required when using -update")
			slog.ErrorContext(ctx, "update failed: missing -newdesc", "error", err, "id", *updateID)
			return err
		}
		list, err = todo.UpdateDescription(list, *updateID, *newDesc)
		if err != nil {
			slog.ErrorContext(ctx, "update failed", "error", err, "id", *updateID)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)

	case strings.TrimSpace(*desc) != "":
		// Add a new item with optional -status, then save.
		if _, err := todo.Add(&list, *desc, todo.Status(*status)); err != nil {
			slog.ErrorContext(ctx, "add failed", "error", err)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)

	default:
		// No mode selected; show usage and examples.
		usage()
		fmt.Println("\nExamples:")
		fmt.Println("  go run . -list")
		fmt.Println("  go run . -add \"Buy milk\" -status started")
		fmt.Println("  go run . -update 3 -newdesc \"Buy oat milk\"")
		fmt.Println("  go run . -delete 2")
		return nil
	}
}
