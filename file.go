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

// Search finds hosts matching the pattern in the list.
//
// This function performs fuzzy matching on the host identifier only.
//
// Parameter Format:
//   Input: Hostname or IP address (partial match supported)
//   Example: "github", "192.168", "git"
//
// Matching Behavior:
//   - Searches only in the host part (first space-delimited field)
//   - Uses substring matching (contains, not exact)
//   - Case-sensitive
//   - Returns complete host lines for all matches
//
// Examples:
//
//   // Find all hosts containing "git"
//   results := Search(hosts, "git")
//   // Returns: ["github.com ssh-rsa...", "gitlab.com ssh-rsa..."]
//
//   // Find hosts by IP prefix
//   results := Search(hosts, "192.168")
//   // Returns: ["192.168.1.1 ssh-rsa...", "192.168.1.2 ssh-rsa..."]
//
// Design Note:
// Unlike Delete(), Search() only supports fuzzy matching because:
// - Search is inherently about finding partial matches
// - Exact match would defeat the purpose of search
// - TUI search bar uses this for filtering as you type
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

// Delete removes hosts from the list based on the pattern.
//
// This function supports two parameter formats to accommodate different use cases:
//
// Mode 1: Exact Match (Priority 1)
//   Input: Full host line including public key
//   Example: "github.com ssh-rsa AAAAB3NzaC1yc2E..."
//   Use case: TUI deletion, when you have the complete host entry
//   Behavior: String equality check on the full line
//
// Mode 2: Fuzzy Match (Priority 2, fallback)
//   Input: Hostname or IP address only
//   Example: "github.com" or "192.168.1.1"
//   Use case: CLI deletion, when you only know the host identifier
//   Behavior: Contains match on the host part (before first space)
//
// Matching Priority:
// 1. Exact full line match is checked first
// 2. Fuzzy hostname match is checked if exact match fails
//
// Design Rationale:
// The dual-mode design allows this single function to serve both TUI and CLI usage patterns:
// - TUI operates on complete host entries for precise control
// - CLI typically uses short hostnames for convenience
// Priority order ensures TUI's exact-match intent is honored first
//
// Examples:
//
//   // TUI usage (exact match)
//   hosts := Delete(hosts, "github.com ssh-rsa AAAAB3NzaC1yc2E...")
//
//   // CLI usage (fuzzy match)
//   hosts := Delete(hosts, "github.com")
func Delete(input []string, pattern string) []string {
	var out []string

	for _, v := range input {
		// Skip empty lines
		if v == "" {
			continue
		}

		// Priority 1: Exact match with full line (TUI usage)
		if v == pattern {
			continue // Skip (delete) this exact entry
		}

		// Priority 2: Fuzzy match on host part (CLI usage)
		parts := strings.Fields(v)
		if len(parts) > 0 {
			hostPart := parts[0]
			// Only match in the host part (name or IP)
			if strings.Contains(hostPart, pattern) {
				continue // Skip (delete) this entry
			}
		}

		out = append(out, v)
	}

	return out
}
