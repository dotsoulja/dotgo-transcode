// Package scaler contains logic for selecting appropriate resolution presets.
// This implements the core selection algorithm based on source and context
package scaler

import (
	"fmt"
)

// SelectPreset chooses the best resolution preset based on source dimensions
// and optional client context (e.g. bandwidth, device type).
//
// Returns a ScalingDecision or an error if no suitable preset is found.
func SelectPreset(sourceWidth, sourceHeight int, ctx *ClientContext) (*ScalingDecision, error) {
	for _, preset := range StandardPresets {
		// Skip presets larger than the source unless upscaling is preferred
		if preset.Width > sourceWidth || preset.Height > sourceHeight {
			if ctx == nil || !ctx.PreferUpscale {
				continue
			}
		}

		// If bandwidth is provided, skip presets that exceed it
		if ctx != nil && ctx.BandwidthKbps > 0 && preset.MinBitrate > ctx.BandwidthKbps {
			continue
		}

		reason := fmt.Sprintf("Selected %s based on source %dx%d", preset.Label, sourceWidth, sourceHeight)
		if ctx != nil {
			reason += fmt.Sprintf(" and client context %+v", ctx)
		}

		return &ScalingDecision{
			Preset: preset,
			Reason: reason,
		}, nil
	}

	// If no match found, fallback to default preset
	for _, preset := range StandardPresets {
		if preset.IsDefault {
			return &ScalingDecision{
				Preset: preset,
				Reason: "No suitable preset found; falling back to default preset",
			}, nil
		}
	}

	return nil, NewScalerError("SelectPreset", "no valid resolution preset found")
}
