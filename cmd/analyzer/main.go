package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/utils/logging"
)

func main() {
	logger := &logging.UnifiedLogger{}
	files := []string{
		"media/thelostboys.mp4",
		"media/1917.mp4",
		"media/hondo.mp4",
		"media/legendofthelost.mp4",
	}

	for _, f := range files {
		absPath, err := filepath.Abs(f)
		if err != nil {
			log.Printf("❌ Failed to resolve path for %s: %v\n", f, err)
			continue
		}
		// This will assume segmentLength of 0 to ensure full analysis
		info, err := analyzer.AnalyzeMedia(absPath, 0, logger)
		if err != nil {
			log.Printf("❌ Error analyzing %s: %v\n", f, err)
			continue
		}

		fmt.Printf("🎬 File: %s\n", f)
		fmt.Printf("  Duration: %.2f seconds\n", info.Duration)
		fmt.Printf("  Resolution: %dx%d\n", info.Width, info.Height)
		fmt.Printf("  Video Codec: %s\n", info.VideoCodec)
		fmt.Printf("  Audio Codec: %s\n", info.AudioCodec)
		fmt.Printf("  Bitrate: %d kbps\n", info.Bitrate)
		fmt.Printf("  Framerate: %.3f fps\n", info.Framerate)
		fmt.Printf("  Keyframe Interval: %.3f frames\n", info.KeyframeInterval)
		fmt.Printf("  Keyframes: %v\n", info.Keyframes)
		fmt.Println()
	}
}
