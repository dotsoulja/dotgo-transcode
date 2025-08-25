package transcoder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadProfile loads a TranscodeProfile from a JSON or YAML file in the profiles/ directory.
// It infers format from file extension and unmarshals into a validated TranscodeProfile.
// Returns a fully populated profile or a wrapped ConfigError with operation details.
func LoadProfile(filename string) (*TranscodeProfile, error) {
	if filename == "" {
		return nil, &ConfigError{
			Op:   "validate",
			Path: "profiles/",
			Err:  fmt.Errorf("filename is empty"),
		}
	}

	// Determine file extension and validate supported formats
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return nil, &ConfigError{
			Op:   "validate",
			Path: filename,
			Err:  fmt.Errorf("unsupported file extension %q", ext),
		}
	}

	// Construct full path to the config file
	path := filepath.Join("profiles", filename)

	// Read file contents
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigError{
			Op:   "read",
			Path: path,
			Err:  err,
		}
	}

	// Unmarshal into TranscodeProfile based on file format
	var profile TranscodeProfile
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &profile); err != nil {
			return nil, &ConfigError{
				Op:   "unmarshal_json",
				Path: path,
				Err:  err,
			}
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &profile); err != nil {
			return nil, &ConfigError{
				Op:   "unmarshal_yaml",
				Path: path,
				Err:  err,
			}
		}
	}

	// Apply default values to optional fields
	applyDefaults(&profile)

	// Validate required fields and structural integrity
	if err := validateProfile(profile); err != nil {
		return nil, &ConfigError{
			Op:   "validate",
			Path: filename,
			Err:  err,
		}
	}

	return &profile, nil
}

// applyDefaults sets fallback values for optional fields in the TranscodeProfile.
// This ensures browser-friendly behavior and avoids nil references downstream.
func applyDefaults(p *TranscodeProfile) {
	if p.AudioCodec == "" {
		p.AudioCodec = "aac" // Default to AAC for optimal browser compatibility
	}
	if p.Bitrate == nil {
		p.Bitrate = make(map[string]string)
	}
}

// validateProfile performs basic sanity checks on required fields.
// Returns a plain error to be wrapped by the caller for contextual reporting.
func validateProfile(p TranscodeProfile) error {
	if p.InputPath == "" {
		return fmt.Errorf("missing input_path")
	}
	if p.OutputDir == "" {
		return fmt.Errorf("missing output_dir")
	}
	if len(p.Resolutions) == 0 {
		return fmt.Errorf("target_res must include at least one resolution")
	}
	if p.VideoCodec == "" {
		return fmt.Errorf("missing video_codec")
	}
	if p.Container == "" {
		return fmt.Errorf("missing container format")
	}
	if p.SegmentLength <= 0 {
		return fmt.Errorf("segment_length must be a positive integer")
	}
	return nil
}
