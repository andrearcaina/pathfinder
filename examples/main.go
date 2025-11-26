package main

import (
	"fmt"
	"log"

	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
)

func main() {
	fmt.Printf("Supported PathFinder version: %s\n\n", pathfinder.Version())
	fmt.Printf("Supported Languages: %v\n\n", pathfinder.GetSupportedLanguages())

	scanWithConfig()
	scanNoConfig()
}

func scanWithConfig() {
	config := pathfinder.Config{
		PathFlag:       "..",
		RecursiveFlag:  true,
		HiddenFlag:     false,
		DependencyFlag: true,
		BufferSizeFlag: 4,
		MaxDepthFlag:   -1,
	}

	report, err := pathfinder.Scan(&config)
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	printReport(report)
}

func scanNoConfig() {
	report, err := pathfinder.Scan(nil)
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	printReport(report)
}

func printReport(report pathfinder.CodebaseReport) {
	fmt.Printf("Found %d files across %d languages.\n\n",
		report.CodebaseMetrics.TotalFiles,
		report.CodebaseMetrics.TotalLanguages)

	fmt.Println("Language Breakdown:")
	for _, lang := range report.LanguageMetrics {
		fmt.Printf("• %s: %d lines (%.2f%%)\n",
			lang.Metrics.Language,
			lang.Metrics.Lines,
			lang.Percentage,
		)
	}
}
