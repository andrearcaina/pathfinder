package metrics

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func IsBinary(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".class", ".o", ".obj", ".pdf", ".exe":
		return true
	default:
		return false
	}
}

func ExcludeDir(name string) bool {
	// using a map for O(1) lookups (instead of a slice with O(n) lookups)
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

	_, ok := excludedDirsFromScan[name]
	return ok
}

func ExcludeFile(name string) bool {
	var excludedFilesFromScan = map[string]struct{}{
		".DS_Store":         {},
		"desktop.ini":       {},
		".gitignore":        {},
		"package-lock.json": {},
		"yarn.lock":         {},
		"go.sum":            {},
	}

	_, ok := excludedFilesFromScan[name]
	return ok
}

func HasNoExt(name string) string {
	return strings.ToLower(filepath.Ext(name))
}

func TopLevelDir(rel string) string {
	rel = filepath.ToSlash(rel)
	if i := strings.Index(rel, "/"); i != -1 {
		return rel[:i]
	}
	return "."
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%.3fs", d.Seconds())
	}
}
