// Command repomap generates a Markdown/text map of a repository and can
// scaffold instruction files for AI coding agents (Claude Code, GitHub
// Copilot, Cursor, and generic AGENTS.md-based tools).
package main

import (
	"fmt"
	"os"

	"github.com/IBKnight/repository-map/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
