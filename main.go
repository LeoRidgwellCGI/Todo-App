package main

import (
	"fmt"
	"os"

	"todo-cli/cli"
)

// main is the application entrypoint located in the root directory.
// It delegates execution to the CLI logic in the cli package.
func main() {
	if err := cli.New().Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
