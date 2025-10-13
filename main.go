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

// getNextID returns the next unique ID greater than any currently in the list.
// This ensures IDs remain unique even after deletions.
func getNextID(list []Todo) int {
	max := 0
	for _, t := range list {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
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

	id := getNextID(*list) // Generate a new unique ID
	item := Todo{
		ID:          id,
		Description: desc,
		Status:      Status(strings.ToLower(string(status))), // Normalize status to lowercase
		CreatedAt:   time.Now(),                              // Record creation time
	}
	*list = append(*list, item) // Append the new item
	return item, nil
}

// printList displays all to-do items in a neatly formatted table.
func printList(list []Todo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}
	_ = w.Flush()
}

// saveJSON writes the list of todos to a JSON file.
// The output is indented for readability.
func saveJSON(list []Todo, path string) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// loadJSON reads todos from a JSON file and returns them.
// If the file does not exist, returns an empty list without error.
func loadJSON(path string) ([]Todo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Todo{}, nil // No existing file means no todos yet
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

// findIndexByID finds the index of a to-do item by its ID.
// Returns -1 if not found.
func findIndexByID(list []Todo, id int) int {
	for i, t := range list {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// updateDescription updates the description of an existing to-do item by ID.
func updateDescription(list []Todo, id int, newDesc string) ([]Todo, error) {
	newDesc = strings.TrimSpace(newDesc)
	if newDesc == "" {
		return list, errors.New("new description cannot be empty")
	}
	idx := findIndexByID(list, id)
	if idx == -1 {
		return list, fmt.Errorf("no to-do with id %d", id)
	}
	list[idx].Description = newDesc
	return list, nil
}

// deleteTodo removes a to-do item by ID.
// Returns a new slice without that item.
func deleteTodo(list []Todo, id int) ([]Todo, error) {
	idx := findIndexByID(list, id)
	if idx == -1 {
		return list, fmt.Errorf("no to-do with id %d", id)
	}
	return append(list[:idx], list[idx+1:]...), nil
}

// usage displays CLI usage instructions.
func usage() {
	fmt.Fprintf(os.Stderr, `to-do CLI\n\n`)
	fmt.Fprintf(os.Stderr, "Manage to-do items: list, add, update descriptions, or delete by ID.\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  go run . -list [-out todos.json]\n")
	fmt.Fprintf(os.Stderr, "  go run . -add \"<description>\" [-status <not started|started|completed>] [-out todos.json]\n")
	fmt.Fprintf(os.Stderr, "  go run . -update <id> -newdesc \"<new description>\" [-out todos.json]\n")
	fmt.Fprintf(os.Stderr, "  go run . -delete <id> [-out todos.json]\n\n")
	flag.PrintDefaults()
}

// main handles command-line arguments and executes the appropriate action.
// Supported actions:
//
//	-list          Display all current to-dos.
//	-add           Add a new item.
//	-update        Update a description.
//	-delete        Delete a to-do by ID.
//
// The app always saves changes back to the same JSON file.
func main() {
	// Define flags for supported operations.
	var (
		listOnly = flag.Bool("list", false, "display current list and exit")
		desc     = flag.String("add", "", "description for the to-do item to add")
		status   = flag.String("status", string(StatusNotStarted), "status for the new to-do (not started|started|completed)")
		updateID = flag.Int("update", 0, "ID of the to-do to update (description only)")
		newDesc  = flag.String("newdesc", "", "new description for the to-do when using -update")
		deleteID = flag.Int("delete", 0, "ID of the to-do to delete")
		out      = flag.String("out", "todos.json", "path to the JSON file to read/write")
	)

	flag.Usage = usage
	flag.Parse()

	// Load existing to-dos from disk (if the file exists).
	list, err := loadJSON(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load JSON:", err)
		os.Exit(1)
	}

	// LIST MODE: show all to-dos and exit.
	if *listOnly && *desc == "" && *updateID == 0 && *deleteID == 0 {
		printList(list)
		return
	}

	// DELETE MODE: remove a to-do by ID.
	if *deleteID > 0 {
		var e error
		list, e = deleteTodo(list, *deleteID)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		printList(list)
		if err := saveJSON(list, *out); err != nil {
			fmt.Fprintln(os.Stderr, "failed to save JSON:", err)
			os.Exit(1)
		}
		fmt.Println("\nDeleted ID", *deleteID, "and saved to:", *out)
		return
	}

	// UPDATE MODE: change description for an existing to-do.
	if *updateID > 0 {
		if strings.TrimSpace(*newDesc) == "" {
			fmt.Fprintln(os.Stderr, "-newdesc is required when using -update")
			os.Exit(2)
		}
		var e error
		list, e = updateDescription(list, *updateID, *newDesc)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		printList(list)
		if err := saveJSON(list, *out); err != nil {
			fmt.Fprintln(os.Stderr, "failed to save JSON:", err)
			os.Exit(1)
		}
		fmt.Println("\nUpdated ID", *updateID, "and saved to:", *out)
		return
	}

	// ADD MODE: create a new to-do.
	if strings.TrimSpace(*desc) != "" {
		if _, err := addTodo(&list, *desc, Status(*status)); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		printList(list)
		if err := saveJSON(list, *out); err != nil {
			fmt.Fprintln(os.Stderr, "failed to save JSON:", err)
			os.Exit(1)
		}
		fmt.Println("\nSaved to:", *out)
		return
	}

	// Invalid combination of flags: show help.
	usage()
	fmt.Println("\nExamples:")
	fmt.Println("  go run . -list")
	fmt.Println("  go run . -add \"Buy milk\" -status started")
	fmt.Println("  go run . -update 3 -newdesc \"Buy oat milk\"")
	fmt.Println("  go run . -delete 2")
}
