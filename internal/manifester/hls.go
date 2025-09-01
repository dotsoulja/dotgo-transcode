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
		return 3000000
	case "480p":
		return 1500000
	case "360p":
		return 1000000
	case "240p":
		return 500000
	case "144p":
		return 150000
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
	case "144p":
		return "256x144"
	default:
		return "640x360"
	}
}

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
	for _, label := range order {
		if entry, ok := merged[label]; ok {
			sorted = append(sorted, entry)
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
