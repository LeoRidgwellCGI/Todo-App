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

	"todo-cli/todo"
)

// App encapsulates CLI configuration and behavior.
type App struct{}

// New constructs a default App instance.
func New() *App { return &App{} }

// usage prints help text and includes info about global logging/trace flags.
func usage() {
	fmt.Fprintf(os.Stderr, `Todo-CLI

Manage to-do items: list, add, update descriptions, or delete by ID.

Usage:
  go run . -list [-out out/todos.json]
  go run . -add "<description>" [-status <not started|started|completed>] [-out out/todos.json]
  go run . -update <id> -newdesc "<new description>" [-out out/todos.json]
  go run . -delete <id> [-out out/todos.json]

Notes:
  * All output is written under ./out/. If you pass a different -out value,
    it will be normalized to ./out/<basename>.
  * The application stays running and exits only on Ctrl+C (SIGINT).

Global flags (parsed before others in main):
  -logtext              Use plain text logs instead of JSON
  -traceid <value>      Provide an external TraceID (overrides auto-generated)
  --traceid=<value>     Alternate form
`)
}

// normalizeOutPath forces paths to live under ./out by joining the basename
// with the "out" directory, unless the first path segment is already "out".
func normalizeOutPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		p = "todos.json"
	}
	clean := filepath.Clean(p)
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

// Run parses CLI flags and executes the requested command.
func (a *App) Run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("todo", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	listOnly := fs.Bool("list", false, "display current list and exit")
	desc := fs.String("add", "", "description for the to-do item to add")
	status := fs.String("status", string(todo.StatusNotStarted), "status for the new to-do (not started|started|completed)")
	updateID := fs.Int("update", 0, "ID of the to-do to update (description only)")
	newDesc := fs.String("newdesc", "", "new description for the to-do when using -update")
	out := fs.String("out", "out/todos.json", "path to the JSON file to read/write (forced under ./out)")
	deleteID := fs.Int("delete", 0, "ID of the to-do to delete")

	fs.Usage = usage

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		slog.ErrorContext(ctx, "flag parsing failed", "error", err)
		return err
	}

	outPath := normalizeOutPath(*out)

	// Load existing to-dos from file.
	list, err := todo.Load(ctx, outPath)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load todos", "error", err, "path", outPath)
		return err
	}

	// Handle commands.
	switch {
	case *listOnly:
		PrintList(list)
		return nil

	case *deleteID > 0:
		list, err = todo.Delete(list, *deleteID)
		if err != nil {
			slog.ErrorContext(ctx, "delete failed", "error", err, "id", *deleteID)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, outPath)

	case *updateID > 0:
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
		if _, err := todo.Add(&list, *desc, todo.Status(*status)); err != nil {
			slog.ErrorContext(ctx, "add failed", "error", err)
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
