package transcoder

// TranscodeProfile defines the parameters for a transcoding session.
// It supports resolution-specific bitrates and modular codec/container choices,
// and optional hardware acceleration.
type TranscodeProfile struct {
	InputPath        string            `json:"input_path" yaml:"input_path"`                       // Path to source media
	OutputDir        string            `json:"output_dir" yaml:"output_dir"`                       // Where to write output
	Resolutions      []string          `json:"target_res" yaml:"target_res"`                       // e.g. ["1080p", "720p"]
	AudioCodec       string            `json:"audio_codec,omitempty" yaml:"audio_codec,omitempty"` // Optional: "aac", "copy"
	VideoCodec       string            `json:"video_codec" yaml:"video_codec"`                     // e.g. "h264", "vp9"
	Bitrate          map[string]string `json:"bitrate" yaml:"bitrate"`                             // e.g. {"1080p": "5000k"}
	SegmentLength    int               `json:"segment_length" yaml:"segment_length"`               // in seconds
	Container        string            `json:"container" yaml:"container"`                         // e.g. "mp4", "mkv"
	UseHardwareAccel bool              `json:"use_hwaccel,omitempty" yaml:"use_hwaccel,omitempty"` // Enable platform-specific hardware acceleration
}
