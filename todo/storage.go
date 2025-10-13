package todo

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
)

// Save writes the current list to disk and logs a structured event.
func Save(list []Item, path string) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Error("failed to save todos", "error", err, "path", path)
		return err
	}
	slog.Info("todos saved", "path", path, "count", len(list))
	return nil
}

// Load reads todos from a JSON file.
// If missing, returns an empty list.
func Load(path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Item{}, nil
		}
		slog.Error("failed to read file", "error", err, "path", path)
		return nil, err
	}
	if len(b) == 0 {
		return []Item{}, nil
	}
	var list []Item
	if err := json.Unmarshal(b, &list); err != nil {
		slog.Error("failed to unmarshal JSON", "error", err, "path", path)
		return nil, err
	}
	return list, nil
}
