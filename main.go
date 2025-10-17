// This repository provides multiple entrypoints (CLI and API).
// You should not run the root package directly.
//
// Usage:
//
//	CLI (recommended):
//	  go run ./cmd/cli [args]
//	  go build -o bin/todo ./cmd/cli
//	  go install ./cmd/cli
//
//	API (coming soon, will live at ./cmd/api):
//	  go run ./cmd/api
//	  go build -o bin/todo-api ./cmd/api
//	  go install ./cmd/api
//
// Background:
//   - The root module intentionally has no runnable application logic so that
//     shared code can live at the top-level without making the repo itself a program.
//   - Put executables under ./cmd/<name> following Go's standard project layout.
//
// If you reached this by habitually running `go run .`, try the commands above.
// If you're unsure which mode to use, you probably want the CLI: ./cmd/cli.
package main

import (
	"fmt"
	"os"
)

func main() {
	msg := `
🚫 This package is not meant to be run directly.

Run one of the entrypoints instead:

  • CLI:
      go run ./cmd/cli [args]
      go build -o bin/todo ./cmd/cli

  • API (coming soon at ./cmd/api):
      go run ./cmd/api
      go build -o bin/todo-api ./cmd/api

Tip: use "go install ./cmd/cli" (or ./cmd/api later) to put the binary on your PATH.
`
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(2)
}
