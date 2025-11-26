package main

import (
	"fmt"
	"log"

	// Import the pathfinder package
	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
)

func main() {
	config := pathfinder.Config{
		PathFlag:       ".",   // Scan the current directory
		RecursiveFlag:  true,  // Scan subdirectories recursively
		HiddenFlag:     false, // Skip hidden files/directories (like .git)
		DependencyFlag: true,  // Analyze dependency files (go.mod, package.json, etc.)
		BufferSizeFlag: 4096,  // Set read buffer size (4KB is usually optimal)
		MaxDepthFlag:   -1,    // No limit on recursion depth
	}

	report, err := pathfinder.Scan(&config)
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	fmt.Printf("Supported PathFinder version: %s\n\n", pathfinder.Version())

	fmt.Printf("Supported Languages: %v\n\n", pathfinder.GetSupportedLanguages())

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
