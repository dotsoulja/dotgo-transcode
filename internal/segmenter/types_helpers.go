// Package segmenter provides ergonomic methods on core types.
// This file contains helper methods for manifest labeling and variant introspection.
package segmenter

import (
	"strings"
)

// LabelFromFilename extracts the resolution label from a variant's output filename.
// Assumes format like "video_720p_1500kbps.mp4" -> returns "720p"
// This label is used to name segment directories and manifest files.
//
// If no resolution label is found, "unknown" is returned.
func LabelFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	for _, part := range parts {
		if strings.HasSuffix(part, "p") {
			return part
		}
	}
	return "unknown"
}
