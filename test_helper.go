package main

import (
	"os"
	"runtime"
	"testing"
)

// setHomeDir sets the user home directory environment variable for testing.
// On Windows, it sets USERPROFILE. On Unix systems, it sets HOME.
// It returns a function to restore the original value.
func setHomeDir(t *testing.T, newDir string) (restore func()) {
	var envVar string

	if runtime.GOOS == "windows" {
		envVar = "USERPROFILE"
	} else {
		envVar = "HOME"
	}

	oldVal := os.Getenv(envVar)
	if err := os.Setenv(envVar, newDir); err != nil {
		t.Fatalf("Failed to set %s: %v", envVar, err)
	}

	return func() {
		if oldVal == "" {
			os.Unsetenv(envVar)
		} else {
			os.Setenv(envVar, oldVal)
		}
	}
}
