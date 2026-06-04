// Package knownhosts provides known_hosts file I/O operations.
package knownhosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilePath returns the absolute path to the user's known_hosts file.
func FilePath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(h, ".ssh", "known_hosts"), nil
}

// Exists returns true if the known_hosts file exists.
func Exists() bool {
	name, err := FilePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(name)
	return err == nil
}

func stringToLine(input string) (lines []string) {
	start := 0
	for i := 0; i < len(input); i++ {
		if input[i] != '\n' && input[i] != '\r' {
			continue
		}

		line := strings.TrimSpace(input[start:i])
		if line != "" {
			lines = append(lines, line)
		}

		if input[i] == '\r' && i+1 < len(input) && input[i+1] == '\n' {
			i++
		}
		start = i + 1
	}

	// Last line (after final newline or no trailing newline)
	if start < len(input) {
		line := strings.TrimSpace(input[start:])
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

// ReadFile reads the known_hosts file and returns the lines as a string slice.
func ReadFile() ([]string, error) {
	name, err := FilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get known_hosts path: %w", err)
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read known_hosts: %w", err)
	}

	return stringToLine(string(b)), nil
}

// SaveFile writes the input string slice to the known_hosts file, preserving
// existing file permissions.
func SaveFile(input []string) error {
	name, err := FilePath()
	if err != nil {
		return fmt.Errorf("failed to get known_hosts path: %w", err)
	}

	// Preserve original file permissions, use 0644 as default
	perm := os.FileMode(0644)
	if info, err := os.Stat(name); err == nil {
		perm = info.Mode().Perm()
	}

	str := strings.Join(input, "\n") + "\n"

	return os.WriteFile(name, []byte(str), perm)
}
