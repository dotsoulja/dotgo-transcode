// Package scaler provides ergonomic methods on core types.
// This file contains helper methods for ResolutionPreset and ScalingDecision.
package scaler

import (
	"fmt"
)

// LabelWithDimensions returns a formatted label like "720p (1280x720)".
func (r ResolutionPreset) LabelWithDimensions() string {
	return fmt.Sprintf("%s (%dx%d)", r.Label, r.Width, r.Height)
}

// IsSD returns true if the resolution is 480p or lower.
func (r ResolutionPreset) IsSD() bool {
	return r.Height <= 480
}

// IsHD returns true if the resolution is 720p or higher.
func (r ResolutionPreset) IsHD() bool {
	return r.Height >= 720
}

// Summary returns a human-readable summary of the scaling decision
func (d ScalingDecision) Summary() string {
	return fmt.Sprintf("Selected %s for playback. Reason: %s", d.Preset.LabelWithDimensions(), d.Reason)
}
