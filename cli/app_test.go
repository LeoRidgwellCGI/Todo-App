package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"todo-app/todo"
)

// readTodos loads todos using the same normalization logic the CLI uses.
// This mirrors how the real CLI resolves -out into ./out/<basename>.
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

// TestAppRun_Add_Update_Delete_List_WithOutDir exercises the CLI happy path
// in a temporary working directory so that "./out" is sandboxed per test run.
func TestAppRun_Add_Update_Delete_List_WithOutDir(t *testing.T) {
	// Isolate test side effects under a temp working directory.
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

	// Intentionally pass a non-"out" path; CLI will normalize to "out/todos.json".
	rawPath := "todos.json"
	normPath := normalizeOutPath(rawPath) // -> out/todos.json (platform path separators may vary)

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

	// UPDATE (id 1 -> new desc)
	if err := app.Run(ctx, []string{"-update", "1", "-newdesc", "Buy oat milk", "-out", rawPath}); err != nil {
		t.Fatalf("Run(update) error: %v", err)
	}
	list = readTodos(t, rawPath)
	if got := list[0].Description; got != "Buy oat milk" {
		t.Fatalf("after update, desc=%q", got)
	}

	// LIST (should not error and should not change file)
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

// TestNormalizeOutPath verifies path normalization to "./out/<basename>"
// and uses filepath.ToSlash for cross-platform comparisons.
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
		// Convert to forward slashes so assertions pass on Windows as well.
		got = filepath.ToSlash(got)
		if got != tt.want {
			t.Fatalf("normalizeOutPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
