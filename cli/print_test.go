// NOTE: Simple stdlib/glue coverage; tests kept minimal per guidance.
package cli

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"time"

	"todo-app/todo"
)

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

// TestPrintList validates the tabular output without relying on literal tabs,
// because text/tabwriter expands tabs into flexible spacing.
func TestPrintList(t *testing.T) {
	// Begin capture
	getOutput := captureStdout(t)

	// Prepare a small set of items with deterministic timestamps
	items := []todo.Item{
		{ID: 1, Description: "Task A", Status: todo.StatusNotStarted, CreatedAt: time.Unix(0, 0)},
		{ID: 2, Description: "Task B", Status: todo.StatusStarted, CreatedAt: time.Unix(10, 0)},
	}
	PrintList(items)

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
