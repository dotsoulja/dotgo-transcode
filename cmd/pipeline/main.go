package main

import (
	"fmt"
	"log"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/manifester"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

func main() {
	profileName := "sample_profile.json"
	inputMovie := "media/thelostboys.mp4"
	streamFormat := "hls" // or "dash"

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

	// Client context (no simulation)
	ctx := scaler.ClientContext{
		DeviceType:      "desktop",
		BandwidthKbps:   6000,
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

	// Transcode
	fmt.Println("\n🎞️ Starting transcoding...")
	result, err := transcoder.Transcode(profile, media)
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

	// Segment
	fmt.Println("\n✂️ Starting segmentation...")
	segResult, err := segmenter.SegmentMedia(result, streamFormat)
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

	// Manifest
	fmt.Println("\n🧾 Generating master manifest...")
	manifestPath, err := manifester.GenerateMasterManifest(segResult)
	if err != nil {
		log.Fatalf("❌ Manifest generation failed: %v", err)
	}
	fmt.Printf("📜 Master manifest generated at: %s\n", manifestPath)

}
