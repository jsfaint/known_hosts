package knownhosts

import "strings"

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

// DeleteMatches removes entries matching the given pattern.
// Priority 1: exact match on full line. Priority 2: exact match on host part only.
// Returns both remaining and removed slices.
// SECURITY: uses exact match (not Contains) to prevent accidental bulk deletion.
func DeleteMatches(input []string, pattern string) (remaining []string, removed []string) {
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

// Delete removes entries matching the pattern. See DeleteMatches for matching rules.
func Delete(input []string, pattern string) []string {
	remaining, _ := DeleteMatches(input, pattern)
	return remaining
}
