package main

import (
	"bytes"
	"os"
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
		name string
		hosts []string
		wantContains []string
	}{
		{
			name: "empty list",
			hosts: []string{},
			wantContains: []string{"Current known hosts:"},
		},
		{
			name: "single host with name only",
			hosts: []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantContains: []string{"Current known hosts:", "github.com"},
		},
		{
			name: "single host with IP only",
			hosts: []string{"192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantContains: []string{"Current known hosts:", "192.168.1.1"},
		},
		{
			name: "host with both name and IP",
			hosts: []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
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
			name: "host with invalid format",
			hosts: []string{"invalid-host", "github.com ssh-rsa key"},
			wantContains: []string{"github.com"},
		},
		{
			name: "skip empty lines",
			hosts: []string{"", "github.com ssh-rsa key", ""},
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
			buf.ReadFrom(r)
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
		name       string
		hosts      []string
		searchTerm string
		wantContains []string
	}{
		{
			name:       "search found",
			hosts:      []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
			searchTerm: "github",
			wantContains: []string{"github.com"},
		},
		{
			name:       "search not found",
			hosts:      []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
			searchTerm: "bitbucket",
			wantContains: []string{"Current known hosts:"},
		},
		{
			name:       "search IP",
			hosts:      []string{"192.168.1.1 ssh-rsa key1", "192.168.1.2 ssh-rsa key2"},
			searchTerm: "192.168.1.1",
			wantContains: []string{"192.168.1.1"},
		},
		{
			name:       "search partial",
			hosts:      []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			searchTerm: "git",
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
			buf.ReadFrom(r)
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

		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

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
		buf.ReadFrom(r)
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

		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		// Create initial known_hosts file
		initialHosts := []string{"github.com ssh-rsa key1"}
		if err := SaveFile(initialHosts); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make the directory read-only to cause save failure
		if err := os.Chmod(sshDir, 0444); err != nil {
			t.Fatalf("Failed to chmod directory: %v", err)
		}
		defer os.Chmod(sshDir, 0755)

		// This test is skipped because deleteHost calls os.Exit(1) on error
		// which cannot be easily tested in unit tests
		t.Skip("deleteHost with save failure calls os.Exit(1)")
	})
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
	buf.ReadFrom(r)
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
