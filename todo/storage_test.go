package todo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	// Start with empty
	items := []Item{}
	if err := Save(ctx, items, path); err != nil {
		t.Fatalf("Save(empty) error: %v", err)
	}

	// Add some data and save
	items = append(items, Item{ID: 1, Description: "Alpha", Status: StatusNotStarted})
	items = append(items, Item{ID: 2, Description: "Beta", Status: StatusStarted})
	if err := Save(ctx, items, path); err != nil {
		t.Fatalf("Save(items) error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Save() did not create file: %v", err)
	}

	// Load and assert
	got, err := Load(ctx, path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(got) != 2 || got[0].Description != "Alpha" || got[1].Status != StatusStarted {
		t.Fatalf("Load() got=%+v", got)
	}
}

func TestLoadMissingReturnsEmpty(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")

	got, err := Load(ctx, path)
	if err != nil {
		t.Fatalf("Load(missing) error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Load(missing) expected empty slice, got=%+v", got)
	}
}
