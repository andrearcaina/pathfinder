package pathfinder

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Version returns the current version of the library.
func Version() string {
	return "v0.1.3"
}

// Scan is the main entry point for the library.
// It takes a Config and returns a detailed CodebaseReport.
func Scan(config *Config) (CodebaseReport, error) {
	if config == nil {
		// create a default config if none is provided
		config = &Config{
			PathFlag:       ".",
			HiddenFlag:     false,
			BufferSizeFlag: 4,
			RecursiveFlag:  false,
			MaxDepthFlag:   -1,
			DependencyFlag: true,
			GitFlag:        false,
		}
		fmt.Println("Using default configuration.")
	}

	if !config.RecursiveFlag && config.MaxDepthFlag != -1 {
		return CodebaseReport{}, errors.New("--max-depth flag is ignored when --recursive is false")
	}

	pathFlag, err := filepath.Abs(config.PathFlag)
	if err != nil {
		return CodebaseReport{}, err
	}

	if config.BufferSizeFlag != 4 && config.BufferSizeFlag != 8 && config.BufferSizeFlag != 16 && config.BufferSizeFlag != 32 && config.BufferSizeFlag != 64 {
		return CodebaseReport{}, errors.New("invalid Buffer Size. Allowed values are 4, 8, 16, 32, 64 (in KB)")
	}

	config = &Config{
		PathFlag:       pathFlag,
		HiddenFlag:     config.HiddenFlag,
		BufferSizeFlag: config.BufferSizeFlag * 1024, // convert KB to bytes for internal use
		RecursiveFlag:  config.RecursiveFlag,
		MaxDepthFlag:   config.MaxDepthFlag,
		DependencyFlag: config.DependencyFlag,
		GitFlag:        config.GitFlag,
	}

	return scanCodebase(config)
}

// GetSupportedLanguages returns a list of all supported programming languages that PathFinder can analyze.
func GetSupportedLanguages() []string {
	languages := make([]string, 0, len(languageDefinitions))

	for _, langDef := range languageDefinitions {
		languages = append(languages, langDef.Name)
	}
	return languages
}
