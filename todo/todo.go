package todo

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

//
// todo/todo.go (package todo)
// ---------------------------
// Domain model and core operations for to-do items.
// No I/O or logging here — just state changes and validation.
//

// Status represents the state of a to-do item.
type Status string

const (
	StatusNotStarted Status = "not started"
	StatusStarted    Status = "started"
	StatusCompleted  Status = "completed"
)

// Validate ensures the status is one of the allowed values (case-insensitive).
func (s Status) Validate() error {
	switch Status(strings.ToLower(string(s))) {
	case StatusNotStarted, StatusStarted, StatusCompleted:
		return nil
	default:
		return fmt.Errorf("invalid status: %q (allowed: %q, %q, %q)", s, StatusNotStarted, StatusStarted, StatusCompleted)
	}
}

// Item is the domain entity persisted in JSON.
// ID is a simple integer; CreatedAt is stored as RFC3339 in the JSON.
type Item struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// getNextID returns the next max(ID)+1 for the given list.
// This is sufficient for a single-process CLI demo.
func getNextID(list []Item) int {
	max := 0
	for _, t := range list {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}

// Add validates input and appends a new item to the list.
// Returns the created item or an error on invalid input.

// Add creates a new item and returns the updated slice plus the created item.
// It follows the mutation pattern used by the other functions (take a slice, return a slice).
func Add(list []Item, desc string, status Status) ([]Item, Item, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return list, Item{}, errors.New("description cannot be empty")
	}
	if err := status.Validate(); err != nil {
		return list, Item{}, err
	}
	item := Item{
		ID:          getNextID(list),
		Description: desc,
		Status:      Status(strings.ToLower(string(status))),
		CreatedAt:   time.Now(),
	}
	list = append(list, item)
	return list, item, nil
}

// UpdateStatus finds an item by id and updates its Status.
// Returns a new slice (copy-on-write style) to make the mutation explicit.
func UpdateStatus(list []Item, id int, s Status) ([]Item, error) {
	if err := s.Validate(); err != nil {
		return list, err
	}
	for i := range list {
		if list[i].ID == id {
			list[i].Status = Status(strings.ToLower(string(s)))
			return list, nil
		}
	}
	return list, fmt.Errorf("no to-do with id %d", id)
}

// UpdateDescription finds an item by id and replaces its Description.
// Returns a new slice (copy-on-write style) to make the mutation explicit.
func UpdateDescription(list []Item, id int, newDesc string) ([]Item, error) {
	newDesc = strings.TrimSpace(newDesc)
	if newDesc == "" {
		return list, errors.New("new description cannot be empty")
	}
	for i := range list {
		if list[i].ID == id {
			list[i].Description = newDesc
			return list, nil
		}
	}
	return list, fmt.Errorf("no to-do with id %d", id)
}

// Delete removes an item by id. If the id does not exist, returns an error.
// Returns the shortened slice to the caller.
func Delete(list []Item, id int) ([]Item, error) {
	for i := range list {
		if list[i].ID == id {
			return append(list[:i], list[i+1:]...), nil
		}
	}
	return list, fmt.Errorf("no to-do with id %d", id)
}
