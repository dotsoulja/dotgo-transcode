package analyzer

import (
	"bufio"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// extractKeyframes streams ffprobe output to identify keyframes in real time.
// It parses frame-level metadata from compact output, filters for keyframes,
// and calculates the average interval between them. Progress is emitted via the AnalyzerLogger.
//
// This version avoids buffering delays by reading ffprobe output line-by-line,
// uses actual duration and framerate to estimate total frames, and throttles progress
// updates based on frame count to avoid flooding the terminal. It also logs every keyframe
// detection attempt and exposes silent failures in timestamp parsing.
func extractKeyframes(path string, duration, framerate float64, logger AnalyzerLogger) ([]float64, float64, error) {
	logger.LogStage("keyframes", "Streaming ffprobe frame metadata")

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v",
		"-show_entries", "frame=pts_time,key_frame",
		"-of", "compact",
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
	var frameCount int

	// Estimate total frames using duration × framerate
	estimatedTotalFrames := int(duration * framerate)
	log.Printf("Estimated total frames : %d, by using duration %d and framerate %d", estimatedTotalFrames, int(duration), int(framerate))
	const emitEveryNFrames = 5000 // Throttle progress updates

	// Stream and parse compact frame lines
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // EOF or pipe closed
		}

		frameCount++ // ✅ Count every frame

		// Parse keyframe flag and timestamp
		var isKeyframe bool
		var ts *float64

		for part := range strings.SplitSeq(line, "|") {
			if part == "key_frame=1" {
				isKeyframe = true
			}

			if val, ok := strings.CutPrefix(part, "pts_time="); ok {
				val = strings.Trim(val, "|\n\r ")
				parsed, err := strconv.ParseFloat(val, 64)
				if err == nil {
					ts = &parsed
				} else {
					log.Printf("⚠️ Failed to parse pts_time '%s' in line: %s", val, strings.TrimSpace(line))
				}
			}
		}

		if isKeyframe {
			if ts != nil {
				timestamps = append(timestamps, *ts)
			} else {
				log.Printf("⚠️ Keyframe detected but missing pts_time: %s", strings.TrimSpace(line))
			}
		}

		// Emit progress every N frames
		if frameCount%emitEveryNFrames == 0 && estimatedTotalFrames > 0 {
			percent := float64(frameCount) / float64(estimatedTotalFrames) * 100
			logger.LogProgress("keyframes", percent)
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

	log.Printf("🧮 Parsed %d frames, found %d keyframes", frameCount, len(timestamps))

	// Fallback if too few keyframes found
	if frameCount > 5000 && len(timestamps) < 2 {
		logger.LogStage("keyframes", "⚠️ Parsed over 5000 frames but found less than 2 keyframes — skipping interval calculation")
		return timestamps, 0, nil
	}

	if len(timestamps) < 2 {
		logger.LogStage("keyframes", "Not enough keyframes found to calculate interval")
		return timestamps, 0, nil
	}

	// Calculate average interval between keyframes
	var total float64
	for i := 1; i < len(timestamps); i++ {
		total += timestamps[i] - timestamps[i-1]
	}
	avgInterval := total / float64(len(timestamps)-1)

	logger.LogStage("keyframes", "✅ Keyframe extraction complete")
	return timestamps, avgInterval, nil
}
