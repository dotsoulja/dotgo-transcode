package transcoder

import "fmt"

// ConfigError represents an error during config loading or validation.
// It wraps the operation, file path, and underlying error for forensic clarity.
type ConfigError struct {
	Op   string // Operation context e.g. "read", "unmarshal", "validate"
	Path string // file path involved
	Err  error  // underlying error
}

// Error returns a formatted string representation of the ConfigError
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s] on %q: %v", e.Op, e.Path, e.Err)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// TranscoderError wraps errors that occur during the transcoding process.
// It provides detailed context for debugging and logging across pipeline stages.
type TranscoderError struct {
	Stage      string   // High-level stage (e.g. "validation", "execution")
	Operation  string   // Specific operation (e.g. "scale", "segment", "mux")
	InputPath  string   // Source media file
	OutputPath string   // Target output (dir or file)
	Command    []string // Command attempted (e.g. ffmpeg args)
	ExitCode   int      // Exit code from subprocess, if available
	Message    string   // Human-readable summary of the error
	Err        error    // Underlying error
}

// Error returns a formatted string representation of the TranscoderError
func (e *TranscoderError) Error() string {
	return fmt.Sprintf(
		"[%s/%s] %s\nInput: %s\nOutput: %s\nCmd: %v\nExitCode: %d\nErr: %v",
		e.Stage, e.Operation, e.Message, e.InputPath, e.OutputPath, e.Command, e.ExitCode, e.Err,
	)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *TranscoderError) Unwrap() error {
	return e.Err
}

// NewTranscoderError creates a new TranscoderError with full context.
// This is the preferred constructor for wrapping errors during transcoding.
func NewTranscoderError(stage, operation, input, output, msg string, cmd []string, code int, err error) *TranscoderError {
	return &TranscoderError{
		Stage:      stage,
		Operation:  operation,
		InputPath:  input,
		OutputPath: output,
		Command:    cmd,
		ExitCode:   code,
		Message:    msg,
		Err:        err,
	}
}
