package logging

import (
	"fmt"

	"github.com/dotsoulja/dotgo-transcode/internal/analyzer"
	"github.com/dotsoulja/dotgo-transcode/internal/transcoder"
)

// UnifiedLogger provides a shared logging implementaion across pipeline stages.
// It satisfies both analyzer.AnalyzerLogger and transcoder.TranscodeLogger interfaces.
// This ensures consistent, formatting, scoped progress tracking, and structured error output
// across concurrent operations like media analysis and multi-variant transcoding.
type UnifiedLogger struct{}

func (u *UnifiedLogger) LogStage(stage, msg string) {
	fmt.Printf("[stage][%s] %s\n", stage, msg)
}

func (u *UnifiedLogger) LogVariant(variant, msg string) {
	fmt.Printf("[variant][%s] %s\n", variant, msg)
}

func (u *UnifiedLogger) LogError(stage string, err error) {
	switch e := err.(type) {
	case *analyzer.AnalyzerError:
		fmt.Printf("[analyzer][%s][error] op=%s path=%q err=%v\n", stage, e.Op, e.Path, e.Err)
	case *transcoder.TranscoderError:
		fmt.Printf("[transcoder][%s][error] stage=%s op=%s input=%q output=%q code=%d err=%v\n",
			stage, e.Stage, e.Operation, e.InputPath, e.OutputPath, e.ExitCode, e.Err)
	default:
		fmt.Printf("[error][%s] %v\n", stage, err)
	}
}

func (u *UnifiedLogger) LogProgress(label string, percent float64) {
	fmt.Printf("[progress][%s] %.2f%%\n", label, percent)
}
