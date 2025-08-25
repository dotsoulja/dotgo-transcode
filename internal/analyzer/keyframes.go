package analyzer

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

// extractKeyframes runs ffprobe to retrieve all video frames,
// filters for keyframes, and calculates the average interval between them.
// This is essential for segment alignment in adaptive streaming.
func extractKeyframes(path string) ([]float64, float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v",
		"-show_frames",
		"-of", "json",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, 0, &AnalyzerError{
			Op:   "exec_ffprobe_keyframes",
			Path: path,
			Err:  err,
		}
	}

	var result ffprobeFramesOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, 0, &AnalyzerError{
			Op:   "unmarshal_keyframes",
			Path: path,
			Err:  err,
		}
	}

	var timestamps []float64
	for _, frame := range result.Frames {
		if frame.KeyFrame == 1 {
			timestamps = append(timestamps, float64(frame.PTS))
		}
	}

	if len(timestamps) < 2 {
		return timestamps, 0, nil // Not enough keyframes to calculate interval
	}

	var total float64
	for i := 1; i < len(timestamps); i++ {
		total += timestamps[i] - timestamps[i-1]
	}
	avgInterval := total / float64(len(timestamps)-1)

	return timestamps, avgInterval, nil
}
