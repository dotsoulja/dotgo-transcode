package segmenter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

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

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, variant := range result.Variants {
		wg.Add(1)
		go func(variant transcoder.ResolutionVariant) {
			defer wg.Done()

			inputPath := filepath.Join(result.OutputDir, variant.OutputFilename)
			label := LabelFromFilename(variant.OutputFilename)
			outputDir := filepath.Join(result.OutputDir, label)

			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				mu.Lock()
				segResult.Success = false
				segResult.Errors = append(segResult.Errors, *NewSegmenterError(
					"filesystem", fmt.Sprintf("failed to create segment dir for %s", label), err,
				))
				mu.Unlock()
				return
			}

			manifestName := fmt.Sprintf("%s.%s", label, manifestExtension(format))
			manifestPath := filepath.Join(outputDir, manifestName)
			cmd := buildSegmentCommand(inputPath, outputDir, manifestName, format)

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
