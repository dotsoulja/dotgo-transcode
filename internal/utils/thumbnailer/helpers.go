package thumbnailer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// GenerateTimestamps returns a slice of timestamps (in seconds) based on the
// total duration and effective segment length. These timestamps are used to
// generate thumbnails at regular intervals.
//
// If duration is 0 or segmentLength is invalid, it returns an empty slice.
func GenerateTimestamps(duration float64, segmentLength int) []float64 {
	if duration <= 0 || segmentLength <= 0 {
		log.Printf("⚠️ Invalid duration (%.2f) or segment length (%d), skipping timestamp generation", duration, segmentLength)
		return []float64{}
	}

	var timestamps []float64
	for t := 0.0; t < duration; t += float64(segmentLength) {
		timestamps = append(timestamps, t)
	}
	return timestamps
}

// EnsureThumbnailDir creates the thumbnails directory inside the given slug path
// if it doesn't already exist. Returns the full path to the directory.
func EnsureThumbnailDir(outputDir string) (string, error) {
	thumbDir := filepath.Join(outputDir, "thumbnails")
	if _, err := os.Stat(thumbDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(thumbDir, 0755); mkErr != nil {
			return "", fmt.Errorf("failed to create thumbnails directory: %w", mkErr)
		}
	}
	return thumbDir, nil
}

// GetVariantPath returns the full path to the transcoded .mp4 file that matches
// the source height. Assumes outputDir already includes the slug directory.
// Filename format: <slug>_<height>p_<bitrate>kbps.mp4
func GetVariantPath(outputDir string, slug string, height int, bitrate int) (string, error) {
	filename := fmt.Sprintf("%s_%dp_%dkbps.mp4", slug, height, bitrate)
	fullPath := filepath.Join(outputDir, filename)

	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("transcoded variant not found: %s", fullPath)
	}
	return fullPath, nil
}

// FormatTimestampFilename returns a filename for a thumbnail based on the timestamp.
// Example: thumb_004.jpg for timestamp 4.0
func FormatTimestampFilename(timestamp float64) string {
	return fmt.Sprintf("thumb_%03d.jpg", int(timestamp))
}
