package analyzer

import (
	"fmt"
	"time"
)

// Purpose of this file is to define logging behavior for the analyzer package
// This ensures that there aren't any long pauses during media analysis portion of the pipeline.

// AnalyzerLogger defines logging behavior for the analyzer package.
type AnalyzerLogger interface {
	LogStage(stage string, msg string)
	LogError(stage string, err error)
	LogProgress(stage string, percent float64)
}

// ConsoleLogger is the default implementation that prints to stdout.
type ConsoleLogger struct{}

func (c *ConsoleLogger) LogStage(stage, msg string) {
	fmt.Printf("[analyzer][%s] %s\n", stage, msg)
}

func (c *ConsoleLogger) LogError(stage string, err error) {
	if ae, ok := err.(*AnalyzerError); ok {
		fmt.Printf("[analyzer][%s][error] op=%s, path=%s, err=%v\n", stage, ae.Op, ae.Path, ae.Err)
	} else {
		fmt.Printf("[analyzer][%s][error] %v\n", stage, err)
	}
}

func (c *ConsoleLogger) LogProgress(stage string, percent float64) {
	fmt.Printf("[analyzer][%s][progress] %.2f%% @ %s\n", stage, percent, time.Now().Format("15:04:05"))
}
