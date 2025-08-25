package analyzer

// MediaInfo holds all extracted metadata about a media file.
// This struct is the foundation for resolution scaling, segment alignment,
// codec decisions, and adaptive streaming logic.
type MediaInfo struct {
	Width            int       // Video width in pixels
	Height           int       // Video height in pixels
	Duration         float64   // Total duration in seconds
	AudioCodec       string    // Audio codec used (e.g. "aac")
	VideoCodec       string    // Video codec used (e.g. "h264")
	Bitrate          int       // Overall bitrate in kbps
	Framerate        float64   // Frames per second (parsed from r_frame_rate)
	KeyframeInterval float64   // Average seconds between keyframes
	Keyframes        []float64 // Timestamps of keyframes in seconds
}
