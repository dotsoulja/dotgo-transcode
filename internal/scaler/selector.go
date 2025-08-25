// Package scaler contains logic for selecting appropriate resolution presets.
// This implements the core selection algorithm based on source and context
package scaler

import (
	"fmt"
)

// SelectPreset chooses the best resolution preset based on source dimensions and client context.
// Returns a ScalingDecision or fallback to default preset if no match is found.
func SelectPreset(sourceWidth, sourceHeight int, ctx *ClientContext) (*ScalingDecision, error) {
	for _, preset := range StandardPresets {
		if IsUpscale(sourceWidth, sourceHeight, preset.Width, preset.Height) && (ctx == nil || !ctx.PreferUpscale) {
			continue
		}
		if ctx != nil && ctx.BandwidthKbps > 0 && preset.MinBitrate > ctx.BandwidthKbps {
			continue
		}
		reason := fmt.Sprintf("Selected %s based on source %dx%d", preset.Label, sourceWidth, sourceHeight)
		if ctx != nil {
			reason += fmt.Sprintf(" and client context %+v", ctx)
		}
		return &ScalingDecision{Preset: preset, Reason: reason}, nil
	}

	for _, preset := range StandardPresets {
		if preset.IsDefault {
			return &ScalingDecision{
				Preset: preset,
				Reason: "No suitable preset found, falling back to default",
			}, nil
		}
	}
	return nil, NewScalerError("SelectPreset", "no valid resolution preset found")
}
