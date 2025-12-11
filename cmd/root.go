package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pathfinder",
	Short: "a CLI tool to track and map codebases for metrics",
	Long: `Pathfinder is a CLI tool designed to help developers track and map their codebases.
It scans a directory and reports per-language lines of code with percentages, plus extra codebase metrics (TODO).

Examples:
  pathfinder scan
  pathfinder scan -p /path/to/codebase`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(versionCmd)
}
