package analyzer

// ffprobeOutput represents the top-level structure returned by ffprobe
// when using -show_format and -show_streams with JSON output.
// This is used in AnalyzeMedia() to extract duration, bitrate, and stream metadata.
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"` // video/audio streams
	Format  ffprobeFormat   `json:"format"`  // container-level metadata
}

// ffprobeStream represents a single stream (video or audio) in ffprobe output
type ffprobeStream struct {
	CodecType  string `json:"codec_type"`             // "video" or "audio"
	CodecName  string `json:"codec_name"`             // e.g. "h264"
	Width      int    `json:"width,omitempty"`        // only for video
	Height     int    `json:"height,omitempty"`       // only for video
	BitRate    string `json:"bit_rate,omitempty"`     // e.g. "1000k"
	RFrameRate string `json:"r_frame_rate,omitempty"` // raw framerate string
}

// ffprobeFormat represents the container-level metadata
type ffprobeFormat struct {
	Duration string `json:"duration"` // in seconds
	BitRate  string `json:"bit_rate"` // in bits per second
}

// ffprobeFrame represents a single decoded frame, used for keyframe analysis.
type ffprobeFrame struct {
	KeyFrame int             `json:"key_frame"` // 1 if keyframe
	PTS      FlexibleFloat64 `json:"pts_time"`  // timestamp in seconds
}

// ffprobeFramesOutput wraps the list of frames returned by ffprobe.
type ffprobeFramesOutput struct {
	Frames []ffprobeFrame `json:"frames"`
}
