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
