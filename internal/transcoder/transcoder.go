package transcoder

import (
	"log"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
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

	// Iterate over each target resolution defined in the profile
	for _, res := range profile.TargetRes {
		cmd := buildFFmpegCommand(profile, res)

		log.Printf("Transcoding to %s: %s", res, strings.Join(cmd, " "))

		// Execute the ffmpeg command and capture error/exit code
		if err := runCommand(cmd); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, NewTranscoderError(
				"execution", "transcode", profile.InputPath, profile.OutputDir,
				"ffmpeg command failed", cmd, 1, err,
			))
			continue
		}

		// Append successful variant metadata
		result.Variants = append(result.Variants, ResolutionVariant{
			Width:     media.Width,  // placeholder - can be scaled later
			Height:    media.Height, // placeholder - canbe scaled later
			Bitrate:   profile.Bitrate[res],
			ScaleFlag: "auto",
		})
	}

	return result, nil
}
