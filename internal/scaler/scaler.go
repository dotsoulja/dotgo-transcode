// Package scaler provides high-level orchestration for resolution scaling.
// This file exposes public entry points for selecting presets and applying scaling logic.
package scaler

import (
	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
)

// SelectResolutions filters resolution presets based on source media and client context.
// Applies bandwidth filtering, device preferences, and upscaling rules.
func SelectResolutions(media *analyzer.MediaInfo, ctx *ClientContext) ([]ResolutionPreset, error) {
	var selected []ResolutionPreset

	for _, preset := range StandardPresets {
		// Skip upscaling unless explicitly allowed
		if IsUpscale(media.Width, media.Height, preset.Width, preset.Height) && (ctx == nil || !ctx.PreferUpscale) {
			continue
		}

		// Skip low resolutions if client disallows them
		if ctx != nil && !ctx.AllowLowRes && preset.IsSD() {
			continue
		}

		// Skip if bandwidth is insufficient
		if ctx != nil && ctx.BandwidthKbps > 0 && preset.MinBitrate > ctx.BandwidthKbps {
			continue
		}

		selected = append(selected, preset)
	}

	if len(selected) == 0 {
		return nil, NewScalerError("SelectResolutions", "no suitable resolutions found")
	}

	return selected, nil
}
