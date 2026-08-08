package ui

import (
	"fmt"
	"strings"

	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
	"github.com/charmbracelet/lipgloss"
)

const maxBarWidth = 40

func PrintReport(report pathfinder.CodebaseReport, throughputMode bool) {
	if report.CodebaseMetrics.TotalFiles == 0 {
		fmt.Println("No files analyzed. Please check the path and try again.")
		return // exit program
	}

	if throughputMode { // display what really matters (the performance)
		fmt.Println(renderThroughputReport(report))
		return
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
	// only show top 10 files
	for i := 0; i < len(report.FileMetrics) && i < 10; i++ {
		f := report.FileMetrics[i]

		ratio := float64(f.Metrics.Lines) / float64(maxLines)
		bar := BarStyle().ViewAs(ratio)

		fmt.Printf("  %s • %s lines\n", f.Path, FormatIntBritishEnglish(f.Metrics.Lines))
		fmt.Println("  " + bar)
	}

	// TODO: handle a flag to show all dirs (not recommended for large codebases)
	// only show top 10 directories
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

	// display dependency metrics if available
	if len(report.DependencyMetrics.DependencyFiles) > 0 {
		fmt.Println(SectionStyle().Render("📦 Dependencies"))

		totalDepsText := fmt.Sprintf("Total Dependencies: %s", FormatIntBritishEnglish(report.DependencyMetrics.TotalDependencies))
		fmt.Println("  " + BadgeStyle().Render(totalDepsText))

		// group dependency files by type
		depByType := make(map[string][]pathfinder.DependencyFile)
		for _, depFile := range report.DependencyMetrics.DependencyFiles {
			depByType[depFile.Type] = append(depByType[depFile.Type], depFile)
		}

		// display each dependency type with styling
		for depType, files := range depByType {
			totalDepsForType := 0
			for _, file := range files {
				totalDepsForType += len(file.Dependencies)
			}

			// style the dependency type header
			typeHeader := fmt.Sprintf("%s: %s dependencies (%d files)",
				depType,
				FormatIntBritishEnglish(totalDepsForType),
				len(files),
			)

			depTypeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true).
				MarginLeft(2)

			fmt.Println(depTypeStyle.Render(typeHeader))

			// style individual dependency files
			fileStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B0B0B0")).
				MarginLeft(4)

			// show dependency files (limit to avoid clutter)
			for i, file := range files {
				if i >= 3 { // show max 3 files per type
					if len(files) > 3 {
						moreFilesText := fmt.Sprintf("... and %d more files", len(files)-3)
						moreStyle := lipgloss.NewStyle().
							Foreground(lipgloss.Color("#808080")).
							Italic(true).
							MarginLeft(4)
						fmt.Println(moreStyle.Render(moreFilesText))
					}
					break
				}

				fileText := fmt.Sprintf("%s (%d deps)", file.Path, len(file.Dependencies))
				fmt.Println(fileStyle.Render(fileText))
			}
		}
	}
}

func renderThroughputReport(report pathfinder.CodebaseReport) string {
	metrics := report.PerformanceMetrics

	workerHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Bold(true)
	throughputTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50C878")).
		Bold(true)
	workerIDStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ACD32")).
		Bold(true)
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#55565B")).
		Padding(0, 1)

	workerLines := []string{
		workerHeaderStyle.Render(fmt.Sprintf("  %-3s %12s %12s %10s", "ID", "FILES", "RATE", "ACTIVE")),
	}
	for _, worker := range metrics.WorkerStats {
		workerLines = append(workerLines, fmt.Sprintf("  %s %12s %12s %10s",
			workerIDStyle.Render(fmt.Sprintf("%-3d", worker.Id)),
			FormatIntBritishEnglish(worker.Processed),
			fmt.Sprintf("%.1f/s", worker.Throughput),
			fmt.Sprintf("%.2fs", worker.Duration),
		))
	}

	scanPanel := renderMetricPanel("SCAN RESULTS", [][2]string{
		{"Files", FormatIntBritishEnglish(report.CodebaseMetrics.TotalFiles)},
		{"Directories", FormatIntBritishEnglish(report.CodebaseMetrics.TotalDirs)},
		{"Lines", FormatIntBritishEnglish(report.CodebaseMetrics.TotalLines)},
		{"Duration", fmt.Sprintf("%.2fs", metrics.TotalTimeSeconds)},
		{"Overall rate", fmt.Sprintf("%.1f/s", metrics.OverallThroughput)},
	})

	runtimePanel := renderMetricPanel("RUNTIME", [][2]string{
		{"File workers", fmt.Sprint(metrics.FileWorkers)},
		{"Dependency workers", fmt.Sprint(metrics.DependencyWorkers)},
		{"Total workers", fmt.Sprint(metrics.TotalWorkers)},
		{"Result consumers", fmt.Sprint(metrics.ResultConsumers)},
		{"Pipeline goroutines", fmt.Sprint(metrics.PipelineGoroutines)},
		{"Logical CPUs", fmt.Sprint(metrics.LogicalCPUs)},
		{"GOMAXPROCS", fmt.Sprint(metrics.GOMAXPROCS)},
		{"OS threads created*", fmt.Sprint(metrics.OSThreadsCreated)},
	})

	return strings.Join([]string{
		throughputTitleStyle.Render("⚡ Pathfinder • Throughput"),
		"",
		workerHeaderStyle.Render("File-worker performance"),
		panelStyle.Render(strings.Join(workerLines, "\n")),
		"",
		workerHeaderStyle.Render("Summary"),
		lipgloss.JoinHorizontal(lipgloss.Top, scanPanel, "  ", runtimePanel),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Render("* OS threads created over the process lifetime"),
	}, "\n")
}

func renderMetricPanel(title string, metrics [][2]string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Width(20)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ACD32")).
		Bold(true)

	lines := []string{titleStyle.Render(title), ""}
	for _, metric := range metrics {
		lines = append(lines, labelStyle.Render(metric[0])+valueStyle.Render(metric[1]))
	}

	return lipgloss.NewStyle().
		Width(34).
		Height(10).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#50C878")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}
