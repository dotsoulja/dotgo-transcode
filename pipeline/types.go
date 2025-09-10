package pipeline

import (
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// TranscodeProfile is a re-export of the internal transcoder.TranscodeProfile.
// This allows external packages to construct dynamic profiles for use with RunPipeline,
// while keeping the internal transcoder package encapsulated.
type TranscodeProfile = transcoder.TranscodeProfile

// Variant is a re-export of the transcoder.Variant type for convenience.
type Variant = transcoder.Variant
