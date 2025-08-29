package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MediaMetadata captures key forensic info for frontend use
type MediaMetadata struct {
	Duration      float64 `json:"duration"`
	SegmentLength int     `json:"segment_length"`
}

// WriteMetadata writes metadata.json into the slugDir
func WriteMetadata(slugDir string, segmentLength int, duration float64) error {
	meta := MediaMetadata{Duration: duration, SegmentLength: segmentLength}
	path := filepath.Join(slugDir, "metadata.json")

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	fmt.Printf("📝 metadata.json written to %s (duration=%.2fs)\n", path, duration)
	return nil
}
