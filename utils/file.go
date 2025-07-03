package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func FileContains(path, line string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(line) {
			return true
		}
	}

	return false
}

func AppendToFile(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func EnsureSymlink(from, to string) error {
	if _, err := os.Lstat(from); err == nil {
		os.Remove(from)
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "mklink", "/J", from, to)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create junction: %w", err)
		}
		return nil
	}

	return os.Symlink(to, from)
}

// EnsureSymlinkIfExists creates a symlink (or junction on Windows) from `from` to `to`
// if the `from` path already exists, it will be removed first.
// If `from` does not exist, nothing is done.
func EnsureSymlinkIfExists(from, to string) error {
	_, err := os.Lstat(from)
	if os.IsNotExist(err) {
		// No file to link from → nothing to do
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", from, err)
	}

	// Remove the existing path (file/dir/symlink)
	if err := os.RemoveAll(from); err != nil {
		return fmt.Errorf("failed to remove existing path at %s: %w", from, err)
	}

	if runtime.GOOS == "windows" {
		// Use mklink /J for directory junction
		cmd := exec.Command("cmd", "/c", "mklink", "/J", from, to)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create Windows junction (mklink /J %s %s): %w", from, to, err)
		}
		return nil
	}

	// Unix-like symlink
	if err := os.Symlink(to, from); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", from, to, err)
	}

	return nil
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
