package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"todo-app/todo"
)

// Store abstracts persistence for to-do lists.
type Store interface {
	Load(ctx context.Context) ([]todo.Item, error)
	Save(ctx context.Context, list []todo.Item) error
}

// FileStore implements Store backed by a JSON file on disk.
type FileStore struct {
	// OutPath is the JSON file path.
	OutPath string
}

func NewFileStore(outPath string) *FileStore {
	return &FileStore{OutPath: outPath}
}

func (f *FileStore) ensureOutPath() string {
	if f.OutPath == "" {
		return filepath.Join("out", "todos.json")
	}
	// If OutPath is a bare filename (no path separators), write under ./out/
	if !strings.Contains(f.OutPath, "/") && !strings.Contains(f.OutPath, `\`) {
		return filepath.Join("out", f.OutPath)
	}
	return f.OutPath
}

func (f *FileStore) Load(ctx context.Context) ([]todo.Item, error) {
	path := f.ensureOutPath()
	list, err := todo.Load(ctx, path)
	if err != nil {
		slog.ErrorContext(ctx, "load failed", "error", err, "path", path)
		return nil, err
	}
	return list, nil
}

func (f *FileStore) Save(ctx context.Context, list []todo.Item) error {
	path := f.ensureOutPath()
	// Ensure directory exists (robust even if todo.WriteJSON already does this)
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := todo.Save(ctx, list, path); err != nil {
		slog.ErrorContext(ctx, "save failed", "error", err, "path", path)
		return err
	}
	return nil
}

// FindByID returns the matching item or false if not found.
func FindByID(list []todo.Item, id int) (todo.Item, bool) {
	for i := range list {
		if list[i].ID == id {
			return list[i], true
		}
	}
	return todo.Item{}, false
}
