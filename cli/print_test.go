package cli

import (
	"bytes"
	"os"
	"regexp"
	"testing"
	"time"

	"todo-cli/todo"
)

// captureStdout returns a function that restores stdout and the captured output.
// Call the returned function exactly once to both restore stdout and get the output.
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

func TestPrintList(t *testing.T) {
	// Start capturing
	getOutput := captureStdout(t)

	items := []todo.Item{
		{ID: 1, Description: "Task A", Status: todo.StatusNotStarted, CreatedAt: time.Unix(0, 0)},
		{ID: 2, Description: "Task B", Status: todo.StatusStarted, CreatedAt: time.Unix(10, 0)},
	}
	PrintList(items)

	// Stop capturing and get the output (also restores stdout)
	out := getOutput()

	// tabwriter expands tabs to spaces; assert with flexible whitespace.
	headerRe := regexp.MustCompile(`(?m)^ID\s+DESCRIPTION\s+STATUS\s+CREATED$`)
	if !headerRe.MatchString(out) {
		t.Fatalf("header not found or malformed in output:\n%s", out)
	}

	if !regexp.MustCompile(`Task A`).MatchString(out) || !regexp.MustCompile(`Task B`).MatchString(out) {
		t.Fatalf("items not found in output:\n%s", out)
	}
}
