package metrics

import (
	"path/filepath"
	"strings"
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
