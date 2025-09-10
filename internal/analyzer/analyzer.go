package analyzer

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"sync"
)

// AnalyzeMedia extracts metadata from a media file using ffprobe.
// It parses duration, bitrate, codec, resolution, and optionally framerate and keyframes.
// This function is concurrency-safe and logs progress via the provided AnalyzerLogger.
//
// Behavior:
//   - If segmentLength == 0 -> keyframes are extracted to calculate segment intervals.
//   - If segmentLength > 0 -> keyframe extraction is skipped to save time.
//
// Parameters:
//   - path: full path to the media file (e.g. "movies/thelostboys/thelostboys.mp4")
//   - segmentLength: segment duration in seconds, if > 0 keyframes are skipped
//   - logger: structured logger for stage-aware progress and error reporting
//
// Returns:
//   - MediaInfo: populated metadata struct
//   - error: if any subprocess or parsing fails
func AnalyzeMedia(path string, segmentLength int, logger AnalyzerLogger) (*MediaInfo, error) {
	// Run ffprobe to extract format and stream-level metadata
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

	// Initialize MediaInfo and parse basic metadata
	info := &MediaInfo{}
	logger.LogStage("parse", "Parsing duration and bitrate") // basic data valuable for frontend consumption (e.g. seek/ scrub bar)

	if d, err := parseFloat(probe.Format.Duration); err == nil {
		info.Duration = d
	} else {
		logger.LogError("parse_duration", err)
	}

	if br, err := parseInt(probe.Format.BitRate); err == nil {
		info.Bitrate = br / 1000 // convert to kbps
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

	// Extract framerate (required for keyframe estimation)
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

	// Conditionally extract keyframes (only if segmentLength == 0)
	if segmentLength == 0 {
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
	} else {
		logger.LogStage("keyframes", "⏩ Skipping keyframe analysis (segment length manually set)")
	}

	logger.LogStage("complete", "✅ Media analysis complete")
	return info, nil
}

// AnalyzeMediaConcurrent is an alias for AnalyzeMedia.
// Retained for semantic clarity and future expansion.
func AnalyzeMediaConcurrent(path string, segmentLength int, logger AnalyzerLogger) (*MediaInfo, error) {
	return AnalyzeMedia(path, segmentLength, logger)
}
