package main

import (
	"fmt"
	"log"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

func main() {
	// List of sample profile filenames to test
	profiles := []string{
		"no_audio_codec.json",
		"no_audio_codec.yml",
		"sample_profile.json",
		"sample_profile.yaml",
	}

	for _, filename := range profiles {
		fmt.Printf("\nTesting profile: %s", filename)

		// Load profile from internal/transcoder
		profile, err := transcoder.LoadProfile(filename)
		if err != nil {
			fmt.Printf("Failed to load profile: %v\n", err)
			continue
		}

		// Print summary of loaded profile
		fmt.Println("\n✅ Loaded TranscodeProfile:")
		fmt.Printf("   📁 InputPath:     %s\n", profile.InputPath)
		fmt.Printf("   📂 OutputDir:     %s\n", profile.OutputDir)
		fmt.Printf("   🎞️ VideoCodec:    %s\n", profile.VideoCodec)
		fmt.Printf("   🔊 AudioCodec:    %s\n", profile.AudioCodec)
		fmt.Printf("   📦 Container:     %s\n", profile.Container)
		fmt.Printf("   ⏱️ SegmentLength: %d\n", profile.SegmentLength)
		fmt.Printf("   📐 TargetRes:     %v\n", profile.TargetRes)
		fmt.Printf("   📊 Bitrate:       %v\n", profile.Bitrate)

		// Analyze media
		media, err := analyzer.AnalyzeMedia(profile.InputPath)
		if err != nil {
			log.Printf("❌ Failed to analyze media: %v\n", err)
			continue
		}
		fmt.Printf("🧠 MediaInfo: Duration=%.2fs, Width=%d, Height=%d\n",
			media.Duration, media.Width, media.Height)

		// Transcode
		result, err := transcoder.Transcode(profile, media)
		if err != nil {
			log.Printf("❌ Transcoding failed: %v\n", err)
			continue
		}

		// Print result summary
		if result.Success {
			fmt.Printf("✅ Transcoding succeeded for %s\n", profile.InputPath)
			for _, variant := range result.Variants {
				fmt.Printf("   🎯 Variant: %dx%d @ %s\n",
					variant.Width, variant.Height, variant.Bitrate)
			}
		} else {
			fmt.Printf("⚠️ Transcoding completed with errors:\n")
			for _, e := range result.Errors {
				fmt.Printf("   ❌ [%s:%s] %s\n", e.Stage, e.Operation, e.Message)
			}
		}
	}
}
