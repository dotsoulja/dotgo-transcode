// Package manifester generates master manifests for adaptive streaming.
// It supports HLS and DASH formats and builds multi-variant playlists
// referencing segmented outputs from the transcoder pipeline
package manifester

import (
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
)

// GenerateMasterManifest creates a multi-variant manifest for adaptive playback.
// It accepts a SegmentResult and writes a master playlist referencing all variants.
// Supports "hls" (.m3u8) and "dash" (.mpd) formats.
func GenerateMasterManifest(seg *segmenter.SegmentResult, preserve bool) (string, error) {
	if seg == nil || len(seg.Manifests) == 0 {
		return "", NewManifesterError("validate", "no manifests to aggregate", nil)
	}

	switch strings.ToLower(seg.Format) {
	case "hls":
		if preserve {
			return reconcileHLSMaster(seg)
		}
		return generateHLSMaster(seg)
	case "dash":
		return generateDASHMaster(seg)
	default:
		return "", NewManifesterError("validate", "unsupported format: "+seg.Format, nil)
	}
}
