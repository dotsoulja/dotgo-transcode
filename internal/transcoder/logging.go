package transcoder

import (
	"fmt"
	"time"
)

// TranscodeLogger defines logging behavior for the transcoding package.
// Supports stage-aware logging, per-variant progress, and structured error reporting.
type TranscodeLogger interface {
	LogStage(stage string, msg string)
	LogVariant(variant string, msg string)
	LogError(stage string, err error)
	LogProgress(variant string, percent float64)
}

// ConsoleLogger is the default implementation that prints to stdout.
type ConsoleLogger struct{}

func (c *ConsoleLogger) LogStage(stage, msg string) {
	fmt.Printf("[transcoder][%s] %s\n", stage, msg)
}

func (c *ConsoleLogger) LogVariant(variant, msg string) {
	fmt.Printf("[transcoder][variant:%s] %s\n", variant, msg)
}

func (c *ConsoleLogger) LogError(stage string, err error) {
	if te, ok := err.(*TranscoderError); ok {
		fmt.Printf("[transcoder][%s][error] stage=%s op=%s input=%q output=%q code=%d err=%v\n",
			stage, te.Stage, te.Operation, te.InputPath, te.OutputPath, te.ExitCode, te.Err)
	} else {
		fmt.Printf("[transcoder][%s][error] %v\n", stage, err)
	}
}

func (c *ConsoleLogger) LogProgress(variant string, percent float64) {
	fmt.Printf("[transcoder][variant:%s][progress] %.2f%% @ %s\n", variant, percent, time.Now().Format("15:04:05"))
}
