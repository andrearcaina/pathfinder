package metrics

type Flags struct {
	PathFlag       string // Path to the codebase/repository
	HiddenFlag     bool   // Include hidden files and directories
	BufferSizeFlag int    // Buffer size for reading files in bytes
	RecursiveFlag  bool   // Scan directories recursively (default: false)
	MaxDepthFlag   int    // Maximum recursion depth. Only works if RecursiveFlag is true
	DependencyFlag bool   // Analyze dependencies (default: false)
	GitFlag        bool   // Analyze git information (default: false)
	// TODO: add more flags (like dependencies)
}

// CommentType defines the comment syntax for a programming language
type CommentType struct {
	SingleLine string // Single line comment prefixes
	BlockStart string // Block comment start
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

type DependencyFile struct {
	Path         string
	Type         string
	Dependencies []string
}

type DependencyMetrics struct {
	TotalDependencies int              // Total number of dependencies (libraries/packages)
	DependencyFiles   []DependencyFile // List of dependency files found
}

type GitMetrics struct {
	TotalCommits    int      // Total number of commits in the git repository
	FirstCommitISO  string   // ISO timestamp of the first commit (initial commit)
	LastCommitISO   string   // ISO timestamp of the last commit
	CommitsThisYear int      // Number of commits made in the current year
	RecentActivity  []string // List of recent commit messages (e.g. last 5 commits)
	Branches        int      // Total number of branches
}

type FileMetricsReport struct {
	Path    string          // File path
	Metrics LanguageMetrics // Language metrics for this specific file
}

type DirMetricsReport struct {
	Directory  string  // Directory path
	Percentage float64 // Percentage of total lines in the specific directory
	Lines      int     // Total lines in the directory (sum of all files)
}

type LanguageMetricsReport struct {
	Percentage float64         // Percentage of total lines in the codebase
	Metrics    LanguageMetrics // Metrics for the language
}

type PerformanceMetrics struct {
	ElapsedTime string // Total elapsed time for the scan
	// TODO: add other performance metrics if needed
}

type CodebaseReport struct {
	LanguageMetrics    []LanguageMetricsReport
	FileMetrics        []FileMetricsReport
	DirMetrics         []DirMetricsReport
	CodebaseMetrics    CodebaseMetrics
	AnnotationMetrics  AnnotationMetrics
	DependencyMetrics  DependencyMetrics
	PerformanceMetrics PerformanceMetrics
	// TODO: add git metrics
}
