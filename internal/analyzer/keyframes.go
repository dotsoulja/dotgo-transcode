package analyzer

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

// extractKeyframes runs ffprobe to retrieve all video frames,
// filters for keyframes, and calculates the average interval between them.
// This is essential for segment alignment in adaptive streaming.
// Accepts an AnalyzerLogger for structured logging and optional progress tracking.
func extractKeyframes(path string, logger AnalyzerLogger) ([]float64, float64, error) {
	logger.LogStage("keyframes", "Running ffprobe to extract frame-level metadata")
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
		logger.LogError("keyframes", err)
		return nil, 0, &AnalyzerError{
			Op:   "exec_ffprobe_keyframes",
			Path: path,
			Err:  err,
		}
	}

	var result ffprobeFramesOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		logger.LogError("keyframes", err)
		return nil, 0, &AnalyzerError{
			Op:   "unmarshal_keyframes",
			Path: path,
			Err:  err,
		}
	}
	logger.LogStage("keyframes", "Filtering keyframes and calculating intervals")

	var timestamps []float64
	totalFrames := len(result.Frames)

	for i, frame := range result.Frames {
		if frame.KeyFrame == 1 {
			timestamps = append(timestamps, float64(frame.PTS))
		}

		// Emit progress every 1000 frames for long videos
		if i > 0 && i%1000 == 0 {
			percent := float64(i) / float64(totalFrames) * 100
			logger.LogProgress("keyframes", percent)
		}
	}

	if len(timestamps) < 2 {
		logger.LogStage("keyframes", "Not enough keyframes found to calculate interval")
		return timestamps, 0, nil
	}

	var total float64
	for i := 1; i < len(timestamps); i++ {
		total += timestamps[i] - timestamps[i-1]
	}
	avgInterval := total / float64(len(timestamps)-1)

	logger.LogStage("keyframes", "Keyframe extraction complete")

	return timestamps, avgInterval, nil
}
