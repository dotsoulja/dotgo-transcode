// Package scaler defines core types used for resolution scaling decisions.
// This file contains the primary data structures used across the scaler package.
package scaler

// ResolutionPreset defines a target resolution and its associated metadata.
// These presets are used to guide scaling decisions and adaptive streaming logic.
type ResolutionPreset struct {
	Width      int    // Horizontal resolution in pixels (e.g. 1920)
	Height     int    // Vertical resolution in pixels (e.g. 1080)
	Label      string // Human-readable label (e.g. "1080p", "720p")
	MinBitrate int    // Minimum recommended bitrate in kbps for this resolution
	IsDefault  bool   // Indicates if this preset is the default fallback
}

// ScalingDecision represents the outcome of a resolution selection process.
// It includes the chosen preset and any notes about why it was selected.
type ScalingDecision struct {
	Preset ResolutionPreset // The selected resolution preset
	Reason string           // Explanation/ rationale for the selection
}
