package metrics

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func ScanCodebase(flags Flags) (CodebaseReport, error) {
	startTime := time.Now()

	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	langStatsMap := map[string]*LanguageMetrics{}
	dirStatsMap := map[string]int{}

	var codebaseStats CodebaseMetrics
	var annotationStats AnnotationMetrics
	var dependencyStats DependencyMetrics
	topFilesList := make([]FileMetricsReport, 0)

	// local variables for synchronization
	var wg sync.WaitGroup
	var mu sync.Mutex

	err := filepath.WalkDir(flags.PathFlag, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// if the user didn't set recursive flag, skip subdirectories
		if !flags.RecursiveFlag && path != flags.PathFlag && d.IsDir() {
			return filepath.SkipDir
		}

		// if the user set max depth, skip directories deeper than max depth
		if flags.RecursiveFlag && flags.MaxDepthFlag != -1 {
			relPath, _ := filepath.Rel(flags.PathFlag, path)
			depth := len(strings.Split(relPath, string(filepath.Separator)))
			if depth > flags.MaxDepthFlag {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		name := d.Name()

		// skips files like .DS_Store, package-lock.json, etc.
		if ExcludeFile(name) {
			return nil
		}

		if !flags.HiddenFlag && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if ExcludeDir(name) {
				return filepath.SkipDir
			}

			mu.Lock()
			codebaseStats.TotalDirs++
			mu.Unlock()
			return nil
		}

		if IsBinary(name) {
			return nil
		}

		ext := HasNoExt(name)
		if ext == "" {
			return nil
		}

		langDefinition := DetermineLangByExt(ext)
		if langDefinition == nil {
			return nil
		}

		wg.Add(1)
		go func(filePath string, langDef *LanguageDefinition, bufferSize int) {
			defer wg.Done()

			fileMetrics, annMetrics, err := FileCounter(filePath, bufferSize, langDef)
			if err != nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			// aggregate codebase stats (this is generic stats for the entire codebase)
			// the difference between codebaseStats and langStatsMap is that codebaseStats is for the entire codebase
			// while langStatsMap is per language
			// so codebaseStats.TotalFiles is the total number of files in the entire code
			// while langStatsMap[fileMetrics.Language].Files is the total number of files for that specific language
			codebaseStats.TotalFiles += fileMetrics.Files
			codebaseStats.TotalCode += fileMetrics.Code
			codebaseStats.TotalComments += fileMetrics.Comments
			codebaseStats.TotalBlanks += fileMetrics.Blanks

			// aggregate annotation metrics (this is for the entire codebase)
			annotationStats.TotalTODO += annMetrics.TotalTODO
			annotationStats.TotalFIXME += annMetrics.TotalFIXME
			annotationStats.TotalHACK += annMetrics.TotalHACK
			annotationStats.TotalAnnotations += annMetrics.TotalAnnotations

			// aggregate directory stats (this is for the entire codebase)
			relativePath, _ := filepath.Rel(flags.PathFlag, filePath)
			topDir := TopLevelDir(relativePath)
			dirStatsMap[topDir] += fileMetrics.Lines

			// this key sorta matters because we want to group by the language being used, but it won't be used in the final report
			stats := langStatsMap[fileMetrics.Language]
			if stats == nil {
				stats = &LanguageMetrics{}
				langStatsMap[fileMetrics.Language] = stats
			}

			stats.Language = fileMetrics.Language // we write fileMetrics.Language here because this is what's actually being used (not the map key)
			stats.Files += fileMetrics.Files      // always being incremented by 1
			stats.Code += fileMetrics.Code
			stats.Comments += fileMetrics.Comments
			stats.Blanks += fileMetrics.Blanks
			stats.Lines += fileMetrics.Lines

			// append the file metrics to the top files list
			// we'll sort and pick the top 10 later in the final report
			topFilesList = append(topFilesList, FileMetricsReport{
				Metrics: fileMetrics,
				Path:    relativePath,
			})
		}(path, langDefinition, flags.BufferSizeFlag)

		return nil
	})

	if err != nil {
		return CodebaseReport{}, err
	}

	wg.Wait()

	// scan for dependencies if flag is enabled
	if flags.DependencyFlag {
		depFiles, err := ScanDependencies(flags.PathFlag, flags)
		if err == nil {
			dependencyStats.DependencyFiles = depFiles
			// count total dependencies across all files
			for _, depFile := range depFiles {
				dependencyStats.TotalDependencies += len(depFile.Dependencies)
			}
		}
	}

	// this is done after the walk is complete and all goroutines have finished
	// because we need the final aggregated stats to calculate percentages
	// it doesn't make sense to calculate percentages while we're still walking the directory
	codebaseStats.TotalLines = codebaseStats.TotalCode + codebaseStats.TotalComments + codebaseStats.TotalBlanks

	// prepare language stats for the final report
	languageStats := make([]LanguageMetricsReport, 0, len(langStatsMap))
	for _, stats := range langStatsMap {
		languageStats = append(languageStats, LanguageMetricsReport{
			Percentage: (float64(stats.Code) / float64(codebaseStats.TotalLines)) * 100,
			Metrics:    *stats,
		})
	}

	// prepare directory stats for the final report
	dirStats := make([]DirMetricsReport, 0, len(dirStatsMap))
	for dir, stat := range dirStatsMap {
		dirStats = append(dirStats, DirMetricsReport{
			Directory:  dir,
			Percentage: (float64(stat) / float64(codebaseStats.TotalLines)) * 100,
			Lines:      stat,
		})
	}

	// total languages is simply the length of the language stats map
	codebaseStats.TotalLanguages = len(languageStats)

	// sort the lists for better presentation in the final report
	sort.Slice(topFilesList, func(i, j int) bool {
		return topFilesList[i].Metrics.Lines > topFilesList[j].Metrics.Lines
	})

	sort.Slice(languageStats, func(i, j int) bool {
		return languageStats[i].Percentage > languageStats[j].Percentage
	})

	sort.Slice(dirStats, func(i, j int) bool {
		return dirStats[i].Percentage > dirStats[j].Percentage
	})

	// honestly not actually the best way to benchmark performance (linux time is better)
	elapsedTime := time.Since(startTime)

	performanceStats := PerformanceMetrics{
		ElapsedTime: formatDuration(elapsedTime),
	}

	return CodebaseReport{
		LanguageMetrics:    languageStats,
		FileMetrics:        topFilesList,
		DirMetrics:         dirStats,
		CodebaseMetrics:    codebaseStats,
		AnnotationMetrics:  annotationStats,
		DependencyMetrics:  dependencyStats,
		PerformanceMetrics: performanceStats,
	}, nil
}
