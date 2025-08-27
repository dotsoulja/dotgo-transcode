package segmenter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/executil"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// SegmentMedia performs adaptive segmentation of transcoded media variants.
// It uses ffmpeg to slice each variant into HLS or DASH segments, aligning
// segment boundaries to keyframes when possible for ABR resilience.
//
// If SegmentLength is not explicitly set in the profile, this function defaults
// to using the average keyframe interval extracted from analyzer.MediaInfo.
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

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, variant := range result.Variants {
		wg.Add(1)
		go func(variant transcoder.ResolutionVariant) {
			defer wg.Done()

			inputPath := filepath.Join(result.OutputDir, variant.OutputFilename)
			label := LabelFromFilename(variant.OutputFilename)
			outputDir := filepath.Join(result.OutputDir, label)

			// Create output directory for segments
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				mu.Lock()
				segResult.Success = false
				segResult.Errors = append(segResult.Errors, *NewSegmenterError(
					"filesystem", fmt.Sprintf("failed to create segment dir for %s", label), err,
				))
				mu.Unlock()
				return
			}

			// Analyze media to extract keyframe interval and timestamps
			mediaInfo, err := analyzer.AnalyzeMedia(inputPath)
			if err != nil {
				log.Printf("⚠️ Failed to analyze media for %s: %v", inputPath, err)
			}

			// Determine segment length
			segmentLength := result.Profile.SegmentLength
			if segmentLength == 0 && mediaInfo != nil && mediaInfo.KeyframeInterval > 0 {
				segmentLength = int(mediaInfo.KeyframeInterval + 0.5) // round up
				log.Printf("⏱️ Using keyframe-aligned segment length: %ds for %s", segmentLength, label)
			}

			// Build ffmpeg command with optional keyframe alignment
			manifestName := fmt.Sprintf("%s.%s", label, manifestExtension(format))
			manifestPath := filepath.Join(outputDir, manifestName)
			cmd := buildSegmentCommand(inputPath, outputDir, manifestName, format, segmentLength, mediaInfo)

			log.Printf("📦 Segmenting %s into %s format", variant.OutputFilename, format)
			if err := executil.RunCommand(cmd); err != nil {
				mu.Lock()
				segResult.Success = false
				segResult.Errors = append(segResult.Errors, *NewSegmenterError(
					"segment", fmt.Sprintf("failed to segment %s", variant.OutputFilename), err,
				))
				mu.Unlock()
				return
			}

			mu.Lock()
			segResult.Manifests = append(segResult.Manifests, manifestPath)
			mu.Unlock()
		}(variant)
	}

	wg.Wait()
	return segResult, nil
}
