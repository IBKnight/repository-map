// Package cmd defines the repomap CLI: the "map" subcommand generates a
// repository map, and the "agents" subcommand scaffolds AI agent
// instruction files.
package cmd

import (
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags "-X github.com/IBKnight/repository-map/cmd.version=vX.Y.Z".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "repomap",
	Short: "Generate repository maps and AI agent instruction files",
	Long: `repomap is a CLI tool with two main capabilities:

  1. "repomap map"    — build a Markdown or text map of an existing
                         repository: directory tree, language breakdown,
                         and file counts.

  2. "repomap agents" — scaffold instruction/skill files for AI coding
                         agents and assistants (Claude Code's CLAUDE.md
                         and .claude/skills, GitHub Copilot, Cursor, and
                         a generic AGENTS.md) based on the repository's
                         detected languages and layout.`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
