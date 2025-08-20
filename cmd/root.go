package cmd

import (
	"os"
	"path/filepath"

	"github.com/andrearcaina/pathfinder/internal/metrics"
	"github.com/andrearcaina/pathfinder/internal/runner"
	"github.com/spf13/cobra"
)

var (
	pathFlag   string
	hiddenFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "pathfinder",
	Short: "a CLI tool to track and map codebases for metrics",
	Long: `Pathfinder is a CLI tool designed to help developers track and map their codebases.
It scans a directory and reports per-language lines of code with percentages, plus extra codebase metrics (TODO).

Examples:
  pathfinder
  pathfinder --path /path/to/codebase`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pathFlag, err := filepath.Abs(pathFlag)
		if err != nil {
			return err
		}

		return runner.Run(metrics.Flags{
			PathFlag:   pathFlag,
			HiddenFlag: hiddenFlag,
		})
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&pathFlag, "path", "p", ".", "Path to codebase/repository")
	rootCmd.Flags().BoolVarP(&hiddenFlag, "hidden", "i", false, "Include hidden files and directories")
}
