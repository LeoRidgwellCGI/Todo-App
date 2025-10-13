package main

import (
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
	StatusNotStarted Status = "not started"
	StatusStarted    Status = "started"
	StatusCompleted  Status = "completed"
)

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
	ID          int
	Description string
	Status      Status
	CreatedAt   time.Time
}

// addTodo adds a new Todo to the provided slice and returns it.
func addTodo(list *[]Todo, desc string, status Status) (Todo, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return Todo{}, errors.New("description cannot be empty")
	}
	if err := status.Validate(); err != nil {
		return Todo{}, err
	}

	id := len(*list) + 1
	item := Todo{
		ID:          id,
		Description: desc,
		Status:      Status(strings.ToLower(string(status))),
		CreatedAt:   time.Now(),
	}
	*list = append(*list, item)
	return item, nil
}

// printList renders the current list of todos in a simple table.
func printList(list []Todo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}
	_ = w.Flush()
}

func usage() {
	fmt.Fprintf(os.Stderr, `to-do CLI\n\n`)
	fmt.Fprintf(os.Stderr, "Adds a single to-do item to an in-memory (empty) list and prints the list.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  go run . -add \"<description>\" [-status <not started|started|completed>]\n\n")
	flag.PrintDefaults()
}

func main() {
	var (
		desc   = flag.String("add", "", "description for the to-do item to add")
		status = flag.String("status", string(StatusNotStarted), "status for the new to-do (not started|started|completed)")
	)
	flag.Usage = usage
	flag.Parse()

	if *desc == "" {
		usage()
		os.Exit(2)
	}

	var list []Todo // empty list at start of the program run
	if _, err := addTodo(&list, *desc, Status(*status)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	printList(list)
}
