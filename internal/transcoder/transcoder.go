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
// This version includes average progress logging across all active variants,
// and gracefully shuts down the progress ticker once transcoding completes.
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

	log.Printf("🚀 Starting concurrent transcoding for %d variants...", len(allowed))
	start := time.Now()

	// Track seen variants to avoid duplicates
	seen := make(map[string]bool)
	var seenMu sync.Mutex

	// Track per-variant progress for average logging
	progressMap := make(map[string]float64)
	var progressMu sync.Mutex

	// Channel to signal when transcoding is complete
	done := make(chan struct{})

	// Launch goroutine to emit average progress every 2 seconds
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				progressMu.Lock()
				if len(progressMap) == 0 {
					progressMu.Unlock()
					continue
				}
				var total float64
				for _, v := range progressMap {
					total += v
				}
				avg := total / float64(len(progressMap))
				log.Printf("[progress][⏳ Average across %d variants] - %.2f%%", len(progressMap), avg)
				progressMu.Unlock()

			case <-done:
				return // ✅ Stop emitting once transcoding is done
			}
		}
	}()

	var wg sync.WaitGroup

	for _, v := range allowed {
		wg.Add(1)
		go func(v Variant) {
			defer wg.Done()

			key := fmt.Sprintf("%s_%s", v.Resolution, v.Bitrate)

			// Ensure variant is not duplicated
			seenMu.Lock()
			if seen[key] {
				logger.LogVariant(key, "⚠️ Skipping duplicate variant")
				seenMu.Unlock()
				return
			}
			seen[key] = true
			seenMu.Unlock()

			// Resolve dimensions
			width, height, err := scaler.DimensionsForLabel(v.Resolution)
			if err != nil {
				logger.LogVariant(v.Resolution, "⚠️ Unknown resolution label - using source dimensions")
				width = media.Width
				height = media.Height
			}

			// Build output path and ffmpeg command
			outputFilename := fmt.Sprintf("%s_%s_%sbps.mp4", slug, v.Resolution, v.Bitrate)
			outputPath := filepath.Join(slugDir, outputFilename)
			cmd := buildFFmpegCommand(profile, v)
			cmd[len(cmd)-1] = outputPath

			logger.LogVariant(key, fmt.Sprintf("🔧 Building ffmpeg command: %s", strings.Join(cmd, " ")))

			// Execute ffmpeg with progress tracking
			err = executil.RunCommandWithProgress(cmd, media.Duration, func(percent float64) {
				progressMu.Lock()
				progressMap[key] = percent
				progressMu.Unlock()
			})
			if err != nil {
				logger.LogError("transcode", err)
				seenMu.Lock()
				result.Success = false
				result.Errors = append(result.Errors, NewTranscoderError(
					"execution", "transcode", profile.InputPath, outputPath,
					"ffmpeg command failed", cmd, 1, err,
				))
				seenMu.Unlock()
				return
			}

			// Record successful variant
			seenMu.Lock()
			result.Variants = append(result.Variants, ResolutionVariant{
				Width:          width,
				Height:         height,
				Bitrate:        v.Bitrate,
				ScaleFlag:      "auto",
				OutputFilename: outputFilename,
			})
			seenMu.Unlock()

			logger.LogVariant(key, fmt.Sprintf("✅ Transcoding succeeded: (%dx%d) @ %s)", width, height, v.Bitrate))
		}(v)
	}

	wg.Wait()
	close(done) // ✅ Signal progress ticker to stop
	logger.LogStage("complete", fmt.Sprintf("🏁 All transcoding tasks completed in %s", time.Since(start)))

	return result, nil
}
