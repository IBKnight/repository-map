package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the repomap version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("repomap " + version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
