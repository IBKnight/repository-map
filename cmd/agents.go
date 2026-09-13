package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/IBKnight/repository-map/internal/agents"
	"github.com/IBKnight/repository-map/internal/scanner"
)

var (
	agentsOutputDir string
	agentsTargets   []string
	agentsForce     bool
	agentsName      string
)

var agentsCmd = &cobra.Command{
	Use:   "agents [path]",
	Short: "Scaffold instruction files for AI coding agents",
	Long: `Analyzes the given repository (default: current directory) — its
languages, top-level layout, and detected build tooling — and generates
instruction files for AI coding agents and assistants:

  claude   CLAUDE.md and a .claude/skills/<project>-overview/SKILL.md
  copilot  .github/copilot-instructions.md
  cursor   .cursor/rules/repository.mdc
  generic  AGENTS.md (a convention read by several agentic coding tools)

Existing files are left untouched unless --force is given.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}

		targets, err := agents.ParseTargets(agentsTargets)
		if err != nil {
			return err
		}

		root, err := scanner.Scan(path, scanner.Options{RespectGitignore: true})
		if err != nil {
			return fmt.Errorf("scanning %s: %w", path, err)
		}

		name := agentsName
		if name == "" {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			name = filepath.Base(abs)
		}

		analysis := agents.Analyze(root, name)

		outDir := agentsOutputDir
		if outDir == "" {
			outDir = path
		}

		result, err := agents.Generate(analysis, targets, outDir, agentsForce)
		if err != nil {
			return fmt.Errorf("generating agent files: %w", err)
		}

		for _, f := range result.Written {
			fmt.Fprintf(cmd.OutOrStdout(), "wrote   %s\n", f)
		}
		for _, f := range result.Skipped {
			fmt.Fprintf(cmd.OutOrStdout(), "skipped %s (already exists, use --force to overwrite)\n", f)
		}

		return nil
	},
}

func init() {
	agentsCmd.Flags().StringVarP(&agentsOutputDir, "output-dir", "o", "", "directory to write files into (default: the scanned path)")
	agentsCmd.Flags().StringSliceVarP(&agentsTargets, "target", "t", nil, "targets to generate: claude, copilot, cursor, generic, all (default: all)")
	agentsCmd.Flags().BoolVar(&agentsForce, "force", false, "overwrite existing files")
	agentsCmd.Flags().StringVar(&agentsName, "name", "", "project name to use in generated files (default: directory name)")

	rootCmd.AddCommand(agentsCmd)
}
