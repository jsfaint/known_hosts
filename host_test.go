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
		{"host only", "github.com rsa thisisafakekey", Host{Name: "github.com", KeyType: "rsa", PubKey: "thisisafakekey"}, false},
		{"ip only", "192.168.1.1 rsa test", Host{IP: "192.168.1.1", KeyType: "rsa", PubKey: "test"}, false},
		{"both", "hello,192.168.1.1 rsa test", Host{Name: "hello", IP: "192.168.1.1", KeyType: "rsa", PubKey: "test"}, false},
		{"invalid1", "rsa test", Host{}, true},
		{"invalid2", "test", Host{}, true},
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
