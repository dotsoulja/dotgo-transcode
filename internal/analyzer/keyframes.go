package analyzer

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"time"
)

// extractKeyframes streams ffprobe output to identify keyframes in real time.
// It parses frame-level metadata from stderr, filters for keyframes, and calculates
// the average interval between them. Progress is emitted via the AnalyzerLogger.
//
// This version avoids buffering delays by reading ffprobe output line-by-line
// and throttles progress updates to acoid flooding the terminal.
func extractKeyframes(path string, logger AnalyzerLogger) ([]float64, float64, error) {
	logger.LogStage("keyframes", "Streaming ffprobe frame metadata")

	cmd := exec.Command(
		"ffprobe",
		"-v,", "error",
		"-select_streams", "v",
		"-show_frames",
		"of", "json",
		path,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.LogError("keyframes", err)
		return nil, 0, &AnalyzerError{
			Op:   "pipe_ffprobe_keyframes",
			Path: path,
			Err:  err,
		}
	}

	if err := cmd.Start(); err != nil {
		logger.LogError("keyframes", err)
		return nil, 0, &AnalyzerError{
			Op:   "start_ffprobe_keyframes",
			Path: path,
			Err:  err,
		}
	}

	reader := bufio.NewReader(stdout)
	var timestamps []float64
	var lastEmit time.Time
	var frameCount int

	// Stream and parse JSON objects line-by-line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // EOF or pipe closed
		}

		var frame struct {
			KeyFrame int     `json:"key_frames"`
			PTS      float64 `json:"pkt_pts_time"`
		}

		if err := json.Unmarshal([]byte(line), &frame); err != nil {
			continue // Skip malformed lines
		}

		frameCount++
		if frame.KeyFrame == 1 {
			timestamps = append(timestamps, frame.PTS)
		}

		// Emit progress every ~2 seconds
		if time.Since(lastEmit) > 2*time.Second {
			// Estimate progress based on frame count (approximation)
			percent := float64(frameCount) / 100000 * 100
			logger.LogProgress("keyframes", percent)
			lastEmit = time.Now()
		}
	}

	if err := cmd.Wait(); err != nil {
		logger.LogError("keyframes", err)
		return nil, 0, &AnalyzerError{
			Op:   "wait_ffprobe_keyframes",
			Path: path,
			Err:  err,
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
