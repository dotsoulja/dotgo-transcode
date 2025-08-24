package main

import (
	"fmt"

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
	}
}
