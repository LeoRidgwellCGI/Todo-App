// NOTE: Simple stdlib/glue coverage; tests kept minimal per guidance.
package todo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestSaveAndLoad verifies end-to-end persistence:
// - Save creates/overwrites the file
// - Load round-trips the JSON data
func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	// Start empty
	items := []Item{}
	if err := Save(ctx, items, path); err != nil {
		t.Fatalf("Save(empty) error: %v", err)
	}

	// Add some data and save again
	items = append(items, Item{ID: 1, Description: "Alpha", Status: StatusNotStarted})
	items = append(items, Item{ID: 2, Description: "Beta", Status: StatusStarted})
	if err := Save(ctx, items, path); err != nil {
		t.Fatalf("Save(items) error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Save() did not create file: %v", err)
	}

	// Load and assert data integrity
	got, err := Load(ctx, path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(got) != 2 || got[0].Description != "Alpha" || got[1].Status != StatusStarted {
		t.Fatalf("Load() got=%+v", got)
	}
}

// TestLoadMissingReturnsEmpty ensures missing files are treated as empty lists.
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
