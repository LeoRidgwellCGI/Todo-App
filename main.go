package main

import (
	"encoding/csv"
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

// Todo represents a single to-do item.
type Todo struct {
	ID          int       // Unique identifier for the to-do item
	Description string    // Text description of the task
	Status      Status    // Current status of the task
	CreatedAt   time.Time // Timestamp of when the task was created
}

// addTodo adds a new Todo to the provided slice and returns it.
// It validates the input, assigns an ID, and appends the new Todo to the list.
func addTodo(list *[]Todo, desc string, status Status) (Todo, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return Todo{}, errors.New("description cannot be empty")
	}
	if err := status.Validate(); err != nil {
		return Todo{}, err
	}

	id := len(*list) + 1 // Generate a simple incremental ID
	item := Todo{
		ID:          id,
		Description: desc,
		Status:      Status(strings.ToLower(string(status))),
		CreatedAt:   time.Now(),
	}
	*list = append(*list, item)
	return item, nil
}

// printList displays the list of to-do items in a tabular format.
// It uses a tabwriter for better column alignment.
func printList(list []Todo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}
	_ = w.Flush()
}

// saveCSV writes the list of todos to a CSV file at the given path.
// The file will contain a header row followed by one row per to-do item.
func saveCSV(list []Todo, path string) error {
	// Create (or truncate) the target file. Ensure the directory exists.
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"id", "description", "status", "created"}); err != nil {
		return err
	}
	// Rows
	for _, t := range list {
		row := []string{
			fmt.Sprintf("%d", t.ID),
			t.Description,
			string(t.Status),
			t.CreatedAt.Format(time.RFC3339),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

// usage prints help text describing how to use the CLI.
func usage() {
	fmt.Fprintf(os.Stderr, `To-do CLI\n\n`)
	fmt.Fprintf(os.Stderr, "Adds a single to-do item to an in-memory (empty) list, prints it, then saves to CSV.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  go run . -add \"<description>\" [-status <not started|started|completed>] [-out todos.csv]\n\n")
	flag.PrintDefaults()
}

// main is the entry point of the CLI application.
// It parses command-line flags, adds a to-do item, prints the list, and saves it to CSV.
func main() {
	var (
		desc   = flag.String("add", "", "description for the to-do item to add")
		status = flag.String("status", string(StatusNotStarted), "status for the new to-do (not started|started|completed)")
		out    = flag.String("out", "todos.csv", "path to the CSV file to write")
	)
	flag.Usage = usage
	flag.Parse()

	// Ensure the user provided a description
	if *desc == "" {
		usage()
		os.Exit(2)
	}

	// Initialize an empty list for the session
	var list []Todo

	// Add the new to-do item
	if _, err := addTodo(&list, *desc, Status(*status)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// Print the resulting list
	printList(list)

	// Save the list to CSV
	if err := saveCSV(list, *out); err != nil {
		fmt.Fprintln(os.Stderr, "failed to save CSV:", err)
		os.Exit(1)
	}

	fmt.Println("\nSaved to:", *out)
}
