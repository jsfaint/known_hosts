package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jsfaint/known_hosts/knownhosts"
)

// setHomeDir sets the user home directory environment variable for testing.
func setHomeDir(t *testing.T, newDir string) func() {
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
			_ = os.Unsetenv(envVar)
		} else {
			_ = os.Setenv(envVar, oldVal)
		}
	}
}

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
			var buf bytes.Buffer
			listHost(&buf, tt.hosts)
			output := buf.String()

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
			var buf bytes.Buffer
			searchHost(&buf, tt.hosts, tt.searchTerm)
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
		sshDir := filepath.Join(tmpDir, ".ssh")

		if err := os.MkdirAll(sshDir, 0755); err != nil {
			t.Fatalf("Failed to create .ssh directory: %v", err)
		}

		restoreHome := setHomeDir(t, tmpDir)
		defer restoreHome()

		initialHosts := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"192.168.1.1 ssh-rsa key3",
		}
		if err := knownhosts.SaveFile(initialHosts); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		var buf bytes.Buffer
		if err := deleteHost(&buf, initialHosts, "gitlab.com", knownhosts.SaveFile); err != nil {
			t.Fatalf("deleteHost() error = %v", err)
		}
		output := buf.String()

		if !strings.Contains(output, "Removing host:") {
			t.Errorf("deleteHost() should output removal message, got: %s", output)
		}

		updatedHosts, err := knownhosts.ReadFile()
		if err != nil {
			t.Fatalf("Failed to read updated file: %v", err)
		}

		for _, host := range updatedHosts {
			if strings.Contains(host, "gitlab.com") {
				t.Error("deleteHost() should have removed gitlab.com")
			}
		}

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
		var buf bytes.Buffer
		saveFail := func([]string) error {
			return fmt.Errorf("save failed")
		}
		err := deleteHost(&buf, []string{"github.com ssh-rsa key1"}, "github.com", saveFail)
		if err == nil {
			t.Error("deleteHost() should return error when save fails")
		}
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
			var buf bytes.Buffer
			previewDelete(&buf, tt.hosts, tt.host)
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
	var buf bytes.Buffer
	printUsage(&buf)
	output := buf.String()

	expectedCommands := []string{"ls", "rm", "search", "tui", "help"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("printUsage() should contain %q", cmd)
		}
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    opts
		wantErr bool
	}{
		{
			name:    "remove command",
			args:    []string{"cmd", "rm", "github.com"},
			want:    opts{operation: cmdRemove, host: "github.com"},
			wantErr: false,
		},
		{
			name:    "remove command with trailing dry-run",
			args:    []string{"cmd", "rm", "github.com", "--dry-run"},
			want:    opts{operation: cmdRemove, host: "github.com", dryRun: true},
			wantErr: false,
		},
		{
			name:    "remove command with leading dry-run",
			args:    []string{"cmd", "rm", "--dry-run", "github.com"},
			want:    opts{operation: cmdRemove, host: "github.com", dryRun: true},
			wantErr: false,
		},
		{
			name:    "list command",
			args:    []string{"cmd", "ls"},
			want:    opts{operation: cmdList},
			wantErr: false,
		},
		{
			name:    "search command",
			args:    []string{"cmd", "search", "git"},
			want:    opts{operation: cmdSearch, host: "git"},
			wantErr: false,
		},
		{
			name:    "tui command",
			args:    []string{"cmd", "tui"},
			want:    opts{operation: cmdTUI},
			wantErr: false,
		},
		{
			name:    "help command",
			args:    []string{"cmd", "help"},
			want:    opts{operation: cmdHelp},
			wantErr: false,
		},
		{
			name:    "list with extra args",
			args:    []string{"cmd", "ls", "extra"},
			want:    opts{},
			wantErr: true,
		},
		{
			name:    "unknown command",
			args:    []string{"cmd", "unknown"},
			want:    opts{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.operation != tt.want.operation {
				t.Errorf("parseArgs() operation = %v, want %v", got.operation, tt.want.operation)
			}
			if got.host != tt.want.host {
				t.Errorf("parseArgs() host = %v, want %v", got.host, tt.want.host)
			}
			if got.dryRun != tt.want.dryRun {
				t.Errorf("parseArgs() dryRun = %v, want %v", got.dryRun, tt.want.dryRun)
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
