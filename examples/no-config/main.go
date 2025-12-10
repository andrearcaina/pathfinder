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
	scanNoConfig()
}

func scanNoConfig() {
	report, err := pathfinder.Scan(pathfinder.Config{})
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	fmt.Println("Scan with default configuration:")
	printReport(report)
	fmt.Println("-------------------------------")
}

func printReport(report pathfinder.CodebaseReport) {
	fmt.Printf("Found %d files across %d languages.\n\n",
		report.CodebaseMetrics.TotalFiles,
		report.CodebaseMetrics.TotalLanguages)

	fmt.Println("Scanned Files:")
	for _, file := range report.ScannedFiles() {
		fmt.Printf("• %s\n", file)
	}
	fmt.Println()

	fmt.Println("Scanned Directories:")
	for _, dir := range report.ScannedDirectories() {
		fmt.Printf("• %s\n", dir)
	}
	fmt.Println()

	fmt.Println("Scanned Languages:")
	for _, lang := range report.ScannedLanguages() {
		fmt.Printf("• %s\n", lang)
	}
	fmt.Println()

	fmt.Println("Language Breakdown:")
	for _, lang := range report.LanguageMetrics {
		fmt.Printf("• %s: %d lines (%.2f%%)\n",
			lang.Metrics.Language,
			lang.Metrics.Lines,
			lang.Percentage,
		)
	}
}
