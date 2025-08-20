package metrics

type Flags struct {
	PathFlag   string // Path to the codebase/repository
	HiddenFlag bool   // Include hidden files and directories
	// TODO: add more flags (like dependencies)
}

// CommentType defines the comment syntax for a programming language
type CommentType struct {
	SingleLine string // Single line comment prefixes
	BLockStart string // Block comment start
	BlockEnd   string // Block comment end
}

type LanguageDefinition struct {
	Name string      // Programming language name
	Type CommentType // Comment syntax of the language
	Ext  []string    // File extensions associated with the language
}

type LanguageMetrics struct {
	Language string // Programming language name
	Files    int    // Number of files of this specific language
	Code     int    // lines of code
	Comments int    // number of comment lines
	Blanks   int    // number of blank lines
	Lines    int    // Total lines (code + comments + blanks)
}

type AnnotationMetrics struct {
	TotalTODO        int // Total TODO annotations
	TotalFIXME       int // Total FIXME annotations
	TotalHACK        int // Total HACK annotations
	TotalAnnotations int // Total of all annotations (TODO + FIXME + HACK)
}

type CodebaseMetrics struct {
	TotalFiles     int // Total number of files
	TotalDirs      int // Total number of directories
	TotalLanguages int // Total number of programming languages
	TotalCode      int // Total code lines
	TotalComments  int // Total comment lines
	TotalBlanks    int // Total blank lines
	TotalLines     int // Total lines (code + comments + blanks)
}

type FileMetricsReport struct {
	Path    string          // File path
	Metrics LanguageMetrics // Language metrics for this specific file
}

type DirMetricsReport struct {
	Directory string // Directory path
	Lines     int    // Total lines in the directory (sum of all files)
}

type LanguageMetricsReport struct {
	Percentage float64         // Percentage of total lines in the codebase
	Metrics    LanguageMetrics // Metrics for the language
}

type CodebaseReport struct {
	LanguageMetrics   []LanguageMetricsReport
	FileMetrics       []FileMetricsReport
	DirMetrics        []DirMetricsReport
	CodebaseMetrics   CodebaseMetrics
	AnnotationMetrics AnnotationMetrics
	// TODO: add more reports (like dependencies)
	// TODO: add time taken to analyze the codebase
}
