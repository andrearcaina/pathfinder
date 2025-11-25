package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andrearcaina/pathfinder/internal/export"
	"github.com/andrearcaina/pathfinder/internal/ui"
	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
	"github.com/spf13/cobra"
)

var (
	debugFlag      bool
	pathFlag       string
	hiddenFlag     bool
	bufferSizeFlag int
	recursiveFlag  bool
	maxDepthFlag   int
	formatFlag     string
	outputFlag     string
	dependencyFlag bool
	gitFlag        bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "scan is a subcommand to scan a codebase and report metrics",
	Long: `scan is a subcommand to scan a codebase and report metrics. Examples are:

pf scan
pf scan -R
pf scan -p /path/to/codebase
pf scan -p /path/to/codebase -R -m 3 -i -b 16
pf scan -p /path/to/codebase -R -m 3 -f json -o report.json,
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !recursiveFlag && maxDepthFlag != -1 {
			return errors.New("--max-depth flag is ignored when --recursive is false")
		}

		pathFlag, err := filepath.Abs(pathFlag)
		if err != nil {
			return err
		}

		if bufferSizeFlag != 4 && bufferSizeFlag != 8 && bufferSizeFlag != 16 && bufferSizeFlag != 32 && bufferSizeFlag != 64 {
			return errors.New("invalid Buffer Size. Allowed values are 4, 8, 16, 32, 64 (in KB)")
		}

		flags := &pathfinder.Config{
			PathFlag:       pathFlag,
			HiddenFlag:     hiddenFlag,
			BufferSizeFlag: bufferSizeFlag * 1024, // convert KB to bytes for internal use
			RecursiveFlag:  recursiveFlag,
			MaxDepthFlag:   maxDepthFlag,
			DependencyFlag: dependencyFlag,
			GitFlag:        gitFlag,
		}

		report, err := pathfinder.Scan(flags)
		if err != nil {
			return err
		}

		if debugFlag { // print raw report for debugging (this is just printing the struct, not really "debugging")
			fmt.Printf("Debug: %+v\n", report)
			return nil
		}

		// validate that both format and output are set together
		if (formatFlag != "" && outputFlag == "") || (outputFlag != "" && formatFlag == "") {
			return errors.New("both --format and --output flags must be set together")
		}

		// handle export if output and format flags are set
		if outputFlag != "" && formatFlag != "" {
			if strings.ToLower(filepath.Ext(outputFlag)) == "" {
				return errors.New("output file must have an extension (e.g. .json)")
			}

			formatFlag = strings.ToLower(formatFlag)
			if formatFlag == "json" && strings.HasSuffix(outputFlag, ".json") {
				if err := export.CreateJSON(report, outputFlag); err != nil {
					return err
				}
				return nil
			} else {
				return fmt.Errorf("unsupported format '%s'. Supported formats: json", formatFlag)
			}
		}

		ui.PrintReport(report)
		return nil
	},
}

func init() {
	scanCmd.Flags().BoolVarP(&debugFlag, "debug", "", false, "Enable debug mode")
	scanCmd.Flags().StringVarP(&pathFlag, "path", "p", ".", "Path to codebase/repository")
	scanCmd.Flags().BoolVarP(&hiddenFlag, "hidden", "i", false, "Include hidden files and directories")
	scanCmd.Flags().IntVarP(&bufferSizeFlag, "buffer-size", "b", 4, "Buffer size for reading files in KB. Options are 4, 8, 16, 32, 64")
	scanCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "R", false, "Scan directories recursively")
	scanCmd.Flags().IntVarP(&maxDepthFlag, "max-depth", "m", -1, "Maximum recursion depth. Only works if --recursive is set")
	scanCmd.Flags().StringVarP(&formatFlag, "format", "f", "", "Output format. Options are: json")
	scanCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Sets output file name.")
	scanCmd.Flags().BoolVarP(&dependencyFlag, "dependencies", "d", false, "Scan for dependencies (supported for some languages)")
	scanCmd.Flags().BoolVarP(&gitFlag, "git", "g", false, "Scan for git information (e.g. number of commits, git history, etc.)")
}
