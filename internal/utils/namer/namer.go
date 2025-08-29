package namer

import (
	"path/filepath"
	"strings"
)

// SlugFromPath returns a filename slug without extension or path
func SlugFromPath(inputPath string) string {
	base := filepath.Base(inputPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
