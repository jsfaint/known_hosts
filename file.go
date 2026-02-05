package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dosFormat  string = "\r\n"
	unixFormat string = "\n"
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

func getLinebreak() string {
	if runtime.GOOS == "windows" {
		return dosFormat
	}

	return unixFormat
}

func stringToLine(input string) (lines []string) {
	// Normalize line endings: handle \r\n (Windows), \n (Unix), \r (old Mac)
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")
	tmp := strings.Split(input, "\n")

	for _, v := range tmp {
		v = strings.TrimSpace(v)
		if v != "" { // Skip empty lines
			lines = append(lines, v)
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

	str := strings.Join(input, getLinebreak()) + getLinebreak()

	return os.WriteFile(name, []byte(str), perm)
}

// Search Host from list
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

// Delete Host from list
func Delete(input []string, pattern string) []string {
	var out []string

	for _, v := range input {
		// Split by whitespace to extract host part
		// Format: [name,]ip keytype publickey
		parts := strings.Fields(v)
		if len(parts) > 0 {
			hostPart := parts[0]
			// Only match in the host part (name or IP)
			if strings.Contains(hostPart, pattern) {
				continue // Skip (delete) this entry
			}
		}

		if v == "" {
			continue
		}

		out = append(out, v)
	}

	return out
}
