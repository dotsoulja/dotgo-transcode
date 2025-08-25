// Package segmenter provides utility functions for command construction and manifest logic.
// This file contains helpers for building ffmpeg commands and determining manifest extension.
package segmenter

import (
	"path/filepath"
	"strings"
)

// buildSegmentCommand constructs the ffmpeg command to segment a media file.
// Supports HLS and DASH formats with browser-friendly flags and naming conventions.
func buildSegmentCommand(inputPath, outputDir, manifestName, format string) []string {
	switch strings.ToLower(format) {
	case "hls":
		return []string{
			"ffmpeg",
			"-i", inputPath,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
			filepath.Join(outputDir, manifestName),
		}
	case "dash":
		return []string{
			"ffmpeg",
			"-i", inputPath,
			"-c", "copy",
			"-f", "dash",
			"-seg_duration", "4",
			"-use_timeline", "1",
			"-use_template", "1",
			filepath.Join(outputDir, manifestName),
		}
	default:
		return []string{"echo", "unsupported format"}
	}
}

// manifestExtension returns the appropriate manifest file extension for a given format.
// e.g. "hls" -> "m3u8", "dash" -> "mpd"
func manifestExtension(format string) string {
	switch strings.ToLower(format) {
	case "hls":
		return "m3u8"
	case "dash":
		return "mpd"
	default:
		return "txt"
	}
}
