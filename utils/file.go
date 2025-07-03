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

func CopyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst)
	}
	return copyFile(src, dst)
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

// CleanAllExceptAIPaths removes all files/folders in current dir except .git and allowed AI paths.
func CleanAllExceptAIPaths(allowedPaths []string) error {
	keep := make(map[string]struct{})

	// Canonicalize allowed paths
	for _, p := range allowedPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("failed to resolve path %s: %w", p, err)
		}
		keep[abs] = struct{}{}
	}

	// Always preserve .git
	gitAbs, _ := filepath.Abs(".git")
	keep[gitAbs] = struct{}{}

	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read current directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		abs, _ := filepath.Abs(name)

		if _, shouldKeep := keep[abs]; shouldKeep {
			continue
		}

		if err := os.RemoveAll(name); err != nil {
			fmt.Fprintf(os.Stderr, "[warn] failed to remove %s: %v\n", name, err)
		} else {
			fmt.Printf("[info] removed: %s\n", name)
		}
	}

	return nil
}
