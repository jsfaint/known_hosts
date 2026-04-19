package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"valid hostname", "github.com", false},
		{"valid IP", "192.168.1.1", false},
		{"valid IPv6", "2001:db8::1", false},
		{"valid host with port", "example.com:22", false},
		{"empty string", "", true},
		{"contains newline", "github.com\n", true},
		{"contains carriage return", "github.com\r", true},
		{"too long", strings.Repeat("a", 1025), true},
		{"exactly 1024", strings.Repeat("a", 1024), false},
		{"contains space", "github com", false},
		{"special characters", "my-server_01.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListHost(t *testing.T) {
	tests := []struct {
		name         string
		hosts        []string
		wantContains []string
	}{
		{
			name:         "empty list",
			hosts:        []string{},
			wantContains: []string{"Current known hosts:"},
		},
		{
			name:         "single host with name only",
			hosts:        []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantContains: []string{"Current known hosts:", "github.com"},
		},
		{
			name:         "single host with IP only",
			hosts:        []string{"192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantContains: []string{"Current known hosts:", "192.168.1.1"},
		},
		{
			name:         "host with both name and IP",
			hosts:        []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantContains: []string{"Current known hosts:", "myserver, 192.168.1.1"},
		},
		{
			name: "multiple hosts",
			hosts: []string{
				"github.com ssh-rsa key1",
				"gitlab.com ssh-rsa key2",
				"192.168.1.1 ssh-rsa key3",
			},
			wantContains: []string{"github.com", "gitlab.com", "192.168.1.1"},
		},
		{
			name:         "host with invalid format",
			hosts:        []string{"invalid-host", "github.com ssh-rsa key"},
			wantContains: []string{"github.com"},
		},
		{
			name:         "skip empty lines",
			hosts:        []string{"", "github.com ssh-rsa key", ""},
			wantContains: []string{"github.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			listHost(tt.hosts)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check that expected strings are present
			for _, expected := range tt.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("listHost() output should contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestSearchHost(t *testing.T) {
	tests := []struct {
		name         string
		hosts        []string
		searchTerm   string
		wantContains []string
	}{
		{
			name:         "search found",
			hosts:        []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
			searchTerm:   "github",
			wantContains: []string{"github.com"},
		},
		{
			name:         "search not found",
			hosts:        []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
			searchTerm:   "bitbucket",
			wantContains: []string{"Current known hosts:"},
		},
		{
			name:         "search IP",
			hosts:        []string{"192.168.1.1 ssh-rsa key1", "192.168.1.2 ssh-rsa key2"},
			searchTerm:   "192.168.1.1",
			wantContains: []string{"192.168.1.1"},
		},
		{
			name:         "search partial",
			hosts:        []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			searchTerm:   "git",
			wantContains: []string{"github.com", "gitlab.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			searchHost(tt.hosts, tt.searchTerm)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			for _, expected := range tt.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("searchHost() output should contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestDeleteHost(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		tmpDir := t.TempDir()
		sshDir := tmpDir + "/.ssh"

		// Create .ssh directory
		if err := os.MkdirAll(sshDir, 0755); err != nil {
			t.Fatalf("Failed to create .ssh directory: %v", err)
		}

		restoreHome := setHomeDir(t, tmpDir)
		defer restoreHome()

		// Create initial known_hosts file
		initialHosts := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"192.168.1.1 ssh-rsa key3",
		}
		if err := SaveFile(initialHosts); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		deleteHost(initialHosts, "gitlab.com")

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "Removing host:") {
			t.Errorf("deleteHost() should output removal message, got: %s", output)
		}

		// Verify file was updated
		updatedHosts, err := ReadFile()
		if err != nil {
			t.Fatalf("Failed to read updated file: %v", err)
		}

		// Check that gitlab.com was removed
		for _, host := range updatedHosts {
			if strings.Contains(host, "gitlab.com") {
				t.Error("deleteHost() should have removed gitlab.com")
			}
		}

		// Check that other hosts remain
		found := false
		for _, host := range updatedHosts {
			if strings.Contains(host, "github.com") {
				found = true
				break
			}
		}
		if !found {
			t.Error("deleteHost() should have kept github.com")
		}
	})

	t.Run("delete with save failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		sshDir := tmpDir + "/.ssh"

		// Create .ssh directory
		if err := os.MkdirAll(sshDir, 0755); err != nil {
			t.Fatalf("Failed to create .ssh directory: %v", err)
		}

		restoreHome := setHomeDir(t, tmpDir)
		defer restoreHome()

		// Create initial known_hosts file
		initialHosts := []string{"github.com ssh-rsa key1"}
		if err := SaveFile(initialHosts); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make the directory read-only to cause save failure
		if err := os.Chmod(sshDir, 0444); err != nil {
			t.Fatalf("Failed to chmod directory: %v", err)
		}
		defer func() { _ = os.Chmod(sshDir, 0755) }()

		// This test is skipped because deleteHost calls os.Exit(1) on error
		// which cannot be easily tested in unit tests
		t.Skip("deleteHost with save failure calls os.Exit(1)")
	})
}

func TestPreviewDelete(t *testing.T) {
	tests := []struct {
		name         string
		hosts        []string
		host         string
		wantContains []string
	}{
		{
			name: "match by host part",
			hosts: []string{
				"github.com ssh-rsa key1",
				"github.com ssh-ed25519 key2",
				"gitlab.com ssh-rsa key3",
			},
			host: "github.com",
			wantContains: []string{
				"Dry run: would remove 2 entries:",
				"- github.com",
			},
		},
		{
			name: "no matching host",
			hosts: []string{
				"github.com ssh-rsa key1",
			},
			host:         "bitbucket.org",
			wantContains: []string{"Dry run: no matching hosts would be removed for: bitbucket.org"},
		},
		{
			name: "full line falls back to host display",
			hosts: []string{
				"myserver,192.168.1.1 ssh-rsa key1",
			},
			host:         "myserver,192.168.1.1 ssh-rsa key1",
			wantContains: []string{"Dry run: would remove 1 entry:", "- myserver, 192.168.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			previewDelete(tt.hosts, tt.host)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			for _, expected := range tt.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("previewDelete() output should contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	expectedCommands := []string{"ls", "rm", "search", "tui", "help"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("printUsage() should contain %q", cmd)
		}
	}
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		setup    func()
		teardown func()
	}{
		{
			name: "correct number of args",
			num:  3,
			setup: func() {
				os.Args = []string{"cmd", "rm", "github.com"}
			},
			teardown: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			// Only test the success case
			// os.Exit() cannot be tested directly in unit tests
			checkArgs(tt.num)
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOpts opts
	}{
		{
			name:     "remove command",
			args:     []string{"cmd", "rm", "github.com"},
			wantOpts: opts{operation: cmdRemove, host: "github.com"},
		},
		{
			name:     "remove command with trailing dry-run",
			args:     []string{"cmd", "rm", "github.com", "--dry-run"},
			wantOpts: opts{operation: cmdRemove, host: "github.com", dryRun: true},
		},
		{
			name:     "remove command with leading dry-run",
			args:     []string{"cmd", "rm", "--dry-run", "github.com"},
			wantOpts: opts{operation: cmdRemove, host: "github.com", dryRun: true},
		},
		{
			name:     "list command",
			args:     []string{"cmd", "ls"},
			wantOpts: opts{operation: cmdList},
		},
		{
			name:     "search command",
			args:     []string{"cmd", "search", "git"},
			wantOpts: opts{operation: cmdSearch, host: "git"},
		},
		{
			name:     "tui command",
			args:     []string{"cmd", "tui"},
			wantOpts: opts{operation: cmdTUI},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = tt.args

			got := parseArgs()

			if got.operation != tt.wantOpts.operation {
				t.Errorf("parseArgs() operation = %v, want %v", got.operation, tt.wantOpts.operation)
			}
			if got.host != tt.wantOpts.host {
				t.Errorf("parseArgs() host = %v, want %v", got.host, tt.wantOpts.host)
			}
			if got.dryRun != tt.wantOpts.dryRun {
				t.Errorf("parseArgs() dryRun = %v, want %v", got.dryRun, tt.wantOpts.dryRun)
			}
		})
	}
}

func TestParseRemoveArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantHost    string
		wantDryRun  bool
		wantErr     bool
		wantErrText string
	}{
		{
			name:       "host only",
			args:       []string{"github.com"},
			wantHost:   "github.com",
			wantDryRun: false,
		},
		{
			name:       "host and dry-run",
			args:       []string{"github.com", "--dry-run"},
			wantHost:   "github.com",
			wantDryRun: true,
		},
		{
			name:       "dry-run then host",
			args:       []string{"--dry-run", "github.com"},
			wantHost:   "github.com",
			wantDryRun: true,
		},
		{
			name:        "duplicate dry-run",
			args:        []string{"--dry-run", "github.com", "--dry-run"},
			wantErr:     true,
			wantErrText: "rm requires a host",
		},
		{
			name:        "two hosts",
			args:        []string{"github.com", "gitlab.com"},
			wantErr:     true,
			wantErrText: "rm accepts exactly one host",
		},
		{
			name:        "missing host",
			args:        []string{"--dry-run"},
			wantErr:     true,
			wantErrText: "host cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotDryRun, err := parseRemoveArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRemoveArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("parseRemoveArgs() error = %v, want substring %q", err, tt.wantErrText)
				}
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("parseRemoveArgs() host = %q, want %q", gotHost, tt.wantHost)
			}
			if gotDryRun != tt.wantDryRun {
				t.Errorf("parseRemoveArgs() dryRun = %v, want %v", gotDryRun, tt.wantDryRun)
			}
		})
	}
}

func TestValidateHostErrors(t *testing.T) {
	err := validateHost("")
	if err == nil {
		t.Error("validateHost() should return error for empty string")
	}

	err = validateHost(strings.Repeat("a", 1025))
	if err == nil {
		t.Error("validateHost() should return error for string > 1024 chars")
	}

	expectedErrMsg := "too long"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("validateHost() error message should contain %q, got: %v", expectedErrMsg, err)
	}
}

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
