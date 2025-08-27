package transcoder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// validatePaths checks that input and output paths are accessible.
// Creates the output directory if it doesn't exist.
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
// Injects hardware acceleration flags if enabled and platform supports it.
// Final output path is injected as the last argument.
func buildFFmpegCommand(profile *TranscodeProfile, res string) []string {
	// Sanitize input filename for output naming
	base := strings.TrimSuffix(filepath.Base(profile.InputPath), filepath.Ext(profile.InputPath))
	safeBase := strings.ReplaceAll(base, " ", "_")

	// Parse bitrate string (e.g. "3000k") into integer
	bitrateStr := profile.Bitrate[res]
	bitrateInt := parseBitrateKbps(bitrateStr)
	if bitrateInt == 0 {
		log.Printf("⚠️ Bitrate parsing failed for resolution %s: %q. Using fallback bitrate.", res, bitrateStr)
		bitrateStr = "2000k"
		bitrateInt = 2000
	}

	// Construct output filename and path
	outputFilename := fmt.Sprintf("%s_%s_%dkbps.%s", safeBase, res, bitrateInt, profile.Container)
	outputPath := filepath.Join(profile.OutputDir, outputFilename)

	// Determine video codec, optionally override for hardware acceleration
	videoCodec := profile.VideoCodec
	if profile.UseHardwareAccel && isMacOS() && strings.EqualFold(videoCodec, "h264") {
		videoCodec = "h264_videotoolbox"
		log.Printf("🍎 Using VideoToolbox hardware acceleration for %s", res)
	}

	// Build ffmpeg command with scale filter and codec settings
	return []string{
		"ffmpeg",
		"-i", profile.InputPath,
		"-vf", fmt.Sprintf("scale=-2:%s", strings.TrimSuffix(res, "p")), // height-driven scaling
		"-c:v", videoCodec,
		"-b:v", bitrateStr,
		"-c:a", profile.AudioCodec,
		"-reset_timestamps", "1",
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

// isMacOS returns true if the current platform is macOS.
// Used to conditionally enable VideoToolbox acceleration.
func isMacOS() bool {
	return strings.Contains(strings.ToLower(runtime.GOOS), "darwin")
}
