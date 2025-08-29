package analyzer

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

// extractFramerate runs ffprobe to retrieve the raw frame rate string (e.g. "30000/1001")
// from the primary video stream, then parses it into a float64 value.
// This is important for segment alignment and playback smoothness
func extractFramerate(path string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=r_frame_rate",
		"-of", "json",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, &AnalyzerError{
			Op:   "exec_ffprobe_framerate",
			Path: path,
			Err:  err,
		}
	}

	var result struct {
		Streams []struct {
			Rate string `json:"r_frame_rate"` // e.g. "30000/1001"
		}
	}

	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return 0, &AnalyzerError{
			Op:   "unmarshal_framerate",
			Path: path,
			Err:  err,
		}
	}

	if len(result.Streams) == 0 {
		return 0, &AnalyzerError{
			Op:   "missing_stream_framerate",
			Path: path,
			Err:  nil,
		}
	}

	fr, err := parseRatio(result.Streams[0].Rate)
	if err != nil {
		return 0, &AnalyzerError{
			Op:   "parse_framerate_ratio",
			Path: path,
			Err:  err,
		}
	}

	return fr, nil
}
