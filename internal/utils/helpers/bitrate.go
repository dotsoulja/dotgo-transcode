package helpers

import (
	"strconv"
	"strings"
)

// ParseBitrateKbps converts a bitrate string like "3000k" to an ineger in kbps.
// Returns 0 if parsing fails.
// Used by transcoder and segmenter for naming, filtering, and validation.
func ParseBitrateKbps(bitrate string) int {
	bitrate = strings.ToLower(strings.TrimSpace(bitrate))
	bitrate = strings.TrimSuffix(bitrate, "k")
	if bitrate == "" {
		return 0
	}
	val, err := strconv.Atoi(bitrate)
	if err != nil {
		return 0
	}
	return val
}
