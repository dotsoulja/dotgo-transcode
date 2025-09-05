package transcoder

// TranscodeProfile defines the parameters for a transcoding session.
// Parsed from a config file (JSON or YAML) and passed through the pipeline.
// Supports resolution-specific bitrates, codec/container choices, and optional hardware acceleration.

// Variant allows for multiple bitrate variants of the same resolution
type Variant struct {
	Resolution string `json:"resolution" yaml:"resolution"`
	Bitrate    string `json:"bitrate" yaml:"bitrate"`
}

type TranscodeProfile struct {
	InputPath        string    `json:"input_path" yaml:"input_path"`                                   // Path to source media file (e.g. "media/movie.mp4")
	OutputDir        string    `json:"output_dir" yaml:"output_dir"`                                   // Directory to write output files (e.g. "media/output/")
	Resolutions      []string  `json:"target_res" yaml:"target_res"`                                   // Target resolutions (e.g. ["1080p", "720p", "480p"])
	AudioCodec       string    `json:"audio_codec,omitempty" yaml:"audio_codec,omitempty"`             // Audio codec (e.g. "aac", "copy"); defaults to "aac"
	VideoCodec       string    `json:"video_codec" yaml:"video_codec"`                                 // Video codec (e.g. "h264", "vp9"); may be overridden for hardware acceleration
	Variants         []Variant `json:"variants" yaml:"variants"`                                       // Bitrate per resolution (e.g. {"720p": "3000k", "480p": "1500k"})
	SegmentLength    int       `json:"segment_length" yaml:"segment_length"`                           // Segment duration in seconds; used during segmentation phase
	Container        string    `json:"container" yaml:"container"`                                     // Output container format (e.g. "mp4", "mkv")
	UseHardwareAccel bool      `json:"use_hwaccel,omitempty" yaml:"use_hwaccel,omitempty"`             // Enable platform-specific hardware acceleration (e.g. VideoToolbox on macOS)
	PreserveManifest bool      `json:"preserve_manifest,omitempty" yaml:"preserve_manifest,omitempty"` // Merge new variants into existing master.m3u8
}
