package transcoder

import "fmt"

// ConfigError represents an error during config loading or validation.
// Used when reading, parsing, or validating profile files (JSON/YAML).
type ConfigError struct {
	Op   string // Operation context (e.g. "read", "unmarshal", "validate")
	Path string // File path involved in the error
	Err  error  // Underlying error (wrapped for traceability)
}

// Error returns a formatted string representation of the ConfigError.
// Includes operation, path, and root error.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s] on %q: %v", e.Op, e.Path, e.Err)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
// Enables structured error handling and inspection.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// TranscoderError wraps errors that occur during the transcoding process.
// Captures full forensic context including stage, operation, command, and exit code.
type TranscoderError struct {
	Stage      string   // High-level stage (e.g. "validation", "execution", "filesystem")
	Operation  string   // Specific operation (e.g. "scale", "segment", "mux")
	InputPath  string   // Source media file path
	OutputPath string   // Target output path (file or directory)
	Command    []string // Command attempted (e.g. ffmpeg args)
	ExitCode   int      // Exit code from subprocess, if available
	Message    string   // Human-readable summary of the error
	Err        error    // Underlying error (wrapped for traceability)
}

// Error returns a formatted string representation of the TranscoderError.
// Includes stage, operation, input/output paths, command, exit code, and root error.
func (e *TranscoderError) Error() string {
	return fmt.Sprintf(
		"[%s/%s] %s\nInput: %s\nOutput: %s\nCmd: %v\nExitCode: %d\nErr: %v",
		e.Stage, e.Operation, e.Message, e.InputPath, e.OutputPath, e.Command, e.ExitCode, e.Err,
	)
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
// Enables structured error handling and inspection.
func (e *TranscoderError) Unwrap() error {
	return e.Err
}

// NewTranscoderError creates a new TranscoderError with full context.
// Preferred constructor for wrapping errors during any pipeline stage.
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
