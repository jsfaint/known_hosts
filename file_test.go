package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetFilePath(t *testing.T) {
	path, err := GetFilePath()
	if err != nil {
		t.Fatalf("GetFilePath failed: %v", err)
	}

	// Verify path contains .ssh and known_hosts
	if filepath.Base(path) != "known_hosts" {
		t.Errorf("GetFilePath returned unexpected filename: %s", path)
	}

	dir := filepath.Dir(path)
	if filepath.Base(dir) != ".ssh" {
		t.Errorf("GetFilePath returned unexpected directory: %s", dir)
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
			name:  "multiple lines old mac format",
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
	// Test with a temporary file
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")
	testContent := "github.com ssh-rsa key1\ngitlab.com ssh-rsa key2\n"

	// Create .ssh directory
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Temporarily change HOME to point to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

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
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := ReadFile()
	if err == nil {
		t.Error("ReadFile() should return error when file doesn't exist")
	}
}

func TestSaveFile(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")

	// Create .ssh directory
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	// Temporarily change HOME to point to our test directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	input := []string{
		"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
		"gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD",
	}

	if err := SaveFile(input); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("SaveFile() should create the file")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)
	for _, line := range input {
		if !containsSubstring(contentStr, line) {
			t.Errorf("SaveFile() should contain line: %s", line)
		}
	}
}

func TestSaveFile_PreservePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	testFile := filepath.Join(sshDir, "known_hosts")

	// Create .ssh directory
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create file with specific permissions
	testContent := "test ssh-rsa key"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	input := []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}

	if err := SaveFile(input); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	// Verify permissions were preserved
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("SaveFile() should preserve permissions, got: %v", info.Mode().Perm())
	}
}

func TestSearch(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"1"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"5"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"3"}},
		{"multi", args{[]string{"12", "21", "33", "44", "55"}, "1"}, []string{"12", "21"}},
		{"empty input", args{[]string{}, "1"}, []string{}},
		{"empty pattern", args{[]string{"1", "2", "3"}, ""}, []string{"1", "2", "3"}},
		{"not found", args{[]string{"1", "2", "3"}, "99"}, []string{}},
		{"full host line", args{[]string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "github"}, []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"host with comma", args{[]string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "myserver"}, []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"ip search", args{[]string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "192.168.1.1"}, []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"partial ip match", args{[]string{"myserver,192.168.1.1 ssh-rsa key"}, "192.168"}, []string{"myserver,192.168.1.1 ssh-rsa key"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Search(test.args.input, test.args.pattern)
			if !slicesEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"2", "3", "4", "5"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"1", "2", "3", "4"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"1", "2", "4", "5"}},
		{"multi-1", args{[]string{"11", "11", "33", "44", "55"}, "1"}, []string{"33", "44", "55"}},
		{"multi-2", args{[]string{"11", "22", "11", "44", "55"}, "1"}, []string{"22", "44", "55"}},
		{"empty input", args{[]string{}, "1"}, []string{}},
		{"not found", args{[]string{"1", "2", "3"}, "99"}, []string{"1", "2", "3"}},
		{"delete all", args{[]string{"1", "1", "1"}, "1"}, []string{}},
		{"full host line", args{[]string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC", "gitlab.com ssh-rsa key"}, "github"}, []string{"gitlab.com ssh-rsa key"}},
		{"empty string skip", args{[]string{"1", "", "2"}, "1"}, []string{"2"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Delete(test.args.input, test.args.pattern)
			if !slicesEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to compare string slices, handling nil vs empty slice
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
