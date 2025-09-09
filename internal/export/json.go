package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrearcaina/pathfinder/internal/metrics"
)

func CreateJSON(report metrics.CodebaseReport, outputPath string) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file %s: %w", outputPath, err)
	}

	fmt.Println("Report written to " + outputPath)
	return nil
}
