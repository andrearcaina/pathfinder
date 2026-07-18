package pathfinder

import (
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type scanJob struct {
	path    string
	langDef *LanguageDefinition
}

// pass data from workers back to main goroutine
type scanResult struct {
	fileMetrics LanguageMetrics
	annMetrics  AnnotationMetrics
	path        string
	err         error
}

type scanAggregation struct {
	langStatsMap    map[string]*LanguageMetrics
	dirStatsMap     map[string]int
	codebaseStats   CodebaseMetrics
	annotationStats AnnotationMetrics
	dependencyStats DependencyMetrics
	topFilesList    []FileMetricsReport
}

func scanCodebase(flags Config) (CodebaseReport, error) {
	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	startTime := time.Now()
	locJobs := make(chan scanJob, 100)
	locResults := make(chan scanResult, 100)
	depJobs := make(chan DependencyFile, 100)
	depResults := make(chan DependencyFile, 100)

	workers, waitForLocWorkers := startScanWorkers(flags, locJobs, locResults)
	waitForDepWorkers := startDependencyWorkers(flags, depJobs, depResults)

	aggregation := newScanAggregation()
	waitForResults := startResultConsumers(flags, locResults, depResults, aggregation)

	totalDirs, walkErr := walkCodebase(flags, locJobs, depJobs)
	close(locJobs)
	if flags.DependencyFlag {
		close(depJobs)
	}

	waitForLocWorkers()
	waitForDepWorkers()
	close(locResults)
	if flags.DependencyFlag {
		close(depResults)
	}
	waitForResults()

	if walkErr != nil {
		return CodebaseReport{}, walkErr
	}

	aggregation.codebaseStats.TotalDirs = totalDirs
	return buildCodebaseReport(flags, startTime, workers, aggregation), nil
}

func startScanWorkers(flags Config, jobs <-chan scanJob, results chan<- scanResult) ([]*WorkerStats, func()) {
	var wg sync.WaitGroup
	workers := make([]*WorkerStats, flags.WorkerFlag)

	for i := range flags.WorkerFlag {
		wg.Add(1)
		workers[i] = &WorkerStats{Id: i, Start: time.Now()}

		go func(ws *WorkerStats) {
			defer wg.Done()

			for job := range jobs {
				fileMetrics, annotationMetrics, err := fileCounter(job.path, flags.BufferSizeFlag, job.langDef)
				ws.Processed++
				results <- scanResult{
					fileMetrics: fileMetrics,
					annMetrics:  annotationMetrics,
					path:        job.path,
					err:         err,
				}
			}

			ws.End = time.Now()
			ws.Duration = ws.End.Sub(ws.Start).Seconds()
			ws.Throughput = float64(ws.Processed) / ws.Duration
		}(workers[i])
	}

	return workers, wg.Wait
}

func startDependencyWorkers(flags Config, jobs <-chan DependencyFile, results chan<- DependencyFile) func() {
	var wg sync.WaitGroup
	if !flags.DependencyFlag {
		return wg.Wait
	}

	for range flags.WorkerFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				dependencies, err := scanDependencyFile(job)
				if err == nil && len(dependencies) > 0 {
					job.Dependencies = dependencies
					results <- job
				}
			}
		}()
	}

	return wg.Wait
}

func scanDependencyFile(file DependencyFile) ([]string, error) {
	switch {
	case strings.HasSuffix(file.Path, "go.mod"):
		return scanGoMod(file.Path)
	case strings.HasSuffix(file.Path, "package.json"):
		return scanPackageJSON(file.Path)
	case strings.HasSuffix(file.Path, "requirements.txt"):
		return scanRequirementsTxt(file.Path)
	case strings.HasSuffix(file.Path, "pom.xml"):
		return scanPomXML(file.Path)
	case strings.HasSuffix(file.Path, ".csproj"):
		return scanCsproj(file.Path)
	default:
		return nil, nil
	}
}

func newScanAggregation() *scanAggregation {
	return &scanAggregation{
		langStatsMap: map[string]*LanguageMetrics{},
		dirStatsMap:  map[string]int{},
		topFilesList: make([]FileMetricsReport, 0),
	}
}

func startResultConsumers(flags Config, locResults <-chan scanResult, depResults <-chan DependencyFile, aggregation *scanAggregation) func() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range locResults {
			aggregateScanResult(flags.PathFlag, result, aggregation)
		}
	}()

	if flags.DependencyFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for result := range depResults {
				aggregation.dependencyStats.DependencyFiles = append(aggregation.dependencyStats.DependencyFiles, result)
				aggregation.dependencyStats.TotalDependencies += len(result.Dependencies)
			}
		}()
	}

	return wg.Wait
}

func aggregateScanResult(rootPath string, result scanResult, aggregation *scanAggregation) {
	if result.err != nil {
		log.Fatal("Error processing file", result.path, ":", result.err)
	}

	aggregation.codebaseStats.TotalFiles += result.fileMetrics.Files
	aggregation.codebaseStats.TotalCode += result.fileMetrics.Code
	aggregation.codebaseStats.TotalComments += result.fileMetrics.Comments
	aggregation.codebaseStats.TotalBlanks += result.fileMetrics.Blanks

	aggregation.annotationStats.TotalTODO += result.annMetrics.TotalTODO
	aggregation.annotationStats.TotalFIXME += result.annMetrics.TotalFIXME
	aggregation.annotationStats.TotalHACK += result.annMetrics.TotalHACK
	aggregation.annotationStats.TotalAnnotations += result.annMetrics.TotalAnnotations

	relPath, _ := filepath.Rel(rootPath, result.path)
	aggregation.dirStatsMap[topLevelDir(relPath)] += result.fileMetrics.Lines

	stats := aggregation.langStatsMap[result.fileMetrics.Language]
	if stats == nil {
		stats = &LanguageMetrics{Language: result.fileMetrics.Language}
		aggregation.langStatsMap[result.fileMetrics.Language] = stats
	}
	stats.Files++
	stats.Code += result.fileMetrics.Code
	stats.Comments += result.fileMetrics.Comments
	stats.Blanks += result.fileMetrics.Blanks
	stats.Lines += result.fileMetrics.Lines

	aggregation.topFilesList = append(aggregation.topFilesList, FileMetricsReport{
		Metrics: result.fileMetrics,
		Path:    relPath,
	})
}

