package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/IBKnight/repository-map/internal/repomap"
	"github.com/IBKnight/repository-map/internal/scanner"
)

var (
	mapOutput      string
	mapFormat      string
	mapMaxDepth    int
	mapShowSizes   bool
	mapNoGitignore bool
)

var mapCmd = &cobra.Command{
	Use:   "map [path]",
	Short: "Generate a map of an existing repository",
	Long: `Walks the given repository (default: current directory), respecting
.gitignore rules, and produces a report of its directory structure and
language breakdown in Markdown or plain text.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) == 1 {
			path = args[0]
		}

		format := repomap.Format(mapFormat)
		if format != repomap.FormatMarkdown && format != repomap.FormatText {
			return fmt.Errorf("invalid --format %q (want: markdown, text)", mapFormat)
		}

		root, err := scanner.Scan(path, scanner.Options{
			MaxDepth:         mapMaxDepth,
			RespectGitignore: !mapNoGitignore,
		})
		if err != nil {
			return fmt.Errorf("scanning %s: %w", path, err)
		}

		out := repomap.Render(root, repomap.Options{
			Format:    format,
			TreeDepth: mapMaxDepth,
			ShowSizes: mapShowSizes,
		})

		if mapOutput == "" || mapOutput == "-" {
			fmt.Print(out)
			return nil
		}

		if err := os.WriteFile(mapOutput, []byte(out), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", mapOutput, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Wrote repository map to %s\n", mapOutput)
		return nil
	},
}

func init() {
	mapCmd.Flags().StringVarP(&mapOutput, "output", "o", "", `output file ("-" or empty for stdout)`)
	mapCmd.Flags().StringVarP(&mapFormat, "format", "f", "markdown", "output format: markdown or text")
	mapCmd.Flags().IntVar(&mapMaxDepth, "max-depth", 0, "limit directory depth (0 = unlimited)")
	mapCmd.Flags().BoolVar(&mapShowSizes, "sizes", false, "show file sizes in the tree")
	mapCmd.Flags().BoolVar(&mapNoGitignore, "no-gitignore", false, "do not apply .gitignore exclusion rules")

	rootCmd.AddCommand(mapCmd)
}
