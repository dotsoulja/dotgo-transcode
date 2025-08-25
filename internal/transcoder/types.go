package transcoder

// ResolutionVariant represents a single output resolution and its settings
type ResolutionVariant struct {
	Width     int
	Height    int
	Bitrate   string
	ScaleFlag string // e.g. "force", "skip", "auto"
}

// TranscodeResult holds metadata about a completed transcoding session.
type TranscodeResult struct {
	InputPath string
	OutputDir string
	Variants  []ResolutionVariant
	Duration  float64
	Success   bool
	Errors    []*TranscoderError
}
