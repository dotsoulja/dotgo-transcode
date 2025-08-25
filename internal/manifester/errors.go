// Package manifester defines custom error types used during manifest generation.
// These errors wrap operation context and underlying causes for forensic clarity.
package manifester

import (
	"fmt"
)

// ManifesterError wraps errors that occur during manifest generation.
// Includes operation context and optional underlying error.
type ManifesterError struct {
	Op  string // e.g. "generate_hls", "validate", "write_file"
	Msg string // Human-readable summary
	Err error  // Optional underlying error
}

// Error implements the error interface for ManifesterError.
func (e *ManifesterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("manifester error [%s]: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("manifester error [%s]: %s", e.Op, e.Msg)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *ManifesterError) Unwrap() error {
	return e.Err
}

// NewManifesterError creates a new ManifesterError with context.
// This is the preferred constructor for wrapping manifest errors.
func NewManifesterError(op, msg string, err error) *ManifesterError {
	return &ManifesterError{
		Op:  op,
		Msg: msg,
		Err: err,
	}
}
