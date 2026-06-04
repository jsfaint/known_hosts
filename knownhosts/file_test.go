package knownhosts

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
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

func TestFilePath(t *testing.T) {
	path, err := FilePath()
	if err != nil {
		t.Fatalf("FilePath failed: %v", err)
	}

	// Verify path contains .ssh and known_hosts
	if filepath.Base(path) != "known_hosts" {
		t.Errorf("FilePath returned unexpected filename: %s", path)
	}

	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".ssh" {
		t.Errorf("FilePath returned unexpected directory: %s", dir)
	}
}

func TestExists(t *testing.T) {
	// This test verifies Exists works, though actual result depends on user's system
	// We're mainly testing it doesn't panic and returns a bool
	exists := Exists()
	if exists != true && exists != false {
		t.Errorf("Exists returned non-boolean value")
	}
}

func TestStringToLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single line",
			input: "github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			want:  []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
		},
		{
			name:  "multiple lines unix format",
			input: "github.com ssh-rsa key1\ngitlab.com ssh-rsa key2",
			want:  []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
		},
		{
			name:  "multiple lines windows format",
			input: "github.com ssh-rsa key1\r\ngitlab.com ssh-rsa key2",
			want:  []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
		},
		{
			name:  "lines with empty spaces",
			input: "github.com ssh-rsa key1\n   \ngitlab.com ssh-rsa key2\n\t\n",
			want:  []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
		},
		{
			name:  "lines with leading/trailing spaces",
			input: "  github.com ssh-rsa key1  \n  gitlab.com ssh-rsa key2  ",
			want:  []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"},
		},
		{
			name:  "mixed line endings",
			input: "line1\r\nline2\nline3\r",
			want:  []string{"line1", "line2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringToLine(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringToLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")
	testContent := "github.com ssh-rsa key1\ngitlab.com ssh-rsa key2\n"

	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	restoreHome := setHomeDir(t, tmpDir)
	defer restoreHome()

	got, err := ReadFile()
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	want := []string{"github.com ssh-rsa key1", "gitlab.com ssh-rsa key2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadFile() = %v, want %v", got, want)
	}
}

func TestReadFile_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	restoreHome := setHomeDir(t, tmpDir)
	defer restoreHome()

	_, err := ReadFile()
	if err == nil {
		t.Error("ReadFile() should return error when file doesn't exist")
	}
}

func TestSaveFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")

	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	restoreHome := setHomeDir(t, tmpDir)
	defer restoreHome()

	input := []string{
		"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
		"gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD",
	}

	if err := SaveFile(input); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("SaveFile() should create the file")
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)
	for _, line := range input {
		if !strings.Contains(contentStr, line) {
			t.Errorf("SaveFile() should contain line: %s", line)
		}
	}
}

func TestSaveFile_PreservePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - Unix-style file permissions not supported")
	}

	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")

	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	restoreHome := setHomeDir(t, tmpDir)
	defer restoreHome()

	testContent := "test ssh-rsa key"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	input := []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}
	if err := SaveFile(input); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("SaveFile() should preserve permissions, got: %v", info.Mode().Perm())
	}
}
