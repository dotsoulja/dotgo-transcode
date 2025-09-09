// Package segmenter provides utility functions for command construction and manifest logic.
// This file contains helpers for building ffmpeg commands and determining manifest extension.
package segmenter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
)

// buildSegmentCommand constructs the ffmpeg command to segment a media file.
// Supports HLS and DASH formats and injects keyframe alignment logic when
// MediaInfo is available. This ensures ABR-safe segment boundaries.
//
// Parameters:
//     - inputPath: full path to input media file
//     - outputDir: directory to write segments and manifest
//     - manifestName: filename of the manifest (e.g. "720p.m3u8")
//     - format: "hls" or "dash"
//     - segmentLength: desired segment duration in seconds
//     - media: optional MediaInfo for keyframe-aware alignment

func buildSegmentCommand(
	inputPath, outputDir, manifestName, format string,
	segmentLength int, media *analyzer.MediaInfo,
) []string {
	segLen := fmt.Sprintf("%d", segmentLength)

	// Optional keyframe alignment expression
	var forceKeyframes []string
	if media != nil && media.KeyframeInterval > 0 {
		expr := fmt.Sprintf("expr:gte(t,n_forced*%.2f)", media.KeyframeInterval)
		forceKeyframes = []string{"-force_key_frames", expr}
	}
	switch strings.ToLower(format) {
	case "hls":
		cmd := []string{
			"ffmpeg",
			"-i", inputPath,
			"-c", "copy",
			"-f", "hls",
			"-hls_time", segLen,
			"-hls_playlist_type", "vod",
			"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		}
		// Append keyframe flags if present
		if len(forceKeyframes) > 0 {
			cmd = append(cmd, forceKeyframes...)
		}

		// Append output manifest path as final positional argument
		cmd = append(cmd, manifestName)

		return cmd

	case "dash":
		return append([]string{
			"ffmpeg",
			"-i", inputPath,
			"-c", "copy",
			"-f", "dash",
			"-seg_duration", segLen,
			"-use_timeline", "1",
			"-use_template", "1",
		}, append(forceKeyframes, filepath.Join(outputDir, manifestName))...)

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
