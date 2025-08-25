// Package scaler provides utility functions for resolution comparison and label normalization.
// This file contains helper functions used across presets, selectors, and scaling logic.
package scaler

import (
	"fmt"
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

// DimensionsForLabel returns the width and height for a given resolution label.
// Returns an error if the label is not found in StandardPresets.
func DimensionsForLabel(label string) (int, int, error) {
	norm := NormalizeLabel(label)
	for _, p := range StandardPresets {
		if NormalizeLabel(p.Label) == norm {
			return p.Width, p.Height, nil
		}
	}
	return 0, 0, fmt.Errorf("resolution label not found: %s", label)
}
