package utils

import (
	"os"
	"os/exec"
)

// GitCommand wraps git command execution
func GitCommand(showProgress bool, workDir string, args ...string) ([]byte, error) {
	cmdArgs := []string{"-C", workDir}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("git", cmdArgs...)
	if showProgress {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return nil, cmd.Run()
	}
	return cmd.CombinedOutput()
}
