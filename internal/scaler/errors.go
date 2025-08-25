// Package scaler provides resolution scaling logic for adaptive streaming.
// This file defines custom error types used throughout the scaler package
package scaler

import (
	"fmt"
)

// ScalerError wraps errors that occur during resolution selection or scaling.
// Provides contextual clarity for debugging and fallback logic.
type ScalerError struct {
	Op  string // Operation or context where the error occurred (e.g. "selectPreset")
	Msg string // Human-readable error message
	Err error  // Optional underlying error for chaining
}

// Error implements the error interface
func (e *ScalerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("scaler error: %s: %s: %v", e.Op, e.Msg, e.Err)
	}

	return fmt.Sprintf("scaler: %s: %s", e.Op, e.Msg)
}

// Unwrap allows errors.Is and errors.As to work with ScalerError.
// It returns the underlying error, if any.
func (e *ScalerError) Unwrap() error {
	return e.Err
}

// WrapScalerError creates a new ScalerError with context.
// This is the preferred way to wrap errors in the scaler package.
//
// Example:
//
//	return WrapScalerError("selectPreset", "invalid resolution", err)
func WrapScalerError(op, msg string, err error) *ScalerError {
	return &ScalerError{
		Op:  op,
		Msg: msg,
		Err: err,
	}
}

// NewScalerError creates a standalone ScalerError with no underlying clause.
// Useful for terminal errors or validation failures.
//
// Example:
//
//	return NewScalerError("validatePreset", "resolution must be divisible by 2")
func NewScalerError(op, msg string) *ScalerError {
	return &ScalerError{
		Op:  op,
		Msg: msg,
	}
}
