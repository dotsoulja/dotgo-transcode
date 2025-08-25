// Package segmenter defines core types used during media segmentation.
// These structs capture manifest paths, success flags, and error metadata.
package segmenter

// SegmentResult captures the outcome of a segmentaion operation.
// Includes manifest paths, output directory, format, and error records.
type SegmentResult struct {
	OutputDir string           // Directory where segments and manifests were written
	Format    string           // "hls" or "dash"
	Success   bool             // Overall success flag
	Manifests []string         // Paths to generated manifest files
	Errors    []SegmenterError // Detailed error records
}
