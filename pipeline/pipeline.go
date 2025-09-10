package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/manifester"
	"github.com/dotsoulja/dotgo-transcode/internal/scaler"
	"github.com/dotsoulja/dotgo-transcode/internal/segmenter"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/logging"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/thumbnailer"
)

type Config struct {
	ProfilePath   string
	StreamFormat  string // "hls" or "dash"
	ClientContext scaler.ClientContext
}

// Report captures the outcome of a full pipeline run.
// It includes, paths, counts, and any errors encountered.
type Report struct {
	InputPath     string
	ManifestPath  string
	VariantCount  int
	ManifestCount int
	Errors        []error
}

// Run executes the full pipeline and assumes a valid json/yaml profile located in /profiles directory.
// It returns a Report summarizing the process and any errors encountered.
func Run(config Config) (*Report, error) {
	var report Report
	logger := &logging.UnifiedLogger{}

	// Load transcode profile
	profile, err := transcoder.LoadProfile(config.ProfilePath)
	if err != nil {
		return nil, wrap("load profile", err)
	}
	report.InputPath = profile.InputPath

	// Analyze input media
	media, err := analyzer.AnalyzeMedia(profile.InputPath, logger)
	if err != nil {
		return nil, wrap("analyze media", err)
	}

	// Select resolution preset
	initialPreset, err := scaler.SelectPreset(media.Width, media.Height, &config.ClientContext)
	if err != nil {
		return nil, wrap("select preset", err)
	}
	_ = initialPreset // optional: log or use for override

	// Transcode media
	result, err := transcoder.Transcode(profile, media, logger)
	if err != nil {
		return nil, wrap("transcode", err)
	}
	report.VariantCount = len(result.Variants)
	for _, e := range result.Errors {
		report.Errors = append(report.Errors, e)
	}

	// Segment variants
	segResult, err := segmenter.SegmentMedia(result, config.StreamFormat, media)
	if err != nil {
		return nil, wrap("segment", err)
	}
	report.ManifestCount = len(segResult.Manifests)
	for _, e := range segResult.Errors {
		report.Errors = append(report.Errors, e)
	}

	// Generate thumbnails
	basename := filepath.Base(profile.InputPath)
	name := strings.TrimSuffix(basename, filepath.Ext(basename))
	if err := thumbnailer.GenerateThumbnails(*media, *result, name); err != nil {
		report.Errors = append(report.Errors, wrap("thumbnail", err))
	}

	// Generate master manifest
	manifestPath, err := manifester.GenerateMasterManifest(segResult, profile.PreserveManifest)
	if err != nil {
		return nil, wrap("manifest", err)
	}
	report.ManifestPath = manifestPath

	return &report, nil
}

// RunPipeline executes the full media pipeline using a provided TranscodeProfile.
// This function is designed for backend automation, allowing dynamic profile construction
// per movie slug or media asset. It performs the following steps.
//
//  1. Analyze media (duration, resolution, framerate, keyframes)
//  2. Transcode into resolution-bitrate variants
//  3. Segment each variant into HLS format (full DASH support coming soon)
//  4. Generate thumbnails for frontend scrubber (based on segment length)
//  5. Build master manifest referencing all variants (master.m3u8)
//
// In this version, the caller is responsible for constructing the TranscodeProfile with appropriate
// input/ output paths and variant ladder. This function returns a structured report
// for logging, retry logic, or frontend introspection.
func RunPipeline(profile *transcoder.TranscodeProfile) (*Report, error) {
	logger := &logging.UnifiedLogger{}
	report := &Report{InputPath: profile.InputPath}

	// Log profile summary before starting
	fmt.Println("\n🎬 Starting pipeline for:")
	fmt.Printf("   📂 InputPath:        %s\n", profile.InputPath)
	fmt.Printf("   📂 OutputDir:        %s\n", profile.OutputDir)
	fmt.Printf("   🎞️ VideoCodec:       %s\n", profile.VideoCodec)
	fmt.Printf("   🎵 AudioCodec:       %s\n", profile.AudioCodec)
	fmt.Printf("   📦 Container:        %s\n", profile.Container)
	fmt.Printf("   ⏰ SegmentLength:    %d\n", profile.SegmentLength)
	fmt.Printf("   🔧 PreserveManifest: %v\n", profile.PreserveManifest)
	fmt.Printf("   🏎️ UseHardwareAccel: %v\n", profile.UseHardwareAccel)

	fmt.Println("   🎯 Variants:")
	for i, v := range profile.Variants {
		fmt.Printf("      • [%d] %s @ %s\n", i, v.Resolution, v.Bitrate)
	}

	// Step 1: Analyze media file for metadata
	media, err := analyzer.AnalyzeMedia(profile.InputPath, logger)
	if err != nil {
		return nil, wrap("analyze media", err)
	}

	// Step 2: Transcode into resolution-bitrate variants
	result, err := transcoder.Transcode(profile, media, logger)
	if err != nil {
		return nil, wrap("transcode", err)
	}
	report.VariantCount = len(result.Variants)
	for _, e := range result.Errors {
		report.Errors = append(report.Errors, e)
	}

	// Step 3: Segment each variant into HLS format
	segResult, err := segmenter.SegmentMedia(result, "hls", media)
	if err != nil {
		return nil, wrap("segment", err)
	}
	report.ManifestCount = len(segResult.Manifests)
	for _, e := range segResult.Errors {
		report.Errors = append(report.Errors, e)
	}

	// Step 4: Generate thumbnails for scrubber
	name := strings.TrimSuffix(filepath.Base(profile.InputPath), filepath.Ext(profile.InputPath))
	if err := thumbnailer.GenerateThumbnails(*media, *result, name); err != nil {
		report.Errors = append(report.Errors, wrap("thumbnail", err))
	}

	// Step 5: Build master manifest referencing all variants
	manifestPath, err := manifester.GenerateMasterManifest(segResult, profile.PreserveManifest)
	if err != nil {
		return nil, wrap("manifest", err)
	}
	report.ManifestPath = manifestPath

	return report, nil

}

// wrap adds stage context to errors for structured logging and debugging.
// Used internally to annotate errors from each pipeline phase.
func wrap(stage string, err error) error {
	return fmt.Errorf("[%s] %v", stage, err)
}
