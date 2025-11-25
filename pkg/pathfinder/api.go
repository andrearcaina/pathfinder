package pathfinder

// Version returns the current version of the library.
func Version() string {
	return "v0.1.1"
}

// Scan is the main entry point for the library.
// It takes a Config and returns a detailed CodebaseReport.
func Scan(config *Config) (CodebaseReport, error) {
	if config == nil {
		// create a default config if none is provided
		config = &Config{
			PathFlag:       ".",
			HiddenFlag:     false,
			BufferSizeFlag: 16 * 1024, // 16 KB
			RecursiveFlag:  false,
			MaxDepthFlag:   -1,
			DependencyFlag: false,
			GitFlag:        false,
		}
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
