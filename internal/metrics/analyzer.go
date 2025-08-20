package metrics

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var excludedDirsFromScan = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
	"out":          {},
	"dist":         {},
	"build":        {},
	"target":       {},
	".idea":        {},
	".vscode":      {},
	".cache":       {},
}

func Analyze(flags Flags) (CodebaseReport, error) {
	if flags.PathFlag == "" { // won't ever happen since default is "." set by cobra
		return CodebaseReport{}, errors.New("path is required")
	}

	excludes := make(map[string]struct{}, len(excludedDirsFromScan))
	for key := range excludedDirsFromScan {
		excludes[key] = struct{}{}
	}

	langStatsMap := map[string]*LanguageMetrics{}
	dirStatsmMap := map[string]int{}
	var topFilesList []FileMetricsReport
	var codebaseStats CodebaseMetrics
	var annotationStats AnnotationMetrics

	err := filepath.WalkDir(flags.PathFlag, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()

		if !flags.HiddenFlag && strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if _, ok := excludes[name]; ok {
				return fs.SkipDir
			}
			codebaseStats.TotalDirs++
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

		code, comments, blanks, ann := CountFile(path, langDefinition.Type)
		if code == 0 && comments == 0 && blanks == 0 {
			return nil
		}

		annotationStats.TotalTODO += ann.TotalTODO
		annotationStats.TotalFIXME += ann.TotalFIXME
		annotationStats.TotalHACK += ann.TotalHACK
		annotationStats.TotalAnnotations += ann.TotalAnnotations

		codebaseStats.TotalFiles++
		codebaseStats.TotalCode += code
		codebaseStats.TotalComments += comments
		codebaseStats.TotalBlanks += blanks

		stats := langStatsMap[langDefinition.Name]
		if stats == nil {
			stats = &LanguageMetrics{}
			langStatsMap[langDefinition.Name] = stats
		}

		stats.Files++
		stats.Language = langDefinition.Name
		stats.Code += code
		stats.Comments += comments
		stats.Blanks += blanks
		stats.Lines += code + comments + blanks

		relativePath, _ := filepath.Rel(flags.PathFlag, path)
		topDir := TopLevelDir(relativePath)
		dirStatsmMap[topDir] += code + comments + blanks

		fileLangMetrics := LanguageMetrics{
			Language: langDefinition.Name,
			Code:     code,
			Comments: comments,
			Blanks:   blanks,
			Lines:    code + comments + blanks,
		}

		topFilesList = append(topFilesList, FileMetricsReport{
			Metrics: fileLangMetrics,
			Path:    relativePath,
		})

		return nil
	})

	if err != nil {
		return CodebaseReport{}, err
	}

	languageStats := make([]LanguageMetricsReport, 0, len(langStatsMap))
	for lang, stats := range langStatsMap {
		langMetrics := LanguageMetrics{
			Language: lang,
			Files:    stats.Files,
			Code:     stats.Code,
			Comments: stats.Comments,
			Blanks:   stats.Blanks,
			Lines:    stats.Lines,
		}

		languageStats = append(languageStats, LanguageMetricsReport{
			// TODO: calculate percentage based on total lines
			Metrics: langMetrics,
		})
	}

	sort.Slice(topFilesList, func(i, j int) bool {
		return topFilesList[i].Metrics.Lines > topFilesList[j].Metrics.Lines
	})

	dirs := make([]DirMetricsReport, 0, len(dirStatsmMap))
	for dir, stat := range dirStatsmMap {
		dirs = append(dirs, DirMetricsReport{
			Directory: dir,
			Lines:     stat,
		})
	}

	codebaseStats.TotalLanguages = len(languageStats)
	codebaseStats.TotalLines = codebaseStats.TotalCode + codebaseStats.TotalComments + codebaseStats.TotalBlanks

	return CodebaseReport{
		LanguageMetrics:   languageStats,
		FileMetrics:       topFilesList,
		DirMetrics:        dirs,
		CodebaseMetrics:   codebaseStats,
		AnnotationMetrics: annotationStats,
	}, nil
}
