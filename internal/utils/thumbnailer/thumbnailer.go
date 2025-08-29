package thumbnailer

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// GenerateThumbnails creates thumbnails for a given media slug using the highest
// resolution transcoded variant. It determines segment length based on profile
// config or keyframe interval, then generates thumbnails at regular intervals.
//
// This function assumes that transcoding has already completed and that the
// output directory contains the expected .mp4 files.
func GenerateThumbnails(media analyzer.MediaInfo, result transcoder.TranscodeResult, slug string) error {
	// Step 1: Determine effective segment length
	effectiveSegmentLength := result.Profile.SegmentLength
	if effectiveSegmentLength == 0 {
		if media.KeyframeInterval >= 3.0 {
			effectiveSegmentLength = int(media.KeyframeInterval)
		} else {
			effectiveSegmentLength = 4 // fallback default
			log.Printf("⚠️ Keyframe interval too short (%.2fs), using fallback segment length: %ds", media.KeyframeInterval, effectiveSegmentLength)
		}
	}

	// Step 2: Generate timestamps
	timestamps := GenerateTimestamps(media.Duration, effectiveSegmentLength)
	if len(timestamps) == 0 {
		log.Printf("🚫 No valid timestamps generated for slug: %s", slug)
		return nil
	}

	// Step 3: Locate highest resolution transcoded variant
	var bitrateStr string
	for _, v := range result.Variants {
		if v.Height == media.Height {
			bitrateStr = v.Bitrate
			break
		}
	}
	if bitrateStr == "" {
		return fmt.Errorf("no variant found matching source height: %dp", media.Height)
	}

	// Parse bitrate string like "5000k" into int kbps
	bitrateKbps, err := parseBitrateKbps(bitrateStr)
	if err != nil {
		return fmt.Errorf("invalid bitrate format: %s", bitrateStr)
	}

	variantPath, err := GetVariantPath(result.OutputDir, slug, media.Height, bitrateKbps)
	if err != nil {
		return fmt.Errorf("failed to locate variant for thumbnail generation: %w", err)
	}

	// Step 4: Prepare thumbnails directory
	thumbDir, err := EnsureThumbnailDir(result.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to prepare thumbnails directory: %w", err)
	}

	// Step 5: Generate thumbnails using ffmpeg
	for _, ts := range timestamps {
		filename := FormatTimestampFilename(ts)
		outputPath := filepath.Join(thumbDir, filename)

		cmd := exec.Command("ffmpeg",
			"-ss", fmt.Sprintf("%.2f", ts),
			"-i", variantPath,
			"-frames:v", "1",
			"-q:v", "2",
			"-y", outputPath,
		)

		if err := cmd.Run(); err != nil {
			log.Printf("❌ Failed to generate thumbnail at %.2fs for slug %s: %v", ts, slug, err)
		} else {
			log.Printf("✅ Thumbnail generated: %s", outputPath)
		}
	}

	return nil
}

// parseBitrateKbps converts a bitrate string like "5000k" to an int (5000)
func parseBitrateKbps(bitrate string) (int, error) {
	bitrate = strings.TrimSuffix(bitrate, "k")
	return strconv.Atoi(bitrate)
}
