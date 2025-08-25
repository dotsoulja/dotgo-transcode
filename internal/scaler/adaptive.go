package scaler

// AdjustResolution dynamically selects a resolution based on bandwidth and playback health.
// It uses failure count, bandwidth, and manual override to decide whether to drop or bump resolution.
func AdjustResolution(current ResolutionPreset, ctx ClientContext) ResolutionPreset {
	// Manual override takes precedence
	if ctx.ManualOverride != "" {
		for _, preset := range StandardPresets {
			if NormalizeLabel(preset.Label) == NormalizeLabel(ctx.ManualOverride) {
				return preset
			}
		}
	}

	if !ctx.AdaptiveEnabled {
		return current
	}

	// Drop resolution if failures exceed threshold
	if ctx.RecentFailures >= 3 {
		for i := len(StandardPresets) - 1; i >= 0; i-- {
			p := StandardPresets[i]
			if p.MinBitrate <= ctx.BandwidthKbps && p.Height < current.Height {
				return p
			}
		}
	}

	// Bump resolution if stable and bandwidth allows
	if ctx.RecentFailures == 0 {
		for _, p := range StandardPresets {
			if p.MinBitrate <= ctx.BandwidthKbps && p.Height > current.Height {
				return p
			}
		}
	}

	// No change
	return current
}
