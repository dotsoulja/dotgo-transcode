// Package segmenter handles adaptive segmentation of transcoded media.
// It slices each resolution variant into HLS or DASH segments, aligning
// segment boundaries to keyframes when needed for ABR resilience.
package segmenter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/executil"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/helpers"
)

// SegmentMedia performs segmentation of transcoded media variants into HLS or DASH format.
// It uses ffmpeg to slice each variant into segments, optionally aligning segment boundaries
// to keyframes for adaptive bitrate (ABR) safety.
//
// Segment length is determined by the TranscodeProfile:
//   - If SegmentLength > 0, that value is used directly.
//   - If SegmentLength == 0, the function falls back to the keyframe interval from MediaInfo.
//
// This function assumes that MediaInfo has already been extracted once upstream (e.g. in main.go)
// and is passed in to avoid redundant analysis.
//
// Output structure per variant:
//
//	media/output/<slug>/<resolution>_<bitrate>kbps/
//	  ├── segment_000.ts
//	  └── <resolution>_<bitrate>.m3u8
func SegmentMedia(result *transcoder.TranscodeResult, format string, media *analyzer.MediaInfo) (*SegmentResult, error) {
	if result == nil || len(result.Variants) == 0 {
		return nil, NewSegmenterError("validate", "no variants to segment", nil)
	}

	// Initialize result container
	segResult := &SegmentResult{
		OutputDir: result.OutputDir,
		Format:    format,
		Success:   true,
		Media:     media,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Segment each resolution variant concurrently
	for _, variant := range result.Variants {
		wg.Add(1)
		go func(variant transcoder.ResolutionVariant) {
			defer wg.Done()

			inputPath := filepath.Join(result.OutputDir, variant.OutputFilename)

			// Normalize bitrate string (e.g. "3000k" -> "3000kbps")
			bitrateInt := helpers.ParseBitrateKbps(variant.Bitrate)
			bitrateLabel := "unknown"
			if bitrateInt > 0 {
				bitrateLabel = fmt.Sprintf("%dkbps", bitrateInt)
			}

			// Construct directory label using resolution and normalized bitrate
			label := fmt.Sprintf("%dp_%s", variant.Height, bitrateLabel)
			outputDir := filepath.Join(result.OutputDir, label)

			// Create output directory for segments
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				mu.Lock()
				segResult.Success = false
				segResult.Errors = append(segResult.Errors, NewSegmenterError(
					"filesystem", fmt.Sprintf("failed to create segment dir for %s", label), err,
				))
				mu.Unlock()
				return
			}

			// Determine segment length based on profile or keyframe interval
			segmentLength := result.Profile.SegmentLength
			if segmentLength == 0 && media != nil && media.KeyframeInterval > 0 {
				segmentLength = int(media.KeyframeInterval + 0.5) // round up to nearest second
				log.Printf("⏰ Using keyframe-aligned segment length: %ds for %s", segmentLength, label)
			} else if segmentLength > 0 {
				log.Printf("📐 Using configured segment length: %ds for %s", segmentLength, label)
			} else {
				log.Printf("⚠️ No segment length or keyframe data available, defaulting to 4s for %s", label)
				segmentLength = 4
			}

			// Build ffmpeg command for segmentation
			manifestName := fmt.Sprintf("%s.%s", label, manifestExtension(format))
			manifestPath := filepath.Join(outputDir, manifestName)
			cmd := buildSegmentCommand(inputPath, outputDir, manifestPath, format, segmentLength, media)

			log.Printf("🔪 Segmenting %s into %s format", variant.OutputFilename, format)
			log.Printf("FFmpeg command: %s", strings.Join(cmd, " "))
			if err := executil.RunCommand(cmd); err != nil {
				mu.Lock()
				segResult.Success = false
				segResult.Errors = append(segResult.Errors, NewSegmenterError(
					"segment", fmt.Sprintf("failed to segment %s", label), err,
				))
				mu.Unlock()
				return
			}

			// Record manifest path
			mu.Lock()
			segResult.Manifests = append(segResult.Manifests, manifestPath)
			mu.Unlock()
		}(variant)
	}

	wg.Wait()
	return segResult, nil
}
