package transcoder

// ResolutionVariant represents a single output resolution and its settings.
// Used to track successful transcodes and feed into segmentation and manifest generation.
type ResolutionVariant struct {
	Width          int    // Output width in pixels (e.g. 1280)
	Height         int    // Output height in pixels (e.g. 720)
	Bitrate        string // Target bitrate string (e.g. "1500k")
	ScaleFlag      string // Scaling behavior: "auto", "force", "skip"
	OutputFilename string // Final output filename (e.g. "video_720p_1500kbps.mp4")
}

// TranscodeResult captures the outcome of a transcoding operation.
// Includes input/output paths, duration, success flag, and a slice of
// ResolutionVariant for each successfully generated output.
// Errors are tracked with full forensic detail for debugging and logging.
type TranscodeResult struct {
	InputPath string              // Original input file path (e.g. "media/movie.mp4")
	OutputDir string              // Directory where outputs were written (e.g. "media/output/movie/")
	Duration  float64             // Duration of input media in seconds
	Success   bool                // Overall success flag (false if any variant failed)
	Variants  []ResolutionVariant // Successfully transcoded variants
	Profile   *TranscodeProfile   // Profile used for transcoding (includes codec, bitrate, etc.)
	Errors    []*TranscoderError  // Detailed error records (stage, command, exit code, etc.)
}
