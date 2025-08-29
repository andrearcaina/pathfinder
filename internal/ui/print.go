package ui

import (
	"fmt"
	"strings"

	"github.com/andrearcaina/pathfinder/internal/metrics"
)

const maxBarWidth = 40

func PrintReport(report metrics.CodebaseReport) {
	if report.CodebaseMetrics.TotalFiles == 0 {
		fmt.Println("No files analyzed. Please check the path and try again.")
		return // exit program
	}

	fmt.Println(TitleStyle().Render("☁️ Pathfinder • Codebase Overview"))

	fmt.Println(strings.Join([]string{
		BadgeDisplay("🗃️ Files", FormatIntBritishEnglish(report.CodebaseMetrics.TotalFiles)),
		BadgeDisplay("📂 Directories", FormatIntBritishEnglish(report.CodebaseMetrics.TotalDirs)),
		BadgeDisplay("🧑‍💻 Languages", FormatIntBritishEnglish(report.CodebaseMetrics.TotalLanguages)),
		BadgeDisplay("📊 Total Lines", FormatIntBritishEnglish(report.CodebaseMetrics.TotalLines)),
		BadgeDisplay("🖥️ Lines of Code", FormatIntBritishEnglish(report.CodebaseMetrics.TotalCode)),
		BadgeDisplay("💬 Comments", FormatIntBritishEnglish(report.CodebaseMetrics.TotalComments)),
		BadgeDisplay("🗑️ Blanks", FormatIntBritishEnglish(report.CodebaseMetrics.TotalBlanks)),
	}, " "))

	fmt.Println(SectionStyle().Render("📋 Languages"))
	for _, lang := range report.LanguageMetrics {
		fmt.Printf("  %s %.2f%%\n", lang.Metrics.Language, lang.Percentage)
		bar := BarStyle().ViewAs(lang.Percentage / 100.0)
		fmt.Printf("  %s %d lines\n", bar, lang.Metrics.Lines)
	}

	fmt.Println(SectionStyle().Render("📄 Top Files"))
	maxLines := 0
	for i := 0; i < len(report.FileMetrics); i++ {
		if report.FileMetrics[i].Metrics.Lines > maxLines {
			maxLines = report.FileMetrics[i].Metrics.Lines
		}
	}

	// TODO: handle a flag to show all files (not recommended for large codebases)
	// Only show top 10 files
	for i := 0; i < len(report.FileMetrics) && i < 10; i++ {
		f := report.FileMetrics[i]

		ratio := float64(f.Metrics.Lines) / float64(maxLines)
		bar := BarStyle().ViewAs(ratio)

		fmt.Printf("  %s • %s lines\n", f.Path, FormatIntBritishEnglish(f.Metrics.Lines))
		fmt.Println("  " + bar)
	}

	// TODO: handle a flag to show all dirs (not recommended for large codebases)
	// Only show top 10 directories
	fmt.Println(SectionStyle().Render("📂 Directories"))
	for i := 0; i < len(report.DirMetrics) && i < 10; i++ {
		d := report.DirMetrics[i]

		dirName := d.Directory

		if d.Directory == "." {
			dirName = "root"
		}

		fmt.Printf("  %s • %.2f%%\n", dirName, d.Percentage)
		bar := BarStyle().ViewAs(d.Percentage / 100.0)
		fmt.Println("  " + bar)
	}

	fmt.Println(SectionStyle().Render("🔖 Annotations"))
	fmt.Printf("  TODO: %s  FIXME: %s  HACK: %s  Total: %s\n",
		FormatIntBritishEnglish(report.AnnotationMetrics.TotalTODO),
		FormatIntBritishEnglish(report.AnnotationMetrics.TotalFIXME),
		FormatIntBritishEnglish(report.AnnotationMetrics.TotalHACK),
		FormatIntBritishEnglish(report.AnnotationMetrics.TotalAnnotations),
	)
}
