package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RunGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git %s failed: %v\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	
	return nil
}

func RunGitWithOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, string(output))
	}
	
	return strings.TrimSpace(string(output)), nil
}

func BranchExists(branch string) bool {
	_, err := RunGitWithOutput("", "rev-parse", "--verify", branch)
	return err == nil
}

func GetCurrentBranch() (string, error) {
	return RunGitWithOutput("", "branch", "--show-current")
}

func PushWithRetry(dir, branch string, maxRetries int) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(i) * time.Second
			time.Sleep(backoff)
		}
		
		err := RunGit(dir, "push", "origin", branch)
		if err == nil {
			return nil
		}
		
		lastErr = err
	}
	
	return fmt.Errorf("push failed after %d retries: %w", maxRetries, lastErr)
}

func IsGitRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func HasUncommittedChanges(dir string) bool {
	output, err := RunGitWithOutput(dir, "status", "--porcelain")
	if err != nil {
		return false
	}
	return output != ""
}