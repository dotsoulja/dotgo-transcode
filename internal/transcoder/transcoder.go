package transcoder

import (
	"fmt"
	"log"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
)

// Transcode orchestrates the transcoding process for a given media file.
// It accepts a validated TranscodeProfile and extracted MediaInfo,
// then iterates through each target resolution, builds ffmpeg commands,
// executes them, and returns a TranscodeResult with success/ failure metadata.
//
// This function does not segment or mux; it focuses on resolution variants.
// Segmenting and manifest generation are handled in later phases.
func Transcode(profile *TranscodeProfile, media *analyzer.MediaInfo) (*TranscodeResult, error) {
	// Validate input/output paths before proceeding
	if err := validatePaths(profile.InputPath, profile.OutputDir); err != nil {
		return nil, NewTranscoderError(
			"validation", "path_check", profile.InputPath, profile.OutputDir,
			"invalid input or output path", nil, 0, err,
		)
	}

	result := &TranscodeResult{
		InputPath: profile.InputPath,
		OutputDir: profile.OutputDir,
		Duration:  media.Duration,
		Success:   true,
	}

	// Track which variants have already been processed to avoid duplicates
	seen := make(map[string]bool)

	// Iterate over each target resolution defined in the profile
	for _, res := range profile.Resolutions {
		cmd := buildFFmpegCommand(profile, res)

		// Extract output filename from command (last arg)
		outputPath := cmd[len(cmd)-1]
		filename := outputPath[strings.LastIndex(outputPath, "/")+1:]

		// Prevent duplicate variants by resolution + bitrate
		key := fmt.Sprintf("%s_%s", res, profile.Bitrate[res])
		if seen[key] {
			log.Printf("⚠️ Skipping duplicate variant: %s", key)
			continue
		}
		seen[key] = true

		log.Printf("🎞️ Transcoding to %s: %s", res, strings.Join(cmd, " "))

		// Execute the ffmpeg command and capture error/exit code
		if err := runCommand(cmd); err != nil {
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
			OutputFilename: filename,
		})
	}

	return result, nil
}
