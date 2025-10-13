package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"todo-cli/todo"
)

// PrintList displays the list of items in a tabular format.
func PrintList(list []todo.Item) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDESCRIPTION\tSTATUS\tCREATED")
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Description, t.Status, t.CreatedAt.Format(time.RFC3339))
	}
	_ = w.Flush()
}
