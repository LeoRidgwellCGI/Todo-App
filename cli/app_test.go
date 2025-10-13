package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"todo-cli/todo"
)

// readTodos reads using the normalized output path (under ./out/).
func readTodos(t *testing.T, path string) []todo.Item {
	t.Helper()
	ctx := context.Background()
	norm := normalizeOutPath(path)
	list, err := todo.Load(ctx, norm)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	return list
}

func TestAppRun_Add_Update_Delete_List_WithOutDir(t *testing.T) {
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	app := New()
	ctx := context.Background()

	rawPath := "todos.json"
	normPath := normalizeOutPath(rawPath) // -> out/todos.json

	// ADD
	if err := app.Run(ctx, []string{"-add", "Buy milk", "-out", rawPath}); err != nil {
		t.Fatalf("Run(add) error: %v", err)
	}
	list := readTodos(t, rawPath)
	if len(list) != 1 || list[0].Description != "Buy milk" {
		b, _ := json.Marshal(list)
		t.Fatalf("after add, got=%s", string(b))
	}
	if _, err := os.Stat(normPath); err != nil {
		t.Fatalf("expected file at %s; err=%v", normPath, err)
	}

	// UPDATE
	if err := app.Run(ctx, []string{"-update", "1", "-newdesc", "Buy oat milk", "-out", rawPath}); err != nil {
		t.Fatalf("Run(update) error: %v", err)
	}
	list = readTodos(t, rawPath)
	if got := list[0].Description; got != "Buy oat milk" {
		t.Fatalf("after update, desc=%q", got)
	}

	// LIST
	if err := app.Run(ctx, []string{"-list", "-out", rawPath}); err != nil {
		t.Fatalf("Run(list) error: %v", err)
	}

	// DELETE
	if err := app.Run(ctx, []string{"-delete", "1", "-out", rawPath}); err != nil {
		t.Fatalf("Run(delete) error: %v", err)
	}
	list = readTodos(t, rawPath)
	if len(list) != 0 {
		t.Fatalf("after delete, expected 0 items, got %d", len(list))
	}
}

func TestNormalizeOutPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "out/todos.json"},
		{"todos.json", "out/todos.json"},
		{"./todos.json", "out/todos.json"},
		{"out/todos.json", "out/todos.json"},
		{"/tmp/abc.json", "out/abc.json"},
		{"nested/dir/abc.json", "out/abc.json"},
	}
	for _, tt := range tests {
		got := normalizeOutPath(tt.in)
		// Normalize separators to forward slashes for cross-platform comparison.
		got = filepath.ToSlash(got)
		if got != tt.want {
			t.Fatalf("normalizeOutPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
