// Package manifester provides HLS master playlist generation.
// This file builds a multi-variant .m3u8 referencing segmented streams.
package manifester

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
//	<resolution_bitrate>/<resolution_bitrate>.m3u8
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

		// Reference manifest as <label>/<label>.m3u8
		uri := filepath.Join(label, fmt.Sprintf("%s.m3u8", label))

		_, _ = f.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n%s\n",
			bitrate, res, uri,
		))
	}

	return masterPath, nil
}

// extractLabel returns the base filename without extension.
// Example: "720p_3000kbps.m3u8" -> "720p_3000kbps"
func extractLabel(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// estimateBitrate parses bitrate from label suffix (e.g. "3000kbps") and returns bits per second.
// Falls back to default if parsing fails.
func estimateBitrate(label string) int {
	parts := strings.Split(label, "_")
	if len(parts) > 1 {
		bitrateStr := strings.TrimSuffix(parts[1], "kbps")
		if kbps, err := strconv.Atoi(bitrateStr); err == nil {
			return kbps * 1000
		}
	}
	return 1000000 // default 1Mbps
}

// resolutionFromLabel parses resolution from label prefix (e.g. "720p") and returns as "widthxheight" string.
func resolutionFromLabel(label string) string {
	parts := strings.Split(label, "_")
	if len(parts) > 0 {
		switch parts[0] {
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
		case "144p":
			return "256x144"
		}
	}
	return "640x360" // default
}

// reconcileHLSMaster merges existing and new manifests, preserving canonical order.
// Useful when adding new variants to an existing master.m3u8
func reconcileHLSMaster(seg *segmenter.SegmentResult) (string, error) {
	masterPath := filepath.Join(seg.OutputDir, "master.m3u8")

	// Read existing master .m3u8
	fmt.Println("🔄 Reconciling with existing master manifest...")
	existing, err := os.ReadFile(masterPath)
	if err != nil {
		return "", NewManifesterError(
			"read_file", "failed to read existing HLS master.m3u8", err,
		)
	}

	// Parse existing entries
	fmt.Printf("Raw entries: \n%s\n", string(existing))
	existingEntries := parseHLSManifest(string(existing))
	fmt.Println("Existing entries:", existingEntries)

	newEntries := make(map[string]ManifestMeta)
	for _, manifest := range seg.Manifests {
		label := extractLabel(manifest)
		newEntries[label] = ManifestMeta{
			Label:       label,
			Bitrate:     estimateBitrate(label),
			Resolution:  resolutionFromLabel(label),
			ManifestURL: filepath.Join(label, filepath.Base(manifest)),
		}
	}

	// Merge and deduplicate
	merged := make(map[string]ManifestMeta)
	for _, entry := range existingEntries {
		merged[entry.Label] = entry
	}
	for label, entry := range newEntries {
		merged[label] = entry // overwrite if exists
	}

	// Sort by canonical resolution order
	order := []string{"144p", "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p"}
	var sorted []ManifestMeta
	for _, res := range order {
		for label, entry := range merged {
			if strings.HasPrefix(label, res) {
				sorted = append(sorted, entry)
			}
		}
	}

	fmt.Printf("Reconciled entries: %v\n", sorted)
	// Write reconciled manifest
	f, err := os.Create(masterPath)
	if err != nil {
		return "", NewManifesterError(
			"write_file", "failed to write reconciled master.m3u8", err,
		)
	}
	defer f.Close()

	_, _ = f.WriteString("#EXTM3U\n")
	_, _ = f.WriteString("#EXT-X-VERSION:3\n")
	for _, entry := range sorted {
		_, _ = f.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n%s\n",
			entry.Bitrate, entry.Resolution, entry.ManifestURL,
		))
	}

	return masterPath, nil
}

// parseHLSManifest extracts ManifestMeta entries from raw master.m3u8 content.
// Used during reconciliation to preserve existing variants.
func parseHLSManifest(raw string) []ManifestMeta {
	lines := strings.Split(raw, "\n")
	var entries []ManifestMeta

	for i := 0; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "#EXT-X-STREAM-INF") {
			meta := ManifestMeta{}
			inf := lines[i]
			next := lines[i+1]

			_, err := fmt.Sscanf(inf, "#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s", &meta.Bitrate, &meta.Resolution)
			if err != nil {
				continue
			}

			meta.ManifestURL = next
			meta.Label = extractLabel(next)
			entries = append(entries, meta)
		}
	}
	return entries
}
