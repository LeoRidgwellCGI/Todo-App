package todo

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// ensureParentDir ensures the directory for a file path exists.
func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}

// Save writes the current list to disk (under ./out) and logs a structured event.
func Save(ctx context.Context, list []Item, path string) error {
	if err := ensureParentDir(path); err != nil {
		slog.ErrorContext(ctx, "failed to create output directory", "error", err, "path", path)
		return err
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		slog.ErrorContext(ctx, "failed to marshal todos", "error", err, "path", path)
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.ErrorContext(ctx, "failed to save todos", "error", err, "path", path)
		return err
	}
	slog.InfoContext(ctx, "todos saved", "path", path, "count", len(list))
	return nil
}

// Load reads todos from a JSON file. If missing, returns an empty list.
func Load(ctx context.Context, path string) ([]Item, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Item{}, nil
		}
		slog.ErrorContext(ctx, "failed to read file", "error", err, "path", path)
		return nil, err
	}
	if len(b) == 0 {
		return []Item{}, nil
	}
	var list []Item
	if err := json.Unmarshal(b, &list); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal JSON", "error", err, "path", path)
		return nil, err
	}
	return list, nil
}
