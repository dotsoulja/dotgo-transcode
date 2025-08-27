package transcoder

// ResolutionVariant represents a single output resolution and its settings
type ResolutionVariant struct {
	Width          int    // Output width in pixels
	Height         int    // Output height in pixels
	Bitrate        string // Target bitrate (e.g. "1500k")
	ScaleFlag      string // e.g. "force", "skip", "auto"
	OutputFilename string // Final output filename (e.g. "video_720p.mp4")
}

// TranscodeResult captures the outcome of a transcoding operation.
// It includes input/output paths, duration, success flag, and a slice of
// ResolutionVariant for each successfully generated output.
// Errors are tracked with full forensic detail for debugging and logging.
type TranscodeResult struct {
	InputPath string              // Original input file path
	OutputDir string              // Directory where outputs were written
	Duration  float64             // Duration of input media in seconds
	Success   bool                // Overall success flag
	Variants  []ResolutionVariant // Successfully transcoded variants
	Profile   *TranscodeProfile   // Profile used for transcoding
	Errors    []*TranscoderError  // Detailed error records (if any)
}
