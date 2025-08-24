package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// MediaInfo holds metadata extracted from a media file.
// This struct drives resolution logic, segment alignment, and codec decisions.
type MediaInfo struct {
	Width      int       // Video width in pixels
	Height     int       // Video height in pixels
	Duration   float64   // Duration in seconds
	AudioCodec string    // e.g. "aac"
	VideoCodec string    // e.g. "h264"
	Bitrate    int       // Overall bitrate in kbps
	Keyframes  []float64 // Optional: timestamps of keyframes
}

// AnalyzeMedia runs ffprobe on the given file and returns parsed MediaInfo
// Returns an AnalyzerError with operation context if anything fails.
func AnalyzeMedia(path string) (*MediaInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, &AnalyzerError{
			Op:   "exec_ffprobe",
			Path: path,
			Err:  err,
		}
	}

	var probe struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width,omitempty"`
			Height    int    `json:"height,omitempty"`
			BitRate   string `json:"bit_rate,omitempty"`
		}
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		}
	}

	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		return nil, &AnalyzerError{
			Op:   "unmarshal_ffprobe",
			Path: path,
			Err:  err,
		}
	}

	info := &MediaInfo{}

	// Parse duration
	if d, err := parseFloat(probe.Format.Duration); err == nil {
		info.Duration = d
	} else {
		logDebug("parseFloat failed", probe.Format.Duration, err)
	}

	// Parse bitrate from format
	if br, err := parseInt(probe.Format.BitRate); err == nil {
		info.Bitrate = br / 1000 // convert to kbps
	} else {
		logDebug("parseInt failed", probe.Format.BitRate, err)
	}

	// Fallback: use highest stream bitrate if format bitrate is missing
	if info.Bitrate == 0 {
		for _, stream := range probe.Streams {
			if br, err := parseInt(stream.BitRate); err == nil && br > info.Bitrate {
				info.Bitrate = br / 1000
			}
		}
	}

	// Parse streams
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			info.VideoCodec = stream.CodecName
			info.Width = stream.Width
			info.Height = stream.Height
		case "audio":
			info.AudioCodec = stream.CodecName
		}
	}

	// TODO: Extract keyframes via -select_streams v -show_frames and parse pts_time
	info.Keyframes = []float64{} // placeholder for future logic

	return info, nil
}

// parseFloat safely parses a string to a float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// parseInt safely parses a string to an int
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// logDebug writes debug info to stderr (optional stub for forensic tracing)
func logDebug(msg string, value string, err error) {
	fmt.Fprintf(os.Stderr, "[debug] %s: value=%q err=%v\n", msg, value, err)
}
