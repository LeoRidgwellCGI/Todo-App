package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// Status represents the state of a to-do item.
type Status string

const (
	// Possible statuses for a to-do item
	StatusNotStarted Status = "not started"
	StatusStarted    Status = "started"
	StatusCompleted  Status = "completed"
)

// Validate checks whether the status value is one of the allowed values.
func (s Status) Validate() error {
	switch Status(strings.ToLower(string(s))) {
	case StatusNotStarted, StatusStarted, StatusCompleted:
		return nil
	default:
		return fmt.Errorf("invalid status: %q (must be one of: %q, %q, %q)", s, StatusNotStarted, StatusStarted, StatusCompleted)
	}
}

// Todo represents a single to-do item with metadata.
type Todo struct {
	ID          int       `json:"id"`          // Unique identifier for the to-do item
	Description string    `json:"description"` // Text description of the task
	Status      Status    `json:"status"`      // Current status of the task
	CreatedAt   time.Time `json:"created_at"`  // Timestamp of when the task was created
}

// addTodo adds a new Todo to the provided slice and returns it.
// It validates the input, assigns an ID, sets the creation time, and appends the new Todo to the list.
func addTodo(list *[]Todo, desc string, status Status) (Todo, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return Todo{}, errors.New("description cannot be empty") // Ensure description is not empty
	}
	if err := status.Validate(); err != nil {
		return Todo{}, err // Validate that the status is correct
	}

	id := len(*list) + 1 // Generate a simple incremental ID based on list length
	item := Todo{
		ID:          id,
		Description: desc,
		Status:      Status(strings.ToLower(string(status))), // Normalize status to lowercase
		CreatedAt:   time.Now(),                              // Record the creation time
	}
	*list = append(*list, item) // Add new item to the list
	return item, nil
}

// printList displays the list of to-do items in a tabular format.
// It uses a tabwriter for neatly aligned columns in the terminal output.
func printList(list []Todo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED") // Header row
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}
	_ = w.Flush() // Ensure all data is written to output
}

// saveJSON writes the list of todos to a JSON file at the given path.
// It marshals the data with indentation for human-readable formatting.
func saveJSON(list []Todo, path string) error {
	// Convert the slice of Todo structs to JSON with indentation.
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err // Return any marshaling errors
	}
	// Write the formatted JSON to the specified file path.
	return os.WriteFile(path, data, 0644)
}

// loadJSON reads todos from the given JSON file path and returns them.
// If the file does not exist, an empty list and nil error are returned so the app can start fresh.
func loadJSON(path string) ([]Todo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) { // File missing is not an error for our workflow
			return []Todo{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return []Todo{}, nil
	}
	var list []Todo
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// usage prints help text describing how to use the CLI.
// Includes options for listing tasks or adding new ones.
func usage() {
	fmt.Fprintf(os.Stderr, `to-do CLI\n\n`)
	fmt.Fprintf(os.Stderr, "Loads existing to-dos from JSON, optionally adds a new item, prints the list, then saves back to JSON.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  go run . [-list] [-out todos.json]\n  go run . -add \"<description>\" [-status <not started|started|completed>] [-out todos.json]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n  -list\n    \tDisplay current list from JSON and exit (no add).\n  -out string\n    \tPath to JSON file to read from and write to (default \"todos.json\").\n")
	flag.PrintDefaults()
}

// main is the entry point of the CLI application.
// It can either display the current list of to-dos or add a new one.
// The workflow is:
//  1. Load existing tasks from disk (if any).
//  2. If -list is used, display and exit.
//  3. Otherwise, add a new task, display, and save the updated list.
func main() {
	var (
		listOnly = flag.Bool("list", false, "display current list and exit")
		desc     = flag.String("add", "", "description for the to-do item to add")
		status   = flag.String("status", string(StatusNotStarted), "status for the new to-do (not started|started|completed)")
		out      = flag.String("out", "todos.json", "path to the JSON file to read/write")
	)

	// Override default usage message
	flag.Usage = usage
	flag.Parse()

	// 1) Load existing to-dos from disk (if the file exists).
	list, err := loadJSON(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load JSON:", err)
		os.Exit(1)
	}

	// 2) If just listing, print current to-dos and exit.
	if *listOnly && *desc == "" {
		printList(list)
		return
	}

	// 3) If no description provided and not listing, show usage and exit.
	if *desc == "" {
		usage()
		os.Exit(2)
	}

	// 4) Add a new to-do item to the list.
	if _, err := addTodo(&list, *desc, Status(*status)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// 5) Print the resulting list in table format.
	printList(list)

	// 6) Save the current to-do list back to the same JSON file.
	if err := saveJSON(list, *out); err != nil {
		fmt.Fprintln(os.Stderr, "failed to save JSON:", err)
		os.Exit(1)
	}

	// 7) Confirm successful save to user.
	fmt.Println("\nSaved to:", *out)
}
