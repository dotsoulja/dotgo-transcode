package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/manifester"
	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/logging"
)

func main() {
	start := time.Now()
	slug := "thelostboys"
	inputDir := filepath.Join("media/output", slug)
	streamFormat := "hls"

	// Manually define the variants you want to segment
	variants := []transcoder.ResolutionVariant{
		{Width: 1920, Height: 1080, Bitrate: "8000k", OutputFilename: slug + "_1080p_8000kbps.mp4"},
		{Width: 1920, Height: 1080, Bitrate: "5000k", OutputFilename: slug + "_1080p_5000kbps.mp4"},
		{Width: 1280, Height: 720, Bitrate: "3000k", OutputFilename: slug + "_720p_3000kbps.mp4"},
		{Width: 854, Height: 480, Bitrate: "1500k", OutputFilename: slug + "_480p_1500kbps.mp4"},
		{Width: 640, Height: 360, Bitrate: "1000k", OutputFilename: slug + "_360p_1000kbps.mp4"},
		{Width: 426, Height: 240, Bitrate: "500k", OutputFilename: slug + "_240p_500kbps.mp4"},
		{Width: 256, Height: 144, Bitrate: "150k", OutputFilename: slug + "_144p_150kbps.mp4"},
	}

	// Load media info once
	logger := &logging.UnifiedLogger{}
	media, err := analyzer.AnalyzeMedia(filepath.Join(inputDir, variants[0].OutputFilename), logger)
	if err != nil {
		log.Fatalf("❌ Failed to analyze media: %v", err)
	}

	// Build a fake TranscodeResult to pass to segmenter
	result := &transcoder.TranscodeResult{
		InputPath: filepath.Join(inputDir, variants[0].OutputFilename),
		OutputDir: inputDir,
		Duration:  media.Duration,
		Success:   true,
		Variants:  variants,
		Profile: &transcoder.TranscodeProfile{
			SegmentLength: 4, // or whatever you want
		},
	}

	// Run segmentation
	fmt.Println("\n✂️ Segmenting existing variants...")
	segResult, err := segmenter.SegmentMedia(result, streamFormat, media)
	if err != nil {
		log.Fatalf("❌ Segmentation failed: %v", err)
	}
	for _, m := range segResult.Manifests {
		fmt.Printf("📄 %s\n", m)
	}

	// Generate master manifest
	fmt.Println("\n🧾 Generating master manifest...")
	manifestPath, err := manifester.GenerateMasterManifest(segResult, false)
	if err != nil {
		log.Fatalf("❌ Manifest generation failed: %v", err)
	}
	fmt.Printf("📜 Master manifest generated at: %s\n", manifestPath)

	fmt.Printf("\n✅ Done in %s\n", time.Since(start))
}
