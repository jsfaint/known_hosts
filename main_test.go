package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureKnownHostsExists(t *testing.T) {
	t.Run("existing known_hosts returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		sshDir := filepath.Join(tmpDir, ".ssh")
		testFile := filepath.Join(sshDir, "known_hosts")
		if err := os.MkdirAll(sshDir, 0755); err != nil {
			t.Fatalf("Failed to create .ssh directory: %v", err)
		}
		if err := os.WriteFile(testFile, []byte("github.com ssh-rsa key1\n"), 0644); err != nil {
			t.Fatalf("Failed to create known_hosts file: %v", err)
		}

		restoreHome := setHomeDir(t, tmpDir)
		defer restoreHome()

		if err := ensureKnownHostsExists(); err != nil {
			t.Fatalf("ensureKnownHostsExists() error = %v, want nil", err)
		}
	})

	t.Run("missing known_hosts returns helpful error", func(t *testing.T) {
		tmpDir := t.TempDir()
		restoreHome := setHomeDir(t, tmpDir)
		defer restoreHome()

		err := ensureKnownHostsExists()
		if err == nil {
			t.Fatal("ensureKnownHostsExists() error = nil, want helpful error")
		}
		if !strings.Contains(err.Error(), "known_hosts file not found") {
			t.Fatalf("ensureKnownHostsExists() error = %v, want missing file guidance", err)
		}
		if !strings.Contains(err.Error(), "create it manually") {
			t.Fatalf("ensureKnownHostsExists() error = %v, want actionable guidance", err)
		}
	})
}
