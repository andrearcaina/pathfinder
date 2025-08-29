package metrics

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var (
	wg sync.WaitGroup
	mu sync.Mutex
)

func Analyze(flags Flags) (CodebaseReport, error) {
	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	langStatsMap := map[string]*LanguageMetrics{}
	dirStatsMap := map[string]int{}

	var codebaseStats CodebaseMetrics
	var annotationStats AnnotationMetrics

	topFilesList := make([]FileMetricsReport, 0)

	err := filepath.WalkDir(flags.PathFlag, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// if the user didn't set recursive flag, skip subdirectories
		if !flags.RecursiveFlag && path != flags.PathFlag && d.IsDir() {
			return fs.SkipDir
		}

		// if the user set max depth, skip directories deeper than max depth
		if flags.RecursiveFlag && flags.MaxDepthFlag != -1 {
			relPath, _ := filepath.Rel(flags.PathFlag, path)
			depth := len(strings.Split(relPath, string(filepath.Separator)))
			if depth > flags.MaxDepthFlag {
				if d.IsDir() {
					return fs.SkipDir
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
				return fs.SkipDir
			}

			return nil
		}

		if d.IsDir() {
			if ExcludeDir(name) {
				return fs.SkipDir
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

			code, comments, blanks, ann := CountFile(filePath, bufferSize, langDef.Type)
			if code == 0 && comments == 0 && blanks == 0 {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			annotationStats.TotalTODO += ann.TotalTODO
			annotationStats.TotalFIXME += ann.TotalFIXME
			annotationStats.TotalHACK += ann.TotalHACK
			annotationStats.TotalAnnotations += ann.TotalAnnotations

			codebaseStats.TotalFiles++
			codebaseStats.TotalCode += code
			codebaseStats.TotalComments += comments
			codebaseStats.TotalBlanks += blanks

			stats := langStatsMap[langDef.Name]
			if stats == nil {
				stats = &LanguageMetrics{}
				langStatsMap[langDef.Name] = stats
			}

			stats.Files++
			stats.Language = langDef.Name
			stats.Code += code
			stats.Comments += comments
			stats.Blanks += blanks
			stats.Lines += code + comments + blanks

			relativePath, _ := filepath.Rel(flags.PathFlag, filePath)
			topDir := TopLevelDir(relativePath)
			dirStatsMap[topDir] += code + comments + blanks

			fileLangMetrics := LanguageMetrics{
				Language: langDef.Name,
				Code:     code,
				Comments: comments,
				Blanks:   blanks,
				Lines:    code + comments + blanks,
			}

			topFilesList = append(topFilesList, FileMetricsReport{
				Metrics: fileLangMetrics,
				Path:    relativePath,
			})

			return
		}(path, langDefinition, flags.BufferSizeFlag)

		return nil
	})

	if err != nil {
		return CodebaseReport{}, err
	}

	wg.Wait()

	codebaseStats.TotalLines = codebaseStats.TotalCode + codebaseStats.TotalComments + codebaseStats.TotalBlanks

	languageStats := make([]LanguageMetricsReport, 0, len(langStatsMap))
	for _, stats := range langStatsMap {
		languageStats = append(languageStats, LanguageMetricsReport{
			Percentage: (float64(stats.Code) / float64(codebaseStats.TotalLines)) * 100,
			Metrics:    *stats,
		})
	}

	dirStats := make([]DirMetricsReport, 0, len(dirStatsMap))
	for dir, stat := range dirStatsMap {
		dirStats = append(dirStats, DirMetricsReport{
			Directory:  dir,
			Percentage: (float64(stat) / float64(codebaseStats.TotalLines)) * 100,
			Lines:      stat,
		})
	}

	codebaseStats.TotalLanguages = len(languageStats)

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
	}, nil
}
