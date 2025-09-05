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

// RunCommandWithProgress executes a shell command and streams stderr for progress
func RunCommandWithProgress(cmd []string, duration float64, onProgress func(percent float64)) error {
	log.Printf("🚀 Executing command with progress: %s", strings.Join(cmd, " "))
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "time=") {
				if ts := extractTimestamp(line); ts > 0 && duration > 0 {
					percent := (ts / duration) * 100
					onProgress(percent)
				}
			}
		}
	}()

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
