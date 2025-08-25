package transcoder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// validatePaths checks that input and output paths are accessible.
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
func buildFFmpegCommand(profile *TranscodeProfile, res string) []string {
	// Extract base name from input file
	base := strings.TrimSuffix(filepath.Base(profile.InputPath), filepath.Ext(profile.InputPath))
	safeBase := strings.ReplaceAll(base, " ", "_")

	// Parse bitrate string to int
	bitrateStr := profile.Bitrate[res]
	bitrateInt, err := strconv.Atoi(bitrateStr)
	if err != nil {
		log.Printf("Invalid bitrate for resolution %s: %s", res, bitrateStr)
		bitrateInt = 0 // fallback to 0 if parsing fails
	}

	// Build unique output filename
	outputFilename := fmt.Sprintf("%s_%sp_%dkbps.%s", safeBase, res, bitrateInt/1000, profile.Container)
	outputPath := filepath.Join(profile.OutputDir, outputFilename)

	return []string{
		"ffmpeg",
		"-i", profile.InputPath,
		"-vf", fmt.Sprintf("scale=-2:%s", res),
		"-c:v", profile.VideoCodec,
		"-b:v", bitrateStr,
		"-c:a", profile.AudioCodec,
		outputPath,
	}
}

// runCommand executes a shell command using os/exec.
// This is a placeholder for real subprocess logic with stderr capture.
func runCommand(cmd []string) error {
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Stdout = nil
	execCmd.Stderr = nil
	return execCmd.Run()
}
