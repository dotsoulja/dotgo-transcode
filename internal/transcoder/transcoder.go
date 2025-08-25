package transcoder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/executil"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
)

// Transcode orchestrates the transcoding process for a given media file.
// It accepts a validated TranscodeProfile and extracted MediaInfo,
// then iterates through each target resolution, builds ffmpeg commands,
// executes them, and returns a TranscodeResult with success/ failure metadata.
//
// This function does not segment or mux; it focuses on resolution variants.
// Segmenting and manifest generation are handled in later phases.
//
// Output structure:
//
//	media/output/<slug>/<slug>_<resolution>_<bitrate>.mp4
func Transcode(profile *TranscodeProfile, media *analyzer.MediaInfo) (*TranscodeResult, error) {
	// Validate input/output paths before proceeding
	if err := validatePaths(profile.InputPath, profile.OutputDir); err != nil {
		return nil, NewTranscoderError(
			"validation", "path_check", profile.InputPath, profile.OutputDir,
			"invalid input or output path", nil, 0, err,
		)
	}

	// Derive slug from input filename (strip extension, lowercase, alphanumeric assumed)
	baseName := filepath.Base(profile.InputPath)
	slug := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Construct slug-specific output directory: media/output/<slug>/
	slugDir := filepath.Join(profile.OutputDir, slug)

	// Ensure slug directory exists
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

	// Track which variants have already been processed to avoid duplicates
	seen := make(map[string]bool)

	// Iterate over each target resolution defined in the profile
	for _, res := range profile.Resolutions {
		// Construct output filename: <slug>_<resolution>_<bitrate>.mp4
		bitrate := profile.Bitrate[res]
		outputFilename := fmt.Sprintf("%s_%s_%skbps.mp4", slug, res, bitrate)
		outputPath := filepath.Join(slugDir, outputFilename)

		// Prevent duplicate variants by resolution + bitrate
		key := fmt.Sprintf("%s_%s", res, bitrate)
		if seen[key] {
			log.Printf("⚠️ Skipping duplicate variant: %s", key)
			continue
		}
		seen[key] = true

		// Build ffmpeg command for this variant
		cmd := buildFFmpegCommand(profile, res)
		// Replace final output path in command with our new slug-based path
		cmd[len(cmd)-1] = outputPath

		log.Printf("🎞️ Transcoding to %s: %s", res, strings.Join(cmd, " "))

		// Execute the ffmpeg command and capture error/exit code
		if err := executil.RunCommand(cmd); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, NewTranscoderError(
				"execution", "transcode", profile.InputPath, outputPath,
				"ffmpeg command failed", cmd, 1, err,
			))
			continue
		}

		// Lookup actual resolution dimensions from scaler
		width, height, err := scaler.DimensionsForLabel(res)
		if err != nil {
			log.Printf("⚠️ Unknown resolution label: %s - using source dimensions", res)
			width = media.Width
			height = media.Height
		}

		// Append successful variant metadata
		result.Variants = append(result.Variants, ResolutionVariant{
			Width:          width,
			Height:         height,
			Bitrate:        profile.Bitrate[res],
			ScaleFlag:      "auto",
			OutputFilename: outputFilename,
		})
	}

	return result, nil
}
