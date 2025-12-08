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
)

// pass data from workers back to main goroutine
type scanResult struct {
	fileMetrics LanguageMetrics
	annMetrics  AnnotationMetrics
	path        string
	err         error
}

func scanCodebase(flags Config) (CodebaseReport, error) {
	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	// limit number of workers to avoid overwhelming the system (TODO: will make a flag later)
	numWorkers := runtime.NumCPU() * 20

	// channel to send jobs to workers (basically we are sending file paths and language definitions to workers)
	locJobs := make(chan struct {
		path    string
		langDef *LanguageDefinition
	}, 100) // buffered channel to avoid sharing memory through mutexes
	locResults := make(chan scanResult, 100)

	// start worker pool for counting lines of code
	var wgLocWorkers sync.WaitGroup

	// start loc workers (separate goroutines that process files concurrently by the go scheduler)
	// and send loc worker results back to the loc results channel which is consumed by the main goroutine
	for _ = range numWorkers {
		wgLocWorkers.Add(1)
		go func() {
			defer wgLocWorkers.Done()
			for job := range locJobs {
				fMetrics, aMetrics, err := fileCounter(job.path, flags.BufferSizeFlag, job.langDef)
				locResults <- scanResult{
					fileMetrics: fMetrics,
					annMetrics:  aMetrics,
					path:        job.path,
					err:         err,
				}
			}
		}()
	}

	// channel to send dependency scan jobs and receive results
	depJobs := make(chan DependencyFile, 100)
	depResults := make(chan DependencyFile, 100)

	// start worker pool for scanning dependencies if flag is enabled
	var wgDepWorkers sync.WaitGroup

	if flags.DependencyFlag {
		for _ = range numWorkers {
			wgDepWorkers.Add(1)
			go func() {
				defer wgDepWorkers.Done()
				for job := range depJobs {
					var deps []string
					var err error

					switch {
					case strings.HasSuffix(job.Path, "go.mod"):
						deps, err = scanGoMod(job.Path)
					case strings.HasSuffix(job.Path, "package.json"):
						deps, err = scanPackageJSON(job.Path)
					case strings.HasSuffix(job.Path, "requirements.txt"):
						deps, err = scanRequirementsTxt(job.Path)
					case strings.HasSuffix(job.Path, "pom.xml"):
						deps, err = scanPomXML(job.Path)
					case strings.HasSuffix(job.Path, ".csproj"):
						deps, err = scanCsproj(job.Path)
					}

					if err == nil && len(deps) > 0 {
						job.Dependencies = deps
						depResults <- job
					}
				}
			}()
		}
	}

	// structs and data structures to hold aggregated results (readable data for user)
	langStatsMap := map[string]*LanguageMetrics{}
	dirStatsMap := map[string]int{}

	var codebaseStats CodebaseMetrics
	var annotationStats AnnotationMetrics
	var dependencyStats DependencyMetrics
	topFilesList := make([]FileMetricsReport, 0)
	var walkErr error

	// start consumer goroutine to aggregate results channel data (from workers)
	var wgFinalResults sync.WaitGroup
	wgFinalResults.Add(1)

	// the actual logic for aggregating results from workers (this is what the main goroutine waits for)
	// this doesn't run until we start sending jobs to the workers below, and continues until the results channel is closed
	go func() {
		defer wgFinalResults.Done()
		for res := range locResults {
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

	if flags.DependencyFlag {
		wgFinalResults.Add(1)
		go func() {
			defer wgFinalResults.Done()
			for res := range depResults {
				dependencyStats.DependencyFiles = append(dependencyStats.DependencyFiles, res)
				dependencyStats.TotalDependencies += len(res.Dependencies)
			}
		}()
	}

	// this traverses the directory tree and sends jobs to the workers (via the jobs channel)
	// this actually runs in the main goroutine
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

		// send job to loc workers to process file if we can determine its language by extension
		ext := hasNoExt(name)
		if ext != "" {
			if langDefinition := determineLangByExt(ext); langDefinition != nil {
				locJobs <- struct {
					path    string
					langDef *LanguageDefinition
				}{path, langDefinition}
			}
		}

		// send job to dependency scanner workers if file is a known dependency file
		if flags.DependencyFlag {
			var depType string
			switch {
			case name == "go.mod":
				depType = "Go Modules"
			case name == "package.json":
				depType = "npm/yarn"
			case name == "requirements.txt":
				depType = "pip"
			case name == "pom.xml":
				depType = "Maven"
			case strings.HasSuffix(name, ".csproj"):
				depType = ".NET/NuGet"
			}

			if depType != "" {
				depJobs <- DependencyFile{
					Path: path,
					Type: depType,
				}
			}
		}

		return nil
	})

	// cleanup and wait for final aggregation (consuming results)
	close(locJobs)
	if flags.DependencyFlag {
		close(depJobs)
	}

	wgLocWorkers.Wait()
	wgDepWorkers.Wait()

	close(locResults)
	if flags.DependencyFlag {
		close(depResults)
	}

	wgFinalResults.Wait()

	if walkErr != nil {
		return CodebaseReport{}, walkErr
	}

	// the final aggregation and more calculations are done below

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

	return CodebaseReport{
		LanguageMetrics:   languageStats,
		FileMetrics:       topFilesList,
		DirMetrics:        dirStats,
		CodebaseMetrics:   codebaseStats,
		AnnotationMetrics: annotationStats,
		DependencyMetrics: dependencyStats,
	}, nil
}
