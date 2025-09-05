package analyzer

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFloat converts a string to a float64, used for duration and timestamps.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// parseInt converts a string to an int, used for bitrate parsing
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseRatio converts a string like "3000/1001" into a float64
// Used for framerate parsing from ffprobe.
func parseRatio(s string) (float64, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid ratio format: %s", s)
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return 0, fmt.Errorf("failed to parse ratio: %s", s)
	}
	return num / den, nil
}
