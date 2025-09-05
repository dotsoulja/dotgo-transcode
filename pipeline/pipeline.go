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

type Report struct {
	InputPath     string
	ManifestPath  string
	VariantCount  int
	ManifestCount int
	Errors        []error
}

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

func wrap(stage string, err error) error {
	return fmt.Errorf("[%s] %v", stage, err)
}
