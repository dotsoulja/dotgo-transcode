// Package segmenter orchestrates the segmentation phase of the transcoding pipeline.
// It takes transcoded variants and slices them into HLS or DASH-compatible segments.
// This file exposes the high-level SegmentMedia function used by downstream workflows
package segmenter

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// SegmentMedia performs media segmentation for adaptive streaming.
// It accepts a TranscodeResult and segments each variant into chunks
// using ffmpeg with HLS or DASH flags. Returns a SegmentResult with
// manifest paths and error metadata
func SegmentMedia(result *transcoder.TranscodeResult, format string) (*SegmentResult, error) {
	if result == nil || len(result.Variants) == 0 {
		return nil, NewSegmenterError("validate", "no variants to segment", nil)
	}

	segResult := &SegmentResult{
		OutputDir: result.OutputDir,
		Format:    format,
		Success:   true,
	}

	for _, variant := range result.Variants {
		inputPath := filepath.Join(result.OutputDir, variant.OutputFilename)
		outputDir := filepath.Join(result.OutputDir, variant.Label())
		manifestName := fmt.Sprintf("%s.%s", variant.Label(), manifestExtension(format))
		manifestPath := filepath.Join(outputDir, manifestName)

		cmd := buildSegmentCommand(inputPath, outputDir, manifestName, format)

		log.Printf("📦 Segmenting %s into %s format", variant.OutputFilename, format)
		if err := runCommand(cmd); err != nil {
			segResult.Success = false
			segResult.Errors = append(segResult.Errors, NewSegmenterError(
				"segment", fmt.Sprintf("failed to segment %s", variant.OutputFilename), err,
			))
			continue
		}

		segResult.Manifests = append(segResult.Manifests, manifestPath)
	}

	return segResult, nil
}
