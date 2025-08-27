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

// Transcode orchestrates resolution-aware transcoding for a given media file.
// It filters out variants that exceed source resolution, then concurrently
// transcodes each allowed variant. If a resolution matches the source exactly,
// the original file is copied into the output directory and marked as passthrough.
func Transcode(profile *TranscodeProfile, media *analyzer.MediaInfo) (*TranscodeResult, error) {
	// Validate input/output paths and ensure output directory exists
	if err := validatePaths(profile.InputPath, profile.OutputDir); err != nil {
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

	// Filter out resolutions that exceed source media height
	allowed := []string{}
	for _, res := range profile.Resolutions {
		_, h, err := scaler.DimensionsForLabel(res)
		if err != nil {
			log.Printf("⚠️ Unknown resolution label: %s — skipping", res)
			continue
		}
		if h <= media.Height {
			allowed = append(allowed, res)
		} else {
			log.Printf("🚫 Skipping %s — source resolution (%dp) too low", res, media.Height)
		}
	}

	// Log resolution filtering summary
	log.Printf("\n📋 Profile requested %d variants: %v", len(profile.Resolutions), profile.Resolutions)
	log.Printf("🎞️ Source resolution: %dx%d", media.Width, media.Height)
	log.Printf("✅ Proceeding with %d allowed variants: %v\n", len(allowed), allowed)

	// Detect and handle passthrough variant before launching goroutines
	filtered := []string{}
	for _, res := range allowed {
		width, height, err := scaler.DimensionsForLabel(res)
		if err != nil {
			continue
		}
		if width == media.Width && height == media.Height {
			bitrate := profile.Bitrate[res]
			outputFilename := fmt.Sprintf("%s_%s_%skbps_passthrough.mp4", slug, res, bitrate)
			outputPath := filepath.Join(slugDir, outputFilename)

			if err := copyFile(profile.InputPath, outputPath); err != nil {
				result.Success = false
				result.Errors = append(result.Errors, NewTranscoderError(
					"filesystem", "copy_passthrough", profile.InputPath, outputPath,
					"failed to copy source file for passthrough variant", nil, 0, err,
				))
			} else {
				result.Variants = append(result.Variants, ResolutionVariant{
					Width:          width,
					Height:         height,
					Bitrate:        bitrate,
					ScaleFlag:      "passthrough",
					OutputFilename: outputFilename,
				})
				log.Printf("📦 Passthrough variant copied for %s (%dx%d @ %s)", res, width, height, bitrate)
			}
		} else {
			filtered = append(filtered, res)
		}
	}

	// Track seen variants to avoid duplicates
	seen := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	log.Printf("🚀 Starting concurrent transcoding for %d variants...", len(filtered))
	start := time.Now()

	for _, res := range filtered {
		wg.Add(1)
		go func(res string) {
			defer wg.Done()

			bitrate := profile.Bitrate[res]
			key := fmt.Sprintf("%s_%s", res, bitrate)

			mu.Lock()
			if seen[key] {
				log.Printf("⚠️ Skipping duplicate variant: %s", key)
				mu.Unlock()
				return
			}
			seen[key] = true
			mu.Unlock()

			width, height, err := scaler.DimensionsForLabel(res)
			if err != nil {
				log.Printf("⚠️ Unknown resolution label: %s — using source dimensions", res)
				width = media.Width
				height = media.Height
			}

			outputFilename := fmt.Sprintf("%s_%s_%skbps.mp4", slug, res, bitrate)
			outputPath := filepath.Join(slugDir, outputFilename)
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
