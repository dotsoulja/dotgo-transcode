package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/manifester"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/logging"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/thumbnailer"
)

func main() {
	start := time.Now()
	logger := &logging.UnifiedLogger{}

	profileName := "sample_profile.json"
	streamFormat := "hls" // or "dash"

	// Load transcode profile
	profile, err := transcoder.LoadProfile(profileName)
	if err != nil {
		log.Fatalf("❌ Failed to load profile: %v", err)
	}

	fmt.Println("\n🎬 Loaded TranscodeProfile:")
	fmt.Printf("   📁 InputPath:        %s\n", profile.InputPath)
	fmt.Printf("   📂 OutputDir:        %s\n", profile.OutputDir)
	fmt.Printf("   🎞️ VideoCodec:       %s\n", profile.VideoCodec)
	fmt.Printf("   🔊 AudioCodec:       %s\n", profile.AudioCodec)
	fmt.Printf("   📦 Container:        %s\n", profile.Container)
	fmt.Printf("   ⏱️ SegmentLength:    %d\n", profile.SegmentLength)
	fmt.Printf("   🔧 PreserveManifest: %v\n", profile.PreserveManifest)

	fmt.Println("   🎯 Variants:")
	for i, v := range profile.Variants {
		fmt.Printf("    • [%d] %s @ %s\n", i, v.Resolution, v.Bitrate)
	}

	// Analyze input media once (shared across pipeline)
	media, err := analyzer.AnalyzeMedia(profile.InputPath, profile.SegmentLength, logger)
	if err != nil {
		log.Fatalf("❌ Failed to analyze media: %v", err)
	}
	fmt.Printf("\n🧠 MediaInfo: Duration=%.2fs, Width=%d, Height=%d, Bitrate=%dkbps\n",
		media.Duration, media.Width, media.Height, media.Bitrate)

	// Define client context for resolution selection
	ctx := scaler.ClientContext{
		DeviceType:      "desktop",
		BandwidthKbps:   6000,
		PreferUpscale:   false,
		AllowLowRes:     true,
		AdaptiveEnabled: true,
	}

	// Select initial resolution preset based on media and context
	initialPreset, err := scaler.SelectPreset(media.Width, media.Height, &ctx)
	if err != nil {
		log.Fatalf("❌ Failed to select initial resolution: %v", err)
	}
	fmt.Printf("\n🚀 Initial resolution selected: %s\n", initialPreset.Preset.LabelWithDimensions())

	// Transcode media into adaptive variants
	fmt.Println("\n🎞️ Starting transcoding...")
	result, err := transcoder.Transcode(profile, media, logger)
	if err != nil {
		log.Fatalf("❌ Transcoding failed: %v", err)
	}

	if result.Success {
		fmt.Printf("✅ Transcoding succeeded for %s\n", profile.InputPath)
		for _, variant := range result.Variants {
			fmt.Printf("   🎯 Variant: %dx%d @ %s\n", variant.Width, variant.Height, variant.Bitrate)
		}
	} else {
		fmt.Println("⚠️ Transcoding completed with errors:")
		for _, e := range result.Errors {
			fmt.Printf("   ❌ [%s:%s] %s\n", e.Stage, e.Operation, e.Message)
		}
	}

	// Segment each variant using shared MediaInfo
	fmt.Println("\n✂️ Starting segmentation...")
	segResult, err := segmenter.SegmentMedia(result, streamFormat, media)
	if err != nil {
		log.Fatalf("❌ Segmentation failed: %v", err)
	}
	if segResult.Success {
		fmt.Printf("✅ Segmentation succeeded. Manifests:\n")
		for _, m := range segResult.Manifests {
			fmt.Printf("   📄 %s\n", m)
		}
	} else {
		fmt.Println("⚠️ Segmentation completed with errors:")
		for _, e := range segResult.Errors {
			fmt.Printf("   ❌ [%s] %s\n", e.Op, e.Msg)
		}
	}

	// 🖼️ Generating thumbnails...
	fmt.Println("\n🖼️ Generating thumbnails...")
	basename := filepath.Base(profile.InputPath)                 // "thelostboys.mp4"
	name := strings.TrimSuffix(basename, filepath.Ext(basename)) // "thelostboys"

	if err := thumbnailer.GenerateThumbnails(*media, *result, name); err != nil {
		log.Printf("❌ Thumbnail generation failed: %v", err)
	}

	// Generate master manifest from segmented variants
	fmt.Println("\n🧾 Generating master manifest...")
	manifestPath, err := manifester.GenerateMasterManifest(segResult, profile.PreserveManifest)
	if err != nil {
		log.Fatalf("❌ Manifest generation failed: %v", err)
	}
	fmt.Printf("📜 Master manifest generated at: %s\n", manifestPath)

	// Final summary
	fmt.Println("\n📦 Final Report")
	fmt.Printf("   🎞️ Input: %s\n", profile.InputPath)
	fmt.Printf("   📐 Variants: %d\n", len(result.Variants))
	fmt.Printf("   📄 Manifests: %d\n", len(segResult.Manifests))
	fmt.Printf("   ⚠️ Errors: %d\n", len(result.Errors)+len(segResult.Errors))
	fmt.Printf("   🕒 Total pipeline time: %s\n", time.Since(start))
}
