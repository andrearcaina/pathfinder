package main

import (
	"fmt"
	"log"

	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
)

func main() {
	fmt.Printf("Supported Pathfinder version: %s\n\n", pathfinder.Version())
	fmt.Printf("Supported Languages: %v\n\n", pathfinder.SupportedLanguages())

	fmt.Println("-------------------------------")
	scanWithConfig()
}

func scanWithConfig() {
	config := pathfinder.Config{
		PathFlag:       "../..",
		RecursiveFlag:  true,
		HiddenFlag:     false,
		DependencyFlag: true,
		BufferSizeFlag: 4,
		MaxDepthFlag:   -1,
		ThroughputFlag: true,
	}

	report, err := pathfinder.Scan(config)
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	fmt.Println("Scan with custom configuration and throughput mode activated:")
	printReport(report)
	fmt.Println("-------------------------------")
}

func printReport(report pathfinder.CodebaseReport) {
	for _, worker := range report.PerformanceMetrics.WorkerStats {
		fmt.Printf("[Worker %d] processed %d files in %.2fs (%.1f files/sec)\n",
			worker.Id, worker.Processed, worker.Duration, worker.Throughput,
		)
	}

	fmt.Println()

	fmt.Printf("Total workers: %d\n", report.PerformanceMetrics.TotalWorkers)
	fmt.Printf("Total scanned files: %d\n", report.CodebaseMetrics.TotalFiles)
	fmt.Printf("Total scanned dirs: %d\n", report.CodebaseMetrics.TotalDirs)
	fmt.Printf("Total lines %d\n", report.CodebaseMetrics.TotalLines)
	fmt.Printf("Total time taken: %.2fs\n", report.PerformanceMetrics.TotalTimeSeconds)
	fmt.Printf("Overall throughput: %.1f files/sec\n", report.PerformanceMetrics.OverallThroughput)
}
