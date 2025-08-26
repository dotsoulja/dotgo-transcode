package analyzer

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"sync"
)

// AnalyzeMedia orchestrates metadata extraction for a given media file.
// It runs ffprobe to gather format and stream-level data, then delegates
// to specialized functions to extract framerate and keyframe information.
// Returns a fully populated MediaInfo struct or an AnalyzerError.
func AnalyzeMedia(path string) (*MediaInfo, error) {
	cmd := exec.Command(
		"ffprobe",
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

	var probe ffprobeOutput
	if err := json.Unmarshal(out.Bytes(), &probe); err != nil {
		return nil, &AnalyzerError{
			Op:   "unmarshal_ffprobe",
			Path: path,
			Err:  err,
		}
	}

	info := &MediaInfo{}

	// Parse duration from format section
	if d, err := parseFloat(probe.Format.Duration); err == nil {
		info.Duration = d
	} else {
		logDebug("parseFloat failed", probe.Format.Duration, err)
	}

	// Parse overall bitrate from format section
	if br, err := parseInt(probe.Format.BitRate); err == nil {
		info.Bitrate = br / 1000 // Convert to kbps
	} else {
		logDebug("parseInt failed", probe.Format.BitRate, err)
	}

	// Fallback: use highest stream-level bitrate if format bitrate is missing
	if info.Bitrate == 0 {
		for _, stream := range probe.Streams {
			if br, err := parseInt(stream.BitRate); err == nil && br > info.Bitrate {
				info.Bitrate = br / 1000
			}
		}
	}

	// Extract codec and resolution from video/audio streams
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

	// Extract framerate and keyframes concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(2)

	go func() {
		defer wg.Done()
		if fr, err := extractFramerate(path); err == nil {
			mu.Lock()
			info.Framerate = fr
			mu.Unlock()
		} else {
			logDebug("extractFramerate failed", "", err)
		}
	}()

	go func() {
		defer wg.Done()
		if kf, interval, err := extractKeyframes(path); err == nil {
			mu.Lock()
			info.Keyframes = kf
			info.KeyframeInterval = interval
			mu.Unlock()
		} else {
			logDebug("extractKeyframes failed", "", err)
		}
	}()

	wg.Wait()

	return info, nil
}

// AnalyzeMediaConcurrent is an alias for AnalyzeMedia with concurrency support.
// This function is retained for semantic clarity and future expansion.
func AnalyzeMediaConcurrent(path string) (*MediaInfo, error) {
	return AnalyzeMedia(path)
}
