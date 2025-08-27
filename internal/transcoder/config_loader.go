package transcoder

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadProfile loads a TranscodeProfile from a JSON or YAML file in the profiles/ directory.
// Infers format from file extension and unmarshals into a validated TranscodeProfile.
// Returns a fully populated profile or a wrapped ConfigError with operation details.
func LoadProfile(filename string) (*TranscodeProfile, error) {
	if filename == "" {
		return nil, &ConfigError{
			Op:   "validate",
			Path: "profiles/",
			Err:  fmt.Errorf("filename is empty"),
		}
	}

	// Infer file format from extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".json" && ext != ".yaml" && ext != ".yml" {
		return nil, &ConfigError{
			Op:   "validate",
			Path: filename,
			Err:  fmt.Errorf("unsupported file extension %q", ext),
		}
	}

	// Construct full path to profile file
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

	var profile TranscodeProfile

	// Unmarshal based on format
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

	// Apply fallback values for optional fields
	applyDefaults(&profile)

	// Validate required fields and log segment length behavior
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
// Ensures audio codec and bitrate map are initialized.
func applyDefaults(p *TranscodeProfile) {
	if p.AudioCodec == "" {
		p.AudioCodec = "aac"
	}
	if p.Bitrate == nil {
		p.Bitrate = make(map[string]string)
	}
}

// validateProfile performs basic sanity checks on required fields.
// Logs segment length behavior and returns error if any required field is missing.
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

	// Interpret segment length behavior
	switch {
	case p.SegmentLength < 0:
		return fmt.Errorf("segment_length must be zero or a positive integer")

	case p.SegmentLength == 0:
		log.Println("📼 segment_length not set in config—using keyframe interval for segmentation")

	default:
		log.Printf("📐 Using configured segment_length: %ds", p.SegmentLength)
	}

	return nil
}
