package transcoder

import "fmt"

// ConfigError represents an error during config loading or validation.
type ConfigError struct {
	Op   string // e.g. "read", "unmarshal", "validate"
	Path string // file path involved
	Err  error  // underlying error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s] on %q: %v", e.Op, e.Path, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// TranscoderError wraps errors that occur during the transcoding process.
// This is a robust error type to dial in what is going on when the error occurs.
// It provides detailed context for debugging and logging across pipeline stages.
type TranscoderError struct {
	Stage      string   // e.g. "validation", "command_build", "execution"
	Operation  string   // e.g. "scale", "segment", "mux"
	InputPath  string   // media file involved
	OutputPath string   // target output (dir or file)
	Command    []string // ffmpeg or other command attempted
	ExitCode   int      // if available from exec
	Message    string   // human-readable summary
	Err        error    // underlying error
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

// NewTranscoderError creates a new TranscoderError with full context
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
