// Package executil provides shared helpers for executing shell commands.
// Used by transcoder, segmenter, muxer, and other pipeline stages.
package executil

import (
	"log"
	"os/exec"
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
