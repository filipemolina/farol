// Command farol is a terminal to-do list: your to-do list, and your
// agent's work queue. Cobra's root command dispatches every subcommand in
// docs/DESIGN.md §9; with no subcommand it launches the TUI, and with one
// ULID-looking argument it marks that task complete.
package main

import (
	"os"

	"github.com/filipemolina/farol/src/cli"
)

func main() {
	// cli.Execute owns the exit-code contract (0/1/2, docs/DESIGN.md §9);
	// main only forwards it.
	os.Exit(cli.Execute(os.Args[1:]))
}
