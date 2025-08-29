package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrearcaina/pathfinder/internal/metrics"
	"github.com/andrearcaina/pathfinder/internal/ui"
	"github.com/spf13/cobra"
)

var (
	pathFlag       string
	hiddenFlag     bool
	bufferSizeFlag int
	recursiveFlag  bool
	maxDepthFlag   int
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
		if !recursiveFlag && maxDepthFlag != -1 {
			fmt.Println("Warning: --max-depth flag is ignored when --recursive is false.")
			return nil
		}

		pathFlag, err := filepath.Abs(pathFlag)
		if err != nil {
			return err
		}

		if bufferSizeFlag != 4 && bufferSizeFlag != 8 && bufferSizeFlag != 16 && bufferSizeFlag != 32 && bufferSizeFlag != 64 {
			fmt.Println("Invalid Buffer Size. Allowed values are 4, 8, 16, 32, 64 (in KB).")
			return nil
		}

		report, err := metrics.Analyze(metrics.Flags{
			PathFlag:       pathFlag,
			HiddenFlag:     hiddenFlag,
			BufferSizeFlag: bufferSizeFlag * 1024, // convert KB to bytes for internal use
			RecursiveFlag:  recursiveFlag,
			MaxDepthFlag:   maxDepthFlag,
		})
		if err != nil {
			return err
		}

		ui.PrintReport(report)
		return nil
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
	rootCmd.Flags().IntVarP(&bufferSizeFlag, "buffer-size", "b", 4, "Buffer size for reading files in KB. Options are 4, 8, 16, 32, 64")
	rootCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "R", false, "Scan directories recursively")
	rootCmd.Flags().IntVarP(&maxDepthFlag, "max-depth", "m", -1, "Maximum recursion depth. Only works if --recursive is set")
}
