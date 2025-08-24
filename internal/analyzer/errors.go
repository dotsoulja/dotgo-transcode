package analyzer

import "fmt"

// AnalyzerError represents an error during media analysis.
// Includes operation context and file path for forensic clarity
type AnalyzerError struct {
	Op   string // e.g. "exec_ffprobe", "unmarshal_ffprobe"
	Path string // media file path
	Err  error  // underlying error
}

func (e *AnalyzerError) Error() string {
	return fmt.Sprintf("analyzer error [%s] on %q: %v", e.Op, e.Path, e.Err)
}

func (e *AnalyzerError) Unwrap() error {
	return e.Err
}
