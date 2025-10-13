package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
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
  go run . -list [-out todos.json]
  go run . -add "<description>" [-status <not started|started|completed>] [-out todos.json]
  go run . -update <id> -newdesc "<new description>" [-out todos.json]
  go run . -delete <id> [-out todos.json]

Global flags (parsed before others in main):
  -logtext              Use plain text logs instead of JSON
  -traceid <value>      Provide an external TraceID (overrides auto-generated)
  --traceid=<value>     Alternate form
`)
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
	deleteID := fs.Int("delete", 0, "ID of the to-do to delete")
	out := fs.String("out", "todos.json", "path to the JSON file to read/write")

	fs.Usage = usage

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		slog.ErrorContext(ctx, "flag parsing failed", "error", err)
		return err
	}

	// Load existing to-dos from file.
	list, err := todo.Load(ctx, *out)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load todos", "error", err, "path", *out)
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
		return todo.Save(ctx, list, *out)

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
		return todo.Save(ctx, list, *out)

	case strings.TrimSpace(*desc) != "":
		if _, err := todo.Add(&list, *desc, todo.Status(*status)); err != nil {
			slog.ErrorContext(ctx, "add failed", "error", err)
			return err
		}
		PrintList(list)
		return todo.Save(ctx, list, *out)

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
