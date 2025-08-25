package main

import (
	"fmt"
	"log"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

func main() {
	// Use a single high-quality movie and profile
	profileName := "sample_profile.json"
	inputMovie := "media/thelostboys.mp4"

	// Load profile
	profile, err := transcoder.LoadProfile(profileName)
	if err != nil {
		log.Fatalf("❌ Failed to load profile: %v", err)
	}
	profile.InputPath = inputMovie

	fmt.Println("\n🎬 Loaded TranscodeProfile:")
	fmt.Printf("   📁 InputPath:     %s\n", profile.InputPath)
	fmt.Printf("   📂 OutputDir:     %s\n", profile.OutputDir)
	fmt.Printf("   🎞️ VideoCodec:    %s\n", profile.VideoCodec)
	fmt.Printf("   🔊 AudioCodec:    %s\n", profile.AudioCodec)
	fmt.Printf("   📦 Container:     %s\n", profile.Container)
	fmt.Printf("   ⏱️ SegmentLength: %d\n", profile.SegmentLength)
	fmt.Printf("   📐 TargetRes:     %v\n", profile.Resolutions)
	fmt.Printf("   📊 Bitrate:       %v\n", profile.Bitrate)

	// Analyze media
	media, err := analyzer.AnalyzeMedia(profile.InputPath)
	if err != nil {
		log.Fatalf("❌ Failed to analyze media: %v", err)
	}
	fmt.Printf("\n🧠 MediaInfo: Duration=%.2fs, Width=%d, Height=%d, Bitrate=%dkbps\n",
		media.Duration, media.Width, media.Height, media.Bitrate)

	// Simulate client context
	ctx := scaler.ClientContext{
		DeviceType:      "desktop",
		BandwidthKbps:   6000, // Start strong
		PreferUpscale:   false,
		AllowLowRes:     true,
		AdaptiveEnabled: true,
	}

	// Initial resolution selection
	initialPreset, err := scaler.SelectPreset(media.Width, media.Height, &ctx)
	if err != nil {
		log.Fatalf("❌ Failed to select initial resolution: %v", err)
	}
	fmt.Printf("\n🚀 Initial resolution selected: %s\n", initialPreset.Preset.LabelWithDimensions())

	// Simulate playback drop
	ctx.BandwidthKbps = 1800
	ctx.RecentFailures = 4
	adjusted := scaler.AdjustResolution(initialPreset.Preset, ctx)
	fmt.Printf("📉 Bandwidth dropped. Adjusted resolution: %s\n", adjusted.LabelWithDimensions())

	// Simulate recovery
	ctx.BandwidthKbps = 6000
	ctx.RecentFailures = 0
	recovered := scaler.AdjustResolution(adjusted, ctx)
	fmt.Printf("📈 Network recovered. Resolution bumped back to: %s\n", recovered.LabelWithDimensions())

	// Transcode using recovered resolution
	fmt.Println("\n🎞️ Starting transcoding...")
	result, err := transcoder.Transcode(profile, media)
	if err != nil {
		log.Fatalf("❌ Transcoding failed: %v", err)
	}

	// Print result summary
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
}
