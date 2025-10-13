package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"todo-cli/todo"
)

// App encapsulates configuration and behavior for the CLI.
type App struct{}

// New constructs a default App.
func New() *App { return &App{} }

// usage prints help text describing how to use the CLI.
func usage() {
	fmt.Fprintf(os.Stderr, `Todo-CLI

Manage to-do items: list, add, update descriptions, or delete by ID.

Usage:
  go run . -list [-out todos.json]
  go run . -add "<description>" [-status <not started|started|completed>] [-out todos.json]
  go run . -update <id> -newdesc "<new description>" [-out todos.json]
  go run . -delete <id> [-out todos.json]
`)
}

// Run parses flags and performs the requested action.
func (a *App) Run(args []string) error {
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
		return err
	}

	// Load existing items
	list, err := todo.Load(*out)
	if err != nil {
		return fmt.Errorf("failed to load JSON: %w", err)
	}

	// LIST MODE
	if *listOnly && *desc == "" && *updateID == 0 && *deleteID == 0 {
		PrintList(list)
		return nil
	}

	// DELETE MODE
	if *deleteID > 0 {
		list, err = todo.Delete(list, *deleteID)
		if err != nil {
			return err
		}
		PrintList(list)
		return todo.Save(list, *out)
	}

	// UPDATE MODE
	if *updateID > 0 {
		if strings.TrimSpace(*newDesc) == "" {
			return errors.New("-newdesc is required when using -update")
		}
		list, err = todo.UpdateDescription(list, *updateID, *newDesc)
		if err != nil {
			return err
		}
		PrintList(list)
		return todo.Save(list, *out)
	}

	// ADD MODE
	if strings.TrimSpace(*desc) != "" {
		if _, err := todo.Add(&list, *desc, todo.Status(*status)); err != nil {
			return err
		}
		PrintList(list)
		return todo.Save(list, *out)
	}

	usage()
	fmt.Println("\nExamples:")
	fmt.Println("  go run . -list")
	fmt.Println("  go run . -add \"Buy milk\" -status started")
	fmt.Println("  go run . -update 3 -newdesc \"Buy oat milk\"")
	fmt.Println("  go run . -delete 2")
	return nil
}
