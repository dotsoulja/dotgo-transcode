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
// Accepts an AnalyzerLogger for structured, stage-aware logging.
//
// This version ensures that extractKeyframes receives accurate duration and framerate
// for progress estimation and interval calculation. Concurrency is used where safe,
// and data dependencies are respected to avoid deadlocks.
func AnalyzeMedia(path string, logger AnalyzerLogger) (*MediaInfo, error) {
	// Run ffprobe to extract format and stream metadata
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
	logger.LogStage("parse", "Parsing duration and bitrate")

	// Parse duration from format section
	if d, err := parseFloat(probe.Format.Duration); err == nil {
		info.Duration = d
	} else {
		logger.LogError("parse_duration", err)
	}

	// Parse overall bitrate from format section
	if br, err := parseInt(probe.Format.BitRate); err == nil {
		info.Bitrate = br / 1000 // Convert to kbps
	} else {
		logger.LogError("parse_bitrate", err)
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

	logger.LogStage("streams", "Extracted codec and resolution metadata")

	// Extract framerate first (required for keyframe progress estimation)
	var frWg sync.WaitGroup
	var mu sync.Mutex

	frWg.Add(1)
	go func() {
		defer frWg.Done()
		logger.LogStage("framerate", "Extracting framerate")
		if fr, err := extractFramerate(path); err == nil {
			mu.Lock()
			info.Framerate = fr
			mu.Unlock()
		} else {
			logger.LogError("framerate", err)
		}
	}()
	frWg.Wait()

	// Extract keyframes using known duration and framerate
	var kfWg sync.WaitGroup
	kfWg.Add(1)
	go func() {
		defer kfWg.Done()
		mu.Lock()
		duration := info.Duration
		framerate := info.Framerate
		mu.Unlock()

		if kf, interval, err := extractKeyframes(path, duration, framerate, logger); err == nil {
			mu.Lock()
			info.Keyframes = kf
			info.KeyframeInterval = interval
			mu.Unlock()
		} else {
			logger.LogError("keyframes", err)
		}
	}()
	kfWg.Wait()

	logger.LogStage("complete", "Media analysis complete")
	return info, nil
}

// AnalyzeMediaConcurrent is an alias for AnalyzeMedia with concurrency support.
// This function is retained for semantic clarity and future expansion.
func AnalyzeMediaConcurrent(path string, logger AnalyzerLogger) (*MediaInfo, error) {
	return AnalyzeMedia(path, logger)
}
