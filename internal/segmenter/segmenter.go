// Package segmenter orchestrates the segmentation phase of the transcoding pipeline.
// It takes transcoded variants and slices them into HLS or DASH-compatible segments.
// This file exposes the high-level SegmentMedia function used by downstream workflows.
package segmenter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dotsoulja/dotgo-transcode/internal/executil"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// SegmentMedia performs media segmentation for adaptive streaming.
// It accepts a TranscodeResult and segments each variant into chunks
// using ffmpeg with HLS or DASH flags. Returns a SegmentResult with
// manifest paths and error metadata.
//
// Output structure:
//
//	media/output/<slug>/<resolution>/
//	  ├── segment_000.ts
//	  └── <resolution>.m3u8
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
		// Input file: media/output/<slug>/<slug>_<resolution>_<bitrate>.mp4
		inputPath := filepath.Join(result.OutputDir, variant.OutputFilename)

		// Resolution label derived from filename (e.g. "720p")
		label := LabelFromFilename(variant.OutputFilename)

		// Output directory for segments: media/output/<slug>/<label>/
		outputDir := filepath.Join(result.OutputDir, label)

		// Ensure resolution-specific directory exists
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			segResult.Success = false
			segResult.Errors = append(segResult.Errors, *NewSegmenterError(
				"filesystem", fmt.Sprintf("failed to create segment dir for %s", label), err,
			))
			continue
		}

		// Manifest filename: <label>.m3u8 or <label>.mpd
		manifestName := fmt.Sprintf("%s.%s", label, manifestExtension(format))
		manifestPath := filepath.Join(outputDir, manifestName)

		// Build ffmpeg command for segmentation
		cmd := buildSegmentCommand(inputPath, outputDir, manifestName, format)

		log.Printf("📦 Segmenting %s into %s format", variant.OutputFilename, format)
		if err := executil.RunCommand(cmd); err != nil {
			segResult.Success = false
			segResult.Errors = append(segResult.Errors, *NewSegmenterError(
				"segment", fmt.Sprintf("failed to segment %s", variant.OutputFilename), err,
			))
			continue
		}

		// Track successful manifest path
		segResult.Manifests = append(segResult.Manifests, manifestPath)
	}

	return segResult, nil
}
