package transcoder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// validatePaths checks that input and output paths are accessible.
// Creates the output directory if it doesn't exist
func validatePaths(input, output string) error {
	if _, err := os.Stat(input); err != nil {
		return fmt.Errorf("input path invalid: %w", err)
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}
	return nil
}

// buildFFmpegCommand constructs the ffmpeg command for a given resolution.
// It generates a unique output filename based on input file name, resolution, and bitrate.
// Falls back to default bitrate if parsing fails.
func buildFFmpegCommand(profile *TranscodeProfile, res string) []string {
	// Extract base name from input file
	base := strings.TrimSuffix(filepath.Base(profile.InputPath), filepath.Ext(profile.InputPath))
	safeBase := strings.ReplaceAll(base, " ", "_")

	// Parse bitrate string to int
	bitrateStr := profile.Bitrate[res]
	bitrateInt := parseBitrateKbps(bitrateStr)

	if bitrateInt == 0 {
		log.Printf("⚠️ Bitrate parsing failed for resolution %s: %q. Using fallback bitrate.", res, bitrateStr)
		bitrateStr = "2000k"
		bitrateInt = 2000
	}

	outputFilename := fmt.Sprintf("%s_%s_%dkbps.%s", safeBase, res, bitrateInt, profile.Container)
	outputPath := filepath.Join(profile.OutputDir, outputFilename)

	log.Printf("🔧 Building ffmpeg command for %s (%dkbps)", res, bitrateInt)

	return []string{
		"ffmpeg",
		"-i", profile.InputPath,
		"-vf", fmt.Sprintf("scale=-2:%s", strings.TrimSuffix(res, "p")),
		"-c:v", profile.VideoCodec,
		"-b:v", bitrateStr,
		"-c:a", profile.AudioCodec,
		outputPath,
	}
}

// parseBitrateKbps converts a bitrate string like "3000k" to an integer in kbps.
// Returns 0 if parsing fails.
func parseBitrateKbps(bitrate string) int {
	bitrate = strings.ToLower(strings.TrimSpace(bitrate))
	bitrate = strings.TrimSuffix(bitrate, "k")
	if bitrate == "" {
		return 0
	}
	val, err := strconv.Atoi(bitrate)
	if err != nil {
		return 0
	}
	return val
}
