# pathfinder
Blazingly fast, lightweight CLI to map & track your codebase.

### Overview

`pathfinder` is a command-line tool written in Go that scans a specified directory (and its subdirectories) to count the number of files, directories, and total lines of code.

It is designed to be fast and efficient, leveraging Go's concurrency features to process files in parallel.

It also has a library API that you can use to integrate its functionality into your own Go applications.

### Installation

**As a CLI Tool:**
```bash
go install github.com/andrearcaina/pathfinder@latest
```

**As a Go Library:**
```bash
go get github.com/andrearcaina/pathfinder
````

### Go Example Usage
Here is a simple example of how to use pathfinder as a library in your Go code:
```go
package main

import (
	"fmt"
	"log"

	// Import the pathfinder package
	"github.com/andrearcaina/pathfinder/pkg/pathfinder"
)

func main() {
	config := pathfinder.Config{
		PathFlag:       ".",   // Scan the current directory
		RecursiveFlag:  true,  // Scan subdirectories recursively
		HiddenFlag:     false, // Skip hidden files/directories (like .git)
		DependencyFlag: true,  // Analyze dependency files (go.mod, package.json, etc.)
		BufferSizeFlag: 4096,  // Set read buffer size (4KB is usually optimal)
	}
	
	report, err := pathfinder.Scan(config)
	if err != nil {
		log.Fatalf("Failed to scan codebase: %v", err)
	}

	fmt.Printf("Found %d files across %d languages.\n\n",
		report.CodebaseMetrics.TotalFiles,
		report.CodebaseMetrics.TotalLanguages)

	fmt.Println("Language Breakdown:")
	for _, lang := range report.LanguageMetrics {
		fmt.Printf("• %s: %d lines (%.2f%%)\n",
			lang.Metrics.Language,
			lang.Metrics.Lines,
			lang.Percentage,
		)
	}
}
```

### CLI Usage

Below I ran `pathfinder` on this codebase with the `-R` flag to recursively scan all subdirectories. Image was taken at 2025-08-30 3:08 PM EST.
![example1.png](images/example1.png)

Then I ran the same command, but instead on my installed Go libraries and packages in WSL to benchmark performance.  Image was taken at 2025-08-30 3:08 PM EST.
![example2.png](images/example2.png)

You can see that it found **73,103** files, **18,219** directories, and **44,858,625** total lines.

```bash
> time ./bin/pathfinder ../../../go/ -R

# time output (as of 2025-08-30 3:08 PM EST)
./bin/pathfinder -p ../../../go/ -R  5.88s user 8.15s system 281% cpu 4.995 total
``` 
This only took 4.995 seconds (on my machine), utilizing 281% of the CPU. 
This is because `pathfinder` uses goroutines for concurrent file reading and processing.

I then ran it again on the same directory:
```bash
> time ./bin/pathfinder ../../../go/ -R

...

# time output (as of 2025-08-30 3:08 PM EST)
./bin/pathfinder -p ../../../go/ -R  4.69s user 2.97s system 686% cpu 1.117 total
```
This time it took only 1.117 seconds to run, utilizing 665% of the CPU.
The second run is much faster because the OS caches file data in memory, 
reducing I/O overhead and allowing goroutines to utilize more cores efficiently.