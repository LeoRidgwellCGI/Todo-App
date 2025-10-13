package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"todo-cli/todo"
)

// helper to read todos from file using the todo package
func readTodos(t *testing.T, path string) []todo.Item {
	t.Helper()
	ctx := context.Background()
	list, err := todo.Load(ctx, path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return list
}

func TestAppRun_Add_Update_Delete_List(t *testing.T) {
	app := New()
	ctx := context.Background()

	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	// Ensure file exists (optional; Save will create it anyway)
	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		t.Fatalf("prep write: %v", err)
	}

	// ADD
	if err := app.Run(ctx, []string{"-add", "Buy milk", "-out", path}); err != nil {
		t.Fatalf("Run(add) error: %v", err)
	}
	list := readTodos(t, path)
	if len(list) != 1 || list[0].Description != "Buy milk" {
		b, _ := json.Marshal(list)
		t.Fatalf("after add, got=%s", string(b))
	}

	// UPDATE
	if err := app.Run(ctx, []string{"-update", "1", "-newdesc", "Buy oat milk", "-out", path}); err != nil {
		t.Fatalf("Run(update) error: %v", err)
	}
	list = readTodos(t, path)
	if list[0].Description != "Buy oat milk" {
		t.Fatalf("after update, desc=%q", list[0].Description)
	}

	// LIST (should not error)
	if err := app.Run(ctx, []string{"-list", "-out", path}); err != nil {
		t.Fatalf("Run(list) error: %v", err)
	}

	// DELETE
	if err := app.Run(ctx, []string{"-delete", "1", "-out", path}); err != nil {
		t.Fatalf("Run(delete) error: %v", err)
	}
	list = readTodos(t, path)
	if len(list) != 0 {
		t.Fatalf("after delete, expected 0 items, got %d", len(list))
	}
}
