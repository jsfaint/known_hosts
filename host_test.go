package main

import (
	"reflect"
	"testing"
)

func Test_getNameIP(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Host
	}{
		{"host only", "github.com", Host{Name: "github.com"}},
		{"ip only", "192.168.31.1", Host{IP: "192.168.31.1"}},
		{"both", "r2d,192.168.31.1", Host{Name: "r2d", IP: "192.168.31.1"}},
		{"IPv6 address", "2001:db8::1", Host{IP: "2001:db8::1"}},
		{"IPv4 localhost", "127.0.0.1", Host{IP: "127.0.0.1"}},
		{"IPv6 localhost", "::1", Host{IP: "::1"}},
		{"hostname with dash", "my-server.example.com", Host{Name: "my-server.example.com"}},
		{"hostname with numbers", "server123.example.com", Host{Name: "server123.example.com"}},
		{"multiple commas ignored", "a,b,c,1.2.3.4", Host{}},
		{"empty string", "", Host{}},
		{"comma at end", "example.com,", Host{Name: "example.com"}},
		{"comma at start", ",192.168.1.1", Host{Name: "", IP: "192.168.1.1"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var h Host
			h.getNameIP(test.input)
			if !reflect.DeepEqual(h, test.want) {
				t.Errorf("Not match. want: %v, got: %v", test.want, h)
			}
		})
	}
}

func TestNewHost(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Host
		wantErr bool
	}{
		{
			name:    "host only",
			input:   "github.com rsa thisisafakekey",
			want:    Host{Name: "github.com", KeyType: "rsa", PubKey: "thisisafakekey"},
			wantErr: false,
		},
		{
			name:    "ip only",
			input:   "192.168.1.1 rsa test",
			want:    Host{IP: "192.168.1.1", KeyType: "rsa", PubKey: "test"},
			wantErr: false,
		},
		{
			name:    "both name and ip",
			input:   "hello,192.168.1.1 rsa test",
			want:    Host{Name: "hello", IP: "192.168.1.1", KeyType: "rsa", PubKey: "test"},
			wantErr: false,
		},
		{
			name:    "with ed25519 key type",
			input:   "github.com ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl",
			want:    Host{Name: "github.com", KeyType: "ed25519", PubKey: "AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"},
			wantErr: false,
		},
		{
			name:    "with ecdsa key type",
			input:   "github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=",
			want:    Host{Name: "github.com", KeyType: "ecdsa-sha2-nistp256", PubKey: "AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg="},
			wantErr: false,
		},
		{
			name:    "invalid - only 2 parts",
			input:   "rsa test",
			want:    Host{},
			wantErr: true,
		},
		{
			name:    "invalid - only 1 part",
			input:   "test",
			want:    Host{},
			wantErr: true,
		},
		{
			name:    "invalid - 4 parts",
			input:   "github.com rsa test extra",
			want:    Host{},
			wantErr: true,
		},
		{
			name:    "invalid - empty string",
			input:   "",
			want:    Host{},
			wantErr: true,
		},
		{
			name:    "invalid - only spaces",
			input:   "   ",
			want:    Host{},
			wantErr: true,
		},
		{
			name:    "IPv6 address",
			input:   "2001:db8::1 rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			want:    Host{IP: "2001:db8::1", KeyType: "rsa", PubKey: "AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantErr: false,
		},
		{
			name:    "name with IPv6",
			input:   "myserver,2001:db8::1 rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			want:    Host{Name: "myserver", IP: "2001:db8::1", KeyType: "rsa", PubKey: "AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantErr: false,
		},
		{
			name:    "with hyphen in name",
			input:   "my-server.example.com rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			want:    Host{Name: "my-server.example.com", KeyType: "rsa", PubKey: "AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantErr: false,
		},
		{
			name:    "localhost IP",
			input:   "127.0.0.1 rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
			want:    Host{IP: "127.0.0.1", KeyType: "rsa", PubKey: "AAAAB3NzaC1yc2EAAAADAQABAAABAQC"},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h, err := NewHost(test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("Error want: %v, got: %v", test.wantErr, (err != nil))
			}

			if !reflect.DeepEqual(h, test.want) {
				t.Errorf("want: %v, got: %v", test.want, h)
			}
		})
	}
}

func TestHost_StringRepresentation(t *testing.T) {
	// Test Host struct fields
	h := Host{
		Name:    "github.com",
		IP:      "192.168.1.1",
		KeyType: "rsa",
		PubKey:  "AAAAB3NzaC1yc2EAAAADAQABAAABAQC",
	}

	if h.Name != "github.com" {
		t.Errorf("Host.Name = %v, want github.com", h.Name)
	}
	if h.IP != "192.168.1.1" {
		t.Errorf("Host.IP = %v, want 192.168.1.1", h.IP)
	}
	if h.KeyType != "rsa" {
		t.Errorf("Host.KeyType = %v, want rsa", h.KeyType)
	}
	if h.PubKey != "AAAAB3NzaC1yc2EAAAADAQABAAABAQC" {
		t.Errorf("Host.PubKey = %v, want AAAAB3NzaC1yc2EAAAADAQABAAABAQC", h.PubKey)
	}
}

func TestHost_EmptyFields(t *testing.T) {
	// Test with empty fields
	h1 := Host{}
	if h1.Name != "" || h1.IP != "" || h1.KeyType != "" || h1.PubKey != "" {
		t.Errorf("Empty Host struct should have empty fields, got: %v", h1)
	}

	// Test with only name
	h2 := Host{Name: "example.com"}
	if h2.Name != "example.com" || h2.IP != "" {
		t.Errorf("Host with only name should have empty IP, got: %v", h2)
	}

	// Test with only IP
	h3 := Host{IP: "192.168.1.1"}
	if h3.Name != "" || h3.IP != "192.168.1.1" {
		t.Errorf("Host with only IP should have empty Name, got: %v", h3)
	}
}
