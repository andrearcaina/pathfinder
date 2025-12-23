# API Documentation

This documentation serves as a reference for the Pathfinder API, detailing its functions, type structs, and usage.

## Type Structs
The main type struct used in the Pathfinder API is `Config`, which holds configuration options for scanning a codebase.
```go
type Config struct {
	PathFlag string
	HiddenFlag bool
	BufferSizeFlag int
	RecursiveFlag bool
	MaxDepthFlag int
	DependencyFlag bool
	GitFlag bool
	WorkerFlag int
	ThroughputFlag bool
}
```
The main output struct is `CodebaseReport`, which contains the results of a codebase scan.
```go
type CodebaseReport struct {
	LanguageMetrics    []LanguageMetricsReport
	FileMetrics        []FileMetricsReport
	DirMetrics         []DirMetricsReport
	CodebaseMetrics    CodebaseMetrics
	AnnotationMetrics  AnnotationMetrics
	DependencyMetrics  DependencyMetrics
	PerformanceMetrics PerformanceMetrics
}
```

For more information on each metric report struct, please refer to the detailed API documentation in the [pkg.go.dev](https://pkg.go.dev/github.com/andrearcaina/pathfinder/pkg/pathfinder).

## Functions
- `func Version() string`: Returns the current version of the Pathfinder API.
- `func GetSupportedLanguages() []string`: Returns a list of supported languages by the Pathfinder API.
- `func Scan(config Config) (CodebaseReport, error)`: Scans a codebase based on the provided configuration and returns a report.
- `func (c CodebaseReport) ScannedFiles() []string`: Returns a list of files that were scanned in the codebase report.
- `func (c CodebaseReport) ScannedLanguages() []string`: Returns a list of scanned language found in the codebase report.
- `func (c CodebaseReport) ScannedDirectories() []string`: Returns a list of scanned directories found in the codebase report.

## Usage
To use the Pathfinder API, first import the package and create a `Config` struct with your desired options. Then, call the `Scan` function to analyze your codebase and obtain a `CodebaseReport`.

```go
package main
import (
	"fmt"
	"log"

	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
)

func main() {
	config := pathfinder.Config{
		PathFlag:       "../..",
		RecursiveFlag:  true,
		HiddenFlag:     false,
		DependencyFlag: true,
		BufferSizeFlag: 4,
		MaxDepthFlag:   5,
		WorkerFlag:     4,
	}

	report, err := pathfinder.Scan(config)
	if err != nil {
		log.Fatalf("Error scanning codebase: %v", err)
	}

	fmt.Printf("Scanned Files: %v\n", report.ScannedFiles())
	fmt.Printf("Scanned Languages: %v\n", report.ScannedLanguages())
	fmt.Printf("Scanned Directories: %v\n", report.ScannedDirectories())
}
```
You can also pass in an empty `Config{}` struct to use default settings for scanning the current directory.

```go
report, err := pathfinder.Scan(pathfinder.Config{})
if err != nil {
	log.Fatalf("Error scanning codebase: %v", err)
}
fmt.Printf("Scanned Files: %v\n", report.ScannedFiles())
```
