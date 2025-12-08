package pathfinder

import (
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// pass data from workers back to main goroutine
type scanResult struct {
	fileMetrics LanguageMetrics
	annMetrics  AnnotationMetrics
	path        string
	err         error
}

func scanCodebase(flags Config) (CodebaseReport, error) {
	startTime := time.Now()

	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	// channel to send jobs to workers (basically we are sending file paths and language definitions to workers)
	jobs := make(chan struct {
		path    string
		langDef *LanguageDefinition
	}, 100) // buffered channel to avoid sharing memory through mutexes
	results := make(chan scanResult, 100)

	// start worker pool
	var wgWorkers sync.WaitGroup
	numWorkers := runtime.NumCPU() * 20 // limit number of workers to avoid overwhelming the system

	for _ = range numWorkers {
		wgWorkers.Add(1)
		go func() {
			defer wgWorkers.Done()
			for job := range jobs {
				fMetrics, aMetrics, err := fileCounter(job.path, flags.BufferSizeFlag, job.langDef)
				results <- scanResult{
					fileMetrics: fMetrics,
					annMetrics:  aMetrics,
					path:        job.path,
					err:         err,
				}
			}
		}()
	}

	var wgConsumer sync.WaitGroup
	wgConsumer.Add(1)

	langStatsMap := map[string]*LanguageMetrics{}
	dirStatsMap := map[string]int{}

	var codebaseStats CodebaseMetrics
	var annotationStats AnnotationMetrics
	var dependencyStats DependencyMetrics
	topFilesList := make([]FileMetricsReport, 0)
	var walkErr error

	go func() {
		defer wgConsumer.Done()
		for res := range results {
			if res.err != nil {
				log.Fatal("Error processing file", res.path, ":", res.err)
			}

			codebaseStats.TotalFiles += res.fileMetrics.Files
			codebaseStats.TotalCode += res.fileMetrics.Code
			codebaseStats.TotalComments += res.fileMetrics.Comments
			codebaseStats.TotalBlanks += res.fileMetrics.Blanks

			annotationStats.TotalTODO += res.annMetrics.TotalTODO
			annotationStats.TotalFIXME += res.annMetrics.TotalFIXME
			annotationStats.TotalHACK += res.annMetrics.TotalHACK
			annotationStats.TotalAnnotations += res.annMetrics.TotalAnnotations

			relPath, _ := filepath.Rel(flags.PathFlag, res.path)
			topDir := topLevelDir(relPath)
			dirStatsMap[topDir] += res.fileMetrics.Lines

			stats := langStatsMap[res.fileMetrics.Language]
			if stats == nil {
				stats = &LanguageMetrics{Language: res.fileMetrics.Language}
				langStatsMap[res.fileMetrics.Language] = stats
			}
			stats.Files++
			stats.Code += res.fileMetrics.Code
			stats.Comments += res.fileMetrics.Comments
			stats.Blanks += res.fileMetrics.Blanks
			stats.Lines += res.fileMetrics.Lines

			topFilesList = append(topFilesList, FileMetricsReport{
				Metrics: res.fileMetrics,
				Path:    relPath,
			})
		}
	}()

	walkErr = filepath.WalkDir(flags.PathFlag, func(path string, d fs.DirEntry, err error) error {
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
		if excludeFile(name) {
			return nil
		}

		if !flags.HiddenFlag && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if excludeDir(name) {
				return filepath.SkipDir
			}
			codebaseStats.TotalDirs++
			return nil
		}

		if isBinary(name) {
			return nil
		}

		ext := hasNoExt(name)
		if ext == "" {
			return nil
		}

		langDefinition := determineLangByExt(ext)
		if langDefinition == nil {
			return nil
		}

		// send job to workers)
		jobs <- struct {
			path    string
			langDef *LanguageDefinition
		}{path, langDefinition}

		return nil
	})

	// cleanup and wait for final aggregation (consuming results)
	close(jobs)
	wgWorkers.Wait()
	close(results)
	wgConsumer.Wait()

	if walkErr != nil {
		return CodebaseReport{}, walkErr
	}

	// scan for dependencies if flag is enabled
	if flags.DependencyFlag {
		depFiles, err := scanDependencies(flags.PathFlag, flags)
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
