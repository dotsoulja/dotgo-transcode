// Package manifester provides HLS master playlist generation.
// This file builds a multi-variant .m3u8 referencing segmented streams.
package manifester

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
)

// generateHLSMaster creates a master .m3u8 playlist referencing all HLS variants.
// Each variant includes resolution and bitrate metadata for adaptive playback.
//
// Output:
//
//	media/output/<slug>/master.m3u8
//
// References:
//
//	<resolution>/<resolution>.m3u8
func generateHLSMaster(seg *segmenter.SegmentResult) (string, error) {
	masterPath := filepath.Join(seg.OutputDir, "master.m3u8")
	f, err := os.Create(masterPath)
	if err != nil {
		return "", NewManifesterError("write_file", "failed to create HLS master playlist", err)
	}
	defer f.Close()

	_, _ = f.WriteString("#EXTM3U\n")
	_, _ = f.WriteString("#EXT-X-VERSION:3\n")

	for _, manifest := range seg.Manifests {
		label := extractLabel(manifest)
		bitrate := estimateBitrate(label)
		res := resolutionFromLabel(label)

		// Reference manifest as <resolution>/<resolution>.m3u8
		uri := filepath.Join(label, filepath.Base(manifest))

		_, _ = f.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n%s\n",
			bitrate, res, uri,
		))
	}

	return masterPath, nil
}

// extractLabel pulls resolution label from manifest filename (e.g. "720p.m3u8" -> "720p")
func extractLabel(path string) string {
	base := filepath.Base(path)
	parts := strings.Split(base, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// estimateBitrate returns a rough bandwidth estimate in bits per second based on label.
func estimateBitrate(label string) int {
	switch strings.ToLower(label) {
	case "1080p":
		return 5000000
	case "720p":
		return 2500000
	case "480p":
		return 1000000
	case "240p":
		return 300000
	default:
		return 1000000
	}
}

// resolutionFromLabel returns resolution string like "1280x720" for HLS metadata.
func resolutionFromLabel(label string) string {
	switch strings.ToLower(label) {
	case "1080p":
		return "1920x1080"
	case "720p":
		return "1280x720"
	case "480p":
		return "854x480"
	case "360p":
		return "640x360"
	case "240p":
		return "426x240"
	default:
		return "640x360"
	}
}
