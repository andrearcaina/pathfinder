package pathfinder

import (
	"errors"
	"path/filepath"
)

// Version returns the current version of the library.
func Version() string {
	return "v0.3.3"
}

// Scan is the main entry point for the library.
// It takes a Config and returns a detailed CodebaseReport.
func Scan(config Config) (CodebaseReport, error) {
	// set defaults if zero-values are present
	if config.PathFlag == "" {
		config.PathFlag = "."
	}
	if config.BufferSizeFlag == 0 {
		config.BufferSizeFlag = 4 // default to 4KB
	}
	if config.MaxDepthFlag == 0 {
		config.MaxDepthFlag = -1 // default to no limit
	}
	if config.WorkerFlag == 0 {
		config.WorkerFlag = 16 // default to 16 concurrent workers
	}

	// validation
	if !config.RecursiveFlag && config.MaxDepthFlag != -1 {
		return CodebaseReport{}, errors.New("--max-depth flag is ignored when --recursive is false")
	}

	absPath, err := filepath.Abs(config.PathFlag)
	if err != nil {
		return CodebaseReport{}, err
	}

	// switch and validate buffer size
	switch config.BufferSizeFlag {
	case 4, 8, 16, 32, 64:
		// valid so do nothing
	default:
		return CodebaseReport{}, errors.New("invalid Buffer Size. Allowed values are 4, 8, 16, 32, 64 (in KB)")
	}

	// prepare internal config (safe modification since we passed by value)
	config.PathFlag = absPath
	config.BufferSizeFlag = config.BufferSizeFlag * 1024

	return scanCodebase(config)
}

// ScannedLanguages returns a list of all programming languages found in the scanned codebase.
func (c CodebaseReport) ScannedLanguages() []string {
	languages := make([]string, 0, len(c.LanguageMetrics))

	for _, langReport := range c.LanguageMetrics {
		languages = append(languages, langReport.Metrics.Language)
	}

	return languages
}

// ScannedDirectories returns a list of all directories that were scanned in the codebase.
func (c CodebaseReport) ScannedDirectories() []string {
	directories := make([]string, 0, len(c.DirMetrics))

	for _, dirReport := range c.DirMetrics {
		directories = append(directories, dirReport.Directory)
	}

	return directories
}

// ScannedFiles returns a list of all files that were scanned in the codebase.
func (c CodebaseReport) ScannedFiles() []string {
	files := make([]string, 0, len(c.FileMetrics))

	for _, fileReport := range c.FileMetrics {
		files = append(files, fileReport.Path)
	}

	return files
}

// SupportedLanguages returns a list of all supported programming languages that Pathfinder can analyze.
func SupportedLanguages() []string {
	languages := make([]string, 0, len(languageDefinitions))

	for _, langDef := range languageDefinitions {
		languages = append(languages, langDef.Name)
	}
	return languages
}
