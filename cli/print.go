package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"todo-app/todo"
)

//
// cli/print.go (package cli)
// --------------------------
// The PrintList helper renders items in a tabular layout using text/tabwriter.
// This keeps presentation concerns out of business logic.
//

// PrintList prints a simple fixed table to stdout.
// We rely on tabwriter to align columns regardless of content width.
// NOTE: stdout is for user-facing output; logs go to stderr via slog.
func PrintList(list []todo.Item) {
	// Create a writer that aligns columns based on tab stops.
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	// Header line (columns are separated by tabs; tabwriter turns tabs into padding).
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")

	// Body rows
	for _, t := range list {
		// Time is formatted as RFC3339 for easy machine readability and consistency.
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}

	// Flush to ensure content is rendered even if buffers are not full.
	_ = w.Flush()
}