func walkCodebase(flags Config, locJobs chan<- scanJob, depJobs chan<- DependencyFile) (int, error) {
	totalDirs := 0
	err := filepath.WalkDir(flags.PathFlag, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		if shouldSkipForDepth(flags, path, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		name := entry.Name()
		if excludeFile(name) {
			return nil
		}
		if !flags.HiddenFlag && strings.HasPrefix(name, ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if excludeDir(name) {
				return filepath.SkipDir
			}
			totalDirs++
			return nil
		}
		if isBinary(name) {
			return nil
		}

		queueScanJob(path, name, locJobs)
		if flags.DependencyFlag {
			queueDependencyJob(path, name, depJobs)
		}
		return nil
	})

	return totalDirs, err
}

func shouldSkipForDepth(flags Config, path string, entry fs.DirEntry) bool {
	if !flags.RecursiveFlag {
		return path != flags.PathFlag && entry.IsDir()
	}
	if flags.MaxDepthFlag == -1 {
		return false
	}

	relPath, _ := filepath.Rel(flags.PathFlag, path)
	depth := len(strings.Split(relPath, string(filepath.Separator)))
	return depth > flags.MaxDepthFlag
}

func queueScanJob(path, name string, jobs chan<- scanJob) {
	ext := hasNoExt(name)
	if ext == "" {
		return
	}
	if langDefinition := determineLangByExt(ext); langDefinition != nil {
		jobs <- scanJob{path: path, langDef: langDefinition}
	}
}

func queueDependencyJob(path, name string, jobs chan<- DependencyFile) {
	var dependencyType string
	switch {
	case name == "go.mod":
		dependencyType = "Go Modules"
	case name == "package.json":
		dependencyType = "npm/yarn"
	case name == "requirements.txt":
		dependencyType = "pip"
	case name == "pom.xml":
		dependencyType = "Maven"
	case strings.HasSuffix(name, ".csproj"):
		dependencyType = ".NET/NuGet"
	}

	if dependencyType != "" {
		jobs <- DependencyFile{Path: path, Type: dependencyType}
	}
}

func buildCodebaseReport(flags Config, startTime time.Time, workers []*WorkerStats, aggregation *scanAggregation) CodebaseReport {
	aggregation.codebaseStats.TotalLines = aggregation.codebaseStats.TotalCode + aggregation.codebaseStats.TotalComments + aggregation.codebaseStats.TotalBlanks
	languageStats := buildLanguageStats(aggregation.langStatsMap, aggregation.codebaseStats.TotalLines)
	dirStats := buildDirectoryStats(aggregation.dirStatsMap, aggregation.codebaseStats.TotalLines)
	aggregation.codebaseStats.TotalLanguages = len(languageStats)

	sort.Slice(aggregation.topFilesList, func(i, j int) bool {
		return aggregation.topFilesList[i].Metrics.Lines > aggregation.topFilesList[j].Metrics.Lines
	})
	sort.Slice(languageStats, func(i, j int) bool {
		return languageStats[i].Percentage > languageStats[j].Percentage
	})
	sort.Slice(dirStats, func(i, j int) bool {
		return dirStats[i].Percentage > dirStats[j].Percentage
	})

	report := CodebaseReport{
		LanguageMetrics:   languageStats,
		FileMetrics:       aggregation.topFilesList,
		DirMetrics:        dirStats,
		CodebaseMetrics:   aggregation.codebaseStats,
		AnnotationMetrics: aggregation.annotationStats,
		DependencyMetrics: aggregation.dependencyStats,
	}
	if flags.ThroughputFlag {
		totalTime := time.Since(startTime).Seconds()
		report.PerformanceMetrics = PerformanceMetrics{
			TotalWorkers:      flags.WorkerFlag,
			WorkerStats:       workers,
			TotalTimeSeconds:  totalTime,
			OverallThroughput: float64(aggregation.codebaseStats.TotalFiles) / totalTime,
		}
	}

	return report
}

func buildLanguageStats(statsMap map[string]*LanguageMetrics, totalLines int) []LanguageMetricsReport {
	stats := make([]LanguageMetricsReport, 0, len(statsMap))
	for _, metrics := range statsMap {
		stats = append(stats, LanguageMetricsReport{
			Percentage: (float64(metrics.Code) / float64(totalLines)) * 100,
			Metrics:    *metrics,
		})
	}
	return stats
}

func buildDirectoryStats(statsMap map[string]int, totalLines int) []DirMetricsReport {
	stats := make([]DirMetricsReport, 0, len(statsMap))
	for directory, lines := range statsMap {
		stats = append(stats, DirMetricsReport{
			Directory:  directory,
			Percentage: (float64(lines) / float64(totalLines)) * 100,
			Lines:      lines,
		})
	}
	return stats
}
