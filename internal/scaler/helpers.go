// Package scaler provides utility functions for resolution comparison and label normalization.
// This file contains helper functions used across presets, selectors, and scaling logic.
package scaler

import (
	"strings"
)

// NormalizeLabel ensures resolution labels are lowercase and trimmed.
// Useful for comparing user input or config values.
func NormalizeLabel(label string) string {
	return strings.ToLower(strings.TrimSpace(label))
}

// IsUpscale returns true if the target resolution is larger than the source.
func IsUpscale(sourceWidth, sourceHeight, targetWidth, targetHeight int) bool {
	return targetWidth > sourceWidth || targetHeight > sourceHeight
}

// AspectRatio returns the aspect ratio as a float64 (e.g. 16:9 = 1.777).
func AspectRatio(width, height int) float64 {
	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}
