package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetFilePath returns the filepath of known_hosts
func GetFilePath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(h, ".ssh", "known_hosts"), nil
}

// Exists returns the file existence
func Exists() bool {
	name, err := GetFilePath()
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

// ReadFile read known_hosts file and returns a string slice
func ReadFile() ([]string, error) {
	name, err := GetFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get known_hosts path: %w", err)
	}

	b, err := os.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read known_hosts: %w", err)
	}

	return stringToLine(string(b)), nil
}

// SaveFile save the input string slice to known_hosts file
func SaveFile(input []string) error {
	name, err := GetFilePath()
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

// Search finds hosts matching the pattern in the host part (first space-delimited field).
// Uses substring (contains) matching. Only matches against the host identifier, not key type or public key.
func Search(input []string, pattern string) []string {
	var out []string

	for _, v := range input {
		// Split by whitespace to extract host part
		// Format: [name,]ip keytype publickey
		parts := strings.Fields(v)
		if len(parts) > 0 {
			hostPart := parts[0]
			// Only match in the host part (name or IP)
			if strings.Contains(hostPart, pattern) {
				out = append(out, v)
			}
		}
	}

	return out
}

// deleteMatches removes entries matching the given pattern.
// Priority 1: exact match on full line. Priority 2: exact match on host part only.
// Returns both remaining and removed slices.
// SECURITY: uses exact match (not Contains) to prevent accidental bulk deletion.
func deleteMatches(input []string, pattern string) (remaining []string, removed []string) {
	for _, v := range input {
		if v == "" {
			continue
		}

		if v == pattern {
			removed = append(removed, v)
			continue
		}

		parts := strings.Fields(v)
		if len(parts) > 0 && parts[0] == pattern {
			removed = append(removed, v)
			continue
		}

		remaining = append(remaining, v)
	}

	return remaining, removed
}

// Delete removes entries matching the pattern. See deleteMatches for matching rules.
func Delete(input []string, pattern string) []string {
	remaining, _ := deleteMatches(input, pattern)
	return remaining
}
