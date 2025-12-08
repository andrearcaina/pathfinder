package pathfinder

// Config configures the behavior of the scanner.
type Config struct {
	// PathFlag is the root path to the codebase or repository to scan.
	PathFlag string

	// HiddenFlag, if true, includes hidden files and directories (starting with .) in the scan.
	HiddenFlag bool

	// BufferSizeFlag sets the buffer size (in bytes) for reading files.
	// specific values like 4KB, 8KB, 16KB is usually the best performance.
	BufferSizeFlag int

	// RecursiveFlag, if true, scans subdirectories recursively.
	RecursiveFlag bool

	// MaxDepthFlag limits the recursion depth. Set to -1 for no limit.
	// Only applies if RecursiveFlag is true.
	MaxDepthFlag int

	// DependencyFlag, if true, attempts to analyze dependency files (e.g. go.mod, package.json).
	DependencyFlag bool

	// GitFlag, if true, analyzes git information (commits, history).
	GitFlag bool
}

// CommentType defines the comment syntax markers for a programming language.
type CommentType struct {
	SingleLine string // Prefix for single-line comments (e.g., "//" or "#")
	BlockStart string // Start marker for block comments (e.g., "/*")
	BlockEnd   string // End marker for block comments (e.g., "*/")
}

// LanguageDefinition maps a programming language to its file extensions and comment syntax.
type LanguageDefinition struct {
	Name string      // The common name of the language (e.g., "Go", "Python")
	Type CommentType // The comment syntax definition
	Ext  []string    // List of file extensions (e.g., ".go", ".py")
}

// LanguageMetrics contains the raw counts for a specific language.
type LanguageMetrics struct {
	Language string // Name of the language
	Files    int    // Number of files detected
	Code     int    // Lines of actual code
	Comments int    // Lines of comments
	Blanks   int    // Empty lines
	Lines    int    // Total lines (Code + Comments + Blanks)
}

// AnnotationMetrics tracks special comment tags like TODO, FIXME, and HACK.
type AnnotationMetrics struct {
	TotalTODO        int // Count of "TODO" tags
	TotalFIXME       int // Count of "FIXME" tags
	TotalHACK        int // Count of "HACK" tags
	TotalAnnotations int // Sum of all annotation types
}

// CodebaseMetrics aggregates statistics for the entire scanned project.
type CodebaseMetrics struct {
	TotalFiles     int // Total files scanned
	TotalDirs      int // Total directories encountered
	TotalLanguages int // Number of distinct languages detected
	TotalCode      int // Total lines of code across all languages
	TotalComments  int // Total lines of comments across all languages
	TotalBlanks    int // Total blank lines across all languages
	TotalLines     int // Grand total of all lines
}

// DependencyFile represents a manifest file found in the project (e.g., go.mod).
type DependencyFile struct {
	Path         string   // File path to the manifest
	Type         string   // Type of dependency manager (e.g., "Go Modules", "npm")
	Dependencies []string // List of dependencies found in the file
}

// DependencyMetrics aggregates dependency information found during the scan.
type DependencyMetrics struct {
	TotalDependencies int              // Total count of individual dependencies found
	DependencyFiles   []DependencyFile // List of files that were parsed for dependencies
}

// FileMetricsReport contains metrics for a single file.
type FileMetricsReport struct {
	Path    string          // Relative path to the file
	Metrics LanguageMetrics // The metrics calculated for this file
}

// DirMetricsReport contains metrics for a specific directory.
type DirMetricsReport struct {
	Directory  string  // Path to the directory
	Percentage float64 // Percentage of the codebase contained in this directory
	Lines      int     // Total lines in this directory
}

// LanguageMetricsReport wraps LanguageMetrics with a percentage relative to the whole codebase.
type LanguageMetricsReport struct {
	Percentage float64         // Percentage of the codebase written in this language
	Metrics    LanguageMetrics // The raw metrics
}

// CodebaseReport is the final output structure containing all analysis results.
type CodebaseReport struct {
	LanguageMetrics   []LanguageMetricsReport
	FileMetrics       []FileMetricsReport
	DirMetrics        []DirMetricsReport
	CodebaseMetrics   CodebaseMetrics
	AnnotationMetrics AnnotationMetrics
	DependencyMetrics DependencyMetrics
}
