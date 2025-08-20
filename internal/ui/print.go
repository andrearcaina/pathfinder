package ui

import (
	"fmt"
	"strings"

	"github.com/andrearcaina/pathfinder/internal/metrics"
)

func PrintReport(report metrics.CodebaseReport) {
	titleStyle := TitleStyle()
	sectionStyle := SectionStyle()
	badgeDisplay := []string{
		BadgeDisplay("🗃️ Files", FormatIntBritishEnglish(report.CodebaseMetrics.TotalFiles)),
		BadgeDisplay("📂 Directories", FormatIntBritishEnglish(report.CodebaseMetrics.TotalDirs)),
		BadgeDisplay("🧑‍💻 Languages", FormatIntBritishEnglish(report.CodebaseMetrics.TotalLanguages)),
		BadgeDisplay("📊 Total Lines", FormatIntBritishEnglish(report.CodebaseMetrics.TotalLines)),
		BadgeDisplay("🖥️ Lines of Code", FormatIntBritishEnglish(report.CodebaseMetrics.TotalCode)),
		BadgeDisplay("💬 Comments", FormatIntBritishEnglish(report.CodebaseMetrics.TotalComments)),
		BadgeDisplay("🗑️ Blank Lines", FormatIntBritishEnglish(report.CodebaseMetrics.TotalBlanks)),
	}

	fmt.Println(titleStyle.Render("☁️ Pathfinder - Map and Track Your Codebase"))

	fmt.Println(strings.Join(badgeDisplay, " "))

	fmt.Println(sectionStyle.Render("📋 Language Breakdown"))

	// TODO: use lipgloss
	// this will be improved later, as it's not very readable if there's many languages, files, and directories
	for i := 0; i < len(report.LanguageMetrics); i++ {
		fmt.Printf("Stat for Language %s:\n", report.LanguageMetrics[i].Metrics.Language)
		fmt.Println("--------------------------")
		fmt.Printf("Percentage: %.2f%%\n", report.LanguageMetrics[i].Percentage)
		fmt.Printf("# of Files: %s\n", FormatIntBritishEnglish(report.LanguageMetrics[i].Metrics.Files))
		fmt.Printf("Lines of Code: %s\n", FormatIntBritishEnglish(report.LanguageMetrics[i].Metrics.Code))
		fmt.Printf("Comment Lines: %s\n", FormatIntBritishEnglish(report.LanguageMetrics[i].Metrics.Comments))
		fmt.Printf("Blank Lines: %s\n", FormatIntBritishEnglish(report.LanguageMetrics[i].Metrics.Blanks))
		fmt.Printf("Total Lines: %s\n", FormatIntBritishEnglish(report.LanguageMetrics[i].Metrics.Lines))
		fmt.Println("--------------------------")
	}

	fmt.Println(sectionStyle.Render("📄 File Breakdown"))
	for i := 0; i < len(report.FileMetrics); i++ {
		fmt.Printf("File: %s\n", report.FileMetrics[i].Path)
		fmt.Println("--------------------------")
		fmt.Printf("Lines of Code: %s\n", FormatIntBritishEnglish(report.FileMetrics[i].Metrics.Code))
		fmt.Printf("Comment Lines: %s\n", FormatIntBritishEnglish(report.FileMetrics[i].Metrics.Comments))
		fmt.Printf("Blank Lines: %s\n", FormatIntBritishEnglish(report.FileMetrics[i].Metrics.Blanks))
		fmt.Printf("Total Lines: %s\n", FormatIntBritishEnglish(report.FileMetrics[i].Metrics.Lines))
		fmt.Println("--------------------------")
		if i < len(report.FileMetrics)-1 {
			fmt.Println()
		}
	}

	fmt.Println(sectionStyle.Render("📂 Directory Breakdown"))
	for i := 0; i < len(report.DirMetrics); i++ {
		fmt.Printf("Directory: %s\n", report.DirMetrics[i].Directory)
		fmt.Println("--------------------------")
		fmt.Printf("Total Lines: %s\n", FormatIntBritishEnglish(report.DirMetrics[i].Lines))
		fmt.Println("--------------------------")
	}

	fmt.Printf("Total TODOs: %s\n", FormatIntBritishEnglish(report.AnnotationMetrics.TotalTODO))
	fmt.Printf("Total FIXMEs: %s\n", FormatIntBritishEnglish(report.AnnotationMetrics.TotalFIXME))
	fmt.Printf("Total HACKs: %s\n", FormatIntBritishEnglish(report.AnnotationMetrics.TotalHACK))
	fmt.Printf("Total Annotations: %s\n", FormatIntBritishEnglish(report.AnnotationMetrics.TotalAnnotations))
}
