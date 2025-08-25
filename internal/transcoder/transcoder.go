package transcoder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/executil"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
)

// Transcode orchestrates the transcoding process for a given media file.
// It accepts a validated TranscodeProfile and extracted MediaInfo,
// then concurrently transcodes each target resolution using goroutines.
// Each variant is processed independently, and results are aggregated.
//
// This function does not segment or mux; it focuses on resolution variants.
// Segmenting and manifest generation are handled in later phases.
//
// Output structure:
//
//	media/output/<slug>/<slug>_<resolution>_<bitrate>.mp4
func Transcode(profile *TranscodeProfile, media *analyzer.MediaInfo) (*TranscodeResult, error) {
	if err := validatePaths(profile.InputPath, profile.OutputDir); err != nil {
		return nil, NewTranscoderError(
			"validation", "path_check", profile.InputPath, profile.OutputDir,
			"invalid input or output path", nil, 0, err,
		)
	}

	baseName := filepath.Base(profile.InputPath)
	slug := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	slugDir := filepath.Join(profile.OutputDir, slug)

	if err := os.MkdirAll(slugDir, os.ModePerm); err != nil {
		return nil, NewTranscoderError(
			"filesystem", "mkdir", profile.InputPath, slugDir,
			"failed to create slug directory", nil, 0, err,
		)
	}

	result := &TranscodeResult{
		InputPath: profile.InputPath,
		OutputDir: slugDir,
		Duration:  media.Duration,
		Success:   true,
	}

	seen := make(map[string]bool)
	var mu sync.Mutex // protects result and seen
	var wg sync.WaitGroup

	log.Printf("🚀 Starting concurrent transcoding for %d variants...", len(profile.Resolutions))
	start := time.Now()

	for _, res := range profile.Resolutions {
		wg.Add(1)
		go func(res string) {
			defer wg.Done()

			bitrate := profile.Bitrate[res]
			outputFilename := fmt.Sprintf("%s_%s_%skbps.mp4", slug, res, bitrate)
			outputPath := filepath.Join(slugDir, outputFilename)
			key := fmt.Sprintf("%s_%s", res, bitrate)

			mu.Lock()
			if seen[key] {
				log.Printf("⚠️ Skipping duplicate variant: %s", key)
				mu.Unlock()
				return
			}
			seen[key] = true
			mu.Unlock()

			cmd := buildFFmpegCommand(profile, res)
			cmd[len(cmd)-1] = outputPath

			log.Printf("🔧 Building ffmpeg command for %s (%s)", res, bitrate)
			log.Printf("🎞️ Transcoding to %s: %s", res, strings.Join(cmd, " "))

			if err := executil.RunCommand(cmd); err != nil {
				mu.Lock()
				result.Success = false
				result.Errors = append(result.Errors, NewTranscoderError(
					"execution", "transcode", profile.InputPath, outputPath,
					"ffmpeg command failed", cmd, 1, err,
				))
				mu.Unlock()
				return
			}

			width, height, err := scaler.DimensionsForLabel(res)
			if err != nil {
				log.Printf("⚠️ Unknown resolution label: %s - using source dimensions", res)
				width = media.Width
				height = media.Height
			}

			mu.Lock()
			result.Variants = append(result.Variants, ResolutionVariant{
				Width:          width,
				Height:         height,
				Bitrate:        bitrate,
				ScaleFlag:      "auto",
				OutputFilename: outputFilename,
			})
			mu.Unlock()

			log.Printf("✅ Transcoding succeeded for %s (%dx%d @ %s)", res, width, height, bitrate)
		}(res)
	}

	wg.Wait()
	log.Printf("⏱️ All variants completed in %s", time.Since(start))

	return result, nil
}
