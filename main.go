package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	// Convert the slice of Todo structs to JSON.
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err // Return any marshaling errors
	}

	// Write the formatted JSON to the specified file path.
	return os.WriteFile(path, data, 0644)
}

// usage prints help text describing how to use the CLI.
// It provides details about the available flags and their purposes.
func usage() {
	fmt.Fprintf(os.Stderr, `to-do CLI\n\n`)
	fmt.Fprintf(os.Stderr, "Adds a single to-do item to an in-memory (empty) list, prints it, then saves to JSON.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  go run . -add \"<description>\" [-status <not started|started|completed>] [-out todos.json]\n\n")
	flag.PrintDefaults()
}

// main is the entry point of the CLI application.
// It handles flag parsing, item creation, printing, and JSON file saving.
func main() {
	var (
		desc   = flag.String("add", "", "description for the to-do item to add")
		status = flag.String("status", string(StatusNotStarted), "status for the new to-do (not started|started|completed)")
		out    = flag.String("out", "todos.json", "path to the JSON file to write")
	)

	// Override default usage message
	flag.Usage = usage
	flag.Parse()

	// Ensure the user provided a description; if not, show usage and exit.
	if *desc == "" {
		usage()
		os.Exit(2)
	}

	// Initialize an empty to-do list.
	var list []Todo

	// Add a new to-do item to the list.
	if _, err := addTodo(&list, *desc, Status(*status)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// Print the resulting list in table format.
	printList(list)

	// Save the current to-do list to a JSON file.
	if err := saveJSON(list, *out); err != nil {
		fmt.Fprintln(os.Stderr, "failed to save JSON:", err)
		os.Exit(1)
	}

	// Confirm successful save to user.
	fmt.Println("\nSaved to:", *out)
}
