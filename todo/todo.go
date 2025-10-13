package todo

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Status represents the state of a to-do item.
type Status string

const (
	StatusNotStarted Status = "not started"
	StatusStarted    Status = "started"
	StatusCompleted  Status = "completed"
)

// Validate checks whether the status value is valid.
func (s Status) Validate() error {
	switch Status(strings.ToLower(string(s))) {
	case StatusNotStarted, StatusStarted, StatusCompleted:
		return nil
	default:
		return fmt.Errorf("invalid status: %q", s)
	}
}

// Item represents a single to-do item.
type Item struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// getNextID returns the next available ID.
func getNextID(list []Item) int {
	max := 0
	for _, t := range list {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}

// Add adds a new item to the list.
func Add(list *[]Item, desc string, status Status) (Item, error) {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return Item{}, errors.New("description cannot be empty")
	}
	if err := status.Validate(); err != nil {
		return Item{}, err
	}

	item := Item{
		ID:          getNextID(*list),
		Description: desc,
		Status:      Status(strings.ToLower(string(status))),
		CreatedAt:   time.Now(),
	}
	*list = append(*list, item)
	return item, nil
}

// UpdateDescription changes the description of an item by ID.
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

// Delete removes an item by ID.
func Delete(list []Item, id int) ([]Item, error) {
	for i, t := range list {
		if t.ID == id {
			return append(list[:i], list[i+1:]...), nil
		}
	}
	return list, fmt.Errorf("no to-do with id %d", id)
}
