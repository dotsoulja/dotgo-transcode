// Package executil provides shared helpers for executing shell commands.
// Used by transcoder, segmenter, muxer, and other pipeline stages.
package executil

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RunCommand executes a shell command using os/exec.
// Logs the command and returns any execution error.
func RunCommand(cmd []string) error {
	log.Printf("🚀 Executing command: %s", strings.Join(cmd, " "))
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	execCmd.Stdout = nil
	execCmd.Stderr = nil
	return execCmd.Run()
}

// RunCommandWithProgress executes a shell command and streams stderr output to extract
// real-time progress information. It supports both traditional ffmpeg logs (e.g. "time=")
// and structured progress logs via "-progress pipe:2" (e.g. "out_time=HH:MM:SS.xx").
//
// Progress updates are emitted via the onProgress callback, throttled to avoid flooding.
// This function is concurrency-safe and designed for long-running transcoding tasks.
func RunCommandWithProgress(cmd []string, duration float64, onProgress func(percent float64)) error {
	log.Printf("🚀 Executing command with progress: %s", strings.Join(cmd, " "))
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	// Open stderr pipe for streaming ffmpeg output
	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command execution
	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	reader := bufio.NewReader(stderr)
	var lastEmit time.Time

	// Stream stderr line-by-line to extract progress
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break // EOF or pipe closed
			}

			line = strings.TrimSpace(line)

			// Parse traditional ffmpeg progress lines (e.g. "time=00:01:23.45")
			if strings.Contains(line, "time=") {
				if ts := extractTimestamp(line); ts > 0 && duration > 0 {
					percent := (ts / duration) * 100
					if time.Since(lastEmit) > 2*time.Second {
						onProgress(percent)
						lastEmit = time.Now()
					}
				}
			}

			// Parse structured progress lines from "-progress pipe:2" (e.g. "out_time=00:01:23.45")
			if strings.HasPrefix(line, "out_time=") {
				ts := parseTimestamp(strings.TrimPrefix(line, "out_time="))
				if ts > 0 && duration > 0 {
					percent := (ts / duration) * 100
					if time.Since(lastEmit) > 2*time.Second {
						onProgress(percent)
						lastEmit = time.Now()
					}
				}
			}
		}
	}()

	// Wait for command to complete
	if err := execCmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// extractTimestamp parses ffmpeg time=HH:MM:SS.xx from stderr and returns seconds.
func extractTimestamp(line string) float64 {
	re := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return 0
	}
	h, _ := strconv.Atoi(matches[1])
	m, _ := strconv.Atoi(matches[2])
	s, _ := strconv.ParseFloat(matches[3], 64)
	return float64(h*3600+m*60) + s
}

// parseTimestamp converts a timestamp string "HH:MM:SS.xx" into seconds.
// Used for structured ffmpeg progress output via "-progress pipe:2".
func parseTimestamp(ts string) float64 {
	parts := strings.Split(ts, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.ParseFloat(parts[2], 64)
	return float64(h*3600+m*60) + s
}
