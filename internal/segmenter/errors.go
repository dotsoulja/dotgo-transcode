// Package segmenter defines custom error types used during media segmentation.
// These errors wrap operation context, variant info, and underlying causes
// to support forensic debugging and resilient fallback logic.
package segmenter

import (
	"fmt"
)

// SegmenterError wraps errors that occur during segmentation.
// Includes operation context and optional underlying error.
type SegmenterError struct {
	Op  string // e.g. "segment", "validate", "build_command"
	Msg string // Human-readable summary
	Err error  // Optional underlying error
}

// Error implements the error interface for SegmenterError
func (e *SegmenterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("segmenter error [%s]: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("segmenter error [%s]: %s", e.Op, e.Msg)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *SegmenterError) Unwrap() error {
	return e.Err
}

// NewSegmenterError creates a new SegmenterError with context
// This is the preferred constructor for wrapping segmentation errors.
func NewSegmenterError(op, msg string, err error) *SegmenterError {
	return &SegmenterError{
		Op:  op,
		Msg: msg,
		Err: err,
	}
}
