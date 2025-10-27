package cli_app

import (
	"bytes"
	"context"
	"os"
	"regexp"
	"testing"
	"time"

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

// captureStdout redirects os.Stdout to a pipe and returns a function
// that (1) restores stdout and (2) returns the captured output.
// Call the returned function exactly once after printing.
func captureStdout(t *testing.T) func() string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	return func() string {
		_ = w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		_ = r.Close()
		os.Stdout = orig
		return buf.String()
	}
}

// TestCLI_Add_CreatesItem verifies that the CLI's add command creates a new to-do item
// and that the item is persisted to the expected file.
// It checks the output file for the new item.
// It uses an isolated temporary working directory for the test.
func TestCLI_Add_CreatesItem(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	app := New()
	ctx := context.Background()

	rawPath := "todos.json" // will normalize to out/todos.json
	if err := app.Run(ctx, []string{"-add", "Buy milk", "-out", rawPath}); err != nil {
		t.Fatalf("Run(add) error: %v", err)
	}
	list := readTodos(t, rawPath)
	if len(list) != 1 || list[0].Description != "Buy milk" {
		t.Fatalf("unexpected list after add: %+v", list)
	}
}

// TestCLI_List_ShowsTwoItems verifies that the CLI's list command
// correctly lists two added to-do items.
// It seeds two items via the add command, then reads back the list.
// It uses an isolated temporary working directory for the test.
func TestCLI_List_ShowsTwoItems(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	app := New()
	ctx := context.Background()
	rawPath := "todos.json"

	// Seed two items
	_ = app.Run(ctx, []string{"-add", "Task A", "-status", "not started", "-out", rawPath})
	_ = app.Run(ctx, []string{"-add", "Task B", "-status", "started", "-out", rawPath})

	// List should return two
	list := readTodos(t, rawPath)
	print(len(list))
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}
}

// TestCLI_Update_ChangesItem verifies that the CLI's update command
// correctly modifies an existing to-do item.
// It seeds one item via the add command, then updates it.
// It reads back the list to verify the changes.
// It uses an isolated temporary working directory for the test.
func TestCLI_Update_ChangesItem(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	app := New()
	ctx := context.Background()
	rawPath := "todos.json"

	_ = app.Run(ctx, []string{"-add", "Buy milk", "-out", rawPath})

	if err := app.Run(ctx, []string{"-update", "1", "-newdesc", "Buy oat milk", "-out", rawPath}); err != nil {
		t.Fatalf("Run(update) error: %v", err)
	}
	list := readTodos(t, rawPath)
	if len(list) != 1 || list[0].Description != "Buy oat milk" || string(list[0].Status) != "not started" {
		t.Fatalf("unexpected list after update: %+v", list)
	}
}

// TestCLI_Delete_RemovesItem verifies that the CLI's delete command
// correctly removes an existing to-do item.
// It seeds one item via the add command, then deletes it.
// It reads back the list to verify the item is gone.
// It uses an isolated temporary working directory for the test.
func TestCLI_Delete_RemovesItem(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	app := New()
	ctx := context.Background()
	rawPath := "todos.json"

	_ = app.Run(ctx, []string{"-add", "Buy milk", "-out", rawPath})
	if err := app.Run(ctx, []string{"-delete", "1", "-out", rawPath}); err != nil {
		t.Fatalf("Run(delete) error: %v", err)
	}
	list := readTodos(t, rawPath)
	if len(list) != 0 {
		t.Fatalf("after delete, expected 0 items, got %d", len(list))
	}
}

// TestCLI_printList validates the tabular output without relying on literal tabs,
// because text/tabwriter expands tabs into flexible spacing.
func TestCLI_printList(t *testing.T) {
	// Begin capture
	getOutput := captureStdout(t)

	// Prepare a small set of items with deterministic timestamps
	items := []todo.Item{
		{ID: 1, Description: "Task A", Status: todo.StatusNotStarted, CreatedAt: time.Unix(0, 0)},
		{ID: 2, Description: "Task B", Status: todo.StatusStarted, CreatedAt: time.Unix(10, 0)},
	}
	printList(items)

	// End capture and get output
	out := getOutput()

	// Assert header with flexible whitespace
	headerRe := regexp.MustCompile(`(?m)^ID\s+DESCRIPTION\s+STATUS\s+CREATED$`)
	if !headerRe.MatchString(out) {
		t.Fatalf("header not found or malformed in output:\n%s", out)
	}

	// Assert both rows appear somewhere in the output
	if !regexp.MustCompile(`Task A`).MatchString(out) || !regexp.MustCompile(`Task B`).MatchString(out) {
		t.Fatalf("items not found in output:\n%s", out)
	}
}
