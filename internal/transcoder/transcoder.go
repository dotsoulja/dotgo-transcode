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
	"github.com/dotsoulja/dotgo-transcode/internal/utils/metadata"
)

// Transcode orchestrates resolution-aware transcoding for a given media file.
// It filters out variants that exceed source resolution, then concurrently
// transcodes each allowed variant. All variants are encoded to ensure uniform
// segment timing and consistent GOP structure.
// Accepts a TranscodeLogger for structured, stage-aware logging.
func Transcode(profile *TranscodeProfile, media *analyzer.MediaInfo, logger TranscodeLogger) (*TranscodeResult, error) {
	// Validate input/output paths and ensure output directory exists
	logger.LogStage("init", "Validating input/output paths")
	if err := validatePaths(profile.InputPath, profile.OutputDir); err != nil {
		logger.LogError("validation", err)
		return nil, NewTranscoderError(
			"validation", "path_check", profile.InputPath, profile.OutputDir,
			"invalid input or output path", nil, 0, err,
		)
	}

	// Derive slug from input filename and create output subdirectory
	baseName := filepath.Base(profile.InputPath)
	slug := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	slugDir := filepath.Join(profile.OutputDir, slug)

	if err := os.MkdirAll(slugDir, os.ModePerm); err != nil {
		logger.LogError("filesystem", err)
		return nil, NewTranscoderError(
			"filesystem", "mkdir", profile.InputPath, slugDir,
			"failed to create slug directory", nil, 0, err,
		)
	}

	// Initialize result container
	result := &TranscodeResult{
		InputPath: profile.InputPath,
		OutputDir: slugDir,
		Duration:  media.Duration,
		Success:   true,
		Profile:   profile,
	}

	// Save duration to json for frontend consumption
	if err := metadata.WriteMetadata(slugDir, profile.SegmentLength, media.Duration); err != nil {
		logger.LogError("metadata", err)
	}

	// Filter out resolutions that exceed source media height
	allowed := []Variant{}
	for _, v := range profile.Variants {
		_, h, err := scaler.DimensionsForLabel(v.Resolution)
		if err != nil {
			logger.LogVariant(v.Resolution, "⚠️ Unknown resolution label - skipping")
			continue
		}
		if h <= media.Height {
			allowed = append(allowed, v)
		} else {
			logger.LogVariant(v.Resolution, fmt.Sprintf("⛔ Skipping - source resolution (%dp) too low", media.Height))
		}
	}

	// Log resolution filtering summary
	logger.LogStage("filter", fmt.Sprintf("🎞️ Source resolution: %dx%d", media.Width, media.Height))
	logger.LogStage("filter", fmt.Sprintf("✅ Proceeding with %d allowed variants", len(allowed)))

	// Track seen variants to avoid duplicates
	seen := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	log.Printf("🚀 Starting concurrent transcoding for %d variants...", len(allowed))
	start := time.Now()

	for _, v := range allowed {
		wg.Add(1)
		go func(v Variant) {
			defer wg.Done()

			key := fmt.Sprintf("%s_%s", v.Resolution, v.Bitrate)

			mu.Lock()
			if seen[key] {
				logger.LogVariant(key, "⚠️ Skipping duplicate variant")
				mu.Unlock()
				return
			}
			seen[key] = true
			mu.Unlock()

			width, height, err := scaler.DimensionsForLabel(v.Resolution)
			if err != nil {
				logger.LogVariant(v.Resolution, "⚠️ Unknown resolution label - using source dimensions")
				width = media.Width
				height = media.Height
			}

			outputFilename := fmt.Sprintf("%s_%s_%sbps.mp4", slug, v.Resolution, v.Bitrate)
			outputPath := filepath.Join(slugDir, outputFilename)
			cmd := buildFFmpegCommand(profile, v)
			cmd[len(cmd)-1] = outputPath

			logger.LogVariant(key, fmt.Sprintf("🔧 Building ffmpeg command: %s", strings.Join(cmd, " ")))

			err = executil.RunCommandWithProgress(cmd, media.Duration, func(percent float64) {
				logger.LogProgress(key, percent)
			})
			if err != nil {
				logger.LogError("transcode", err)
				mu.Lock()
				result.Success = false
				result.Errors = append(result.Errors, NewTranscoderError(
					"execution", "transcode", profile.InputPath, outputPath,
					"ffmpeg command failed", cmd, 1, err,
				))
				mu.Unlock()
				return
			}

			mu.Lock()
			result.Variants = append(result.Variants, ResolutionVariant{
				Width:          width,
				Height:         height,
				Bitrate:        v.Bitrate,
				ScaleFlag:      "auto",
				OutputFilename: outputFilename,
			})
			mu.Unlock()

			logger.LogVariant(key, fmt.Sprintf("✅ Transcoding succeeded: (%dx%d @ %s)", width, height, v.Bitrate))
		}(v)
	}

	wg.Wait()
	logger.LogStage("complete", fmt.Sprintf("⏱️ All variants completed in %s", time.Since(start)))

	return result, nil
}
