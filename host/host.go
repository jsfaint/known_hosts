// Package host provides Host parsing and representation for known_hosts entries.
package host

import (
	"fmt"
	"net"
	"strings"
)

// Host represents a known hosts entry parsed from a single line.
type Host struct {
	Name    string
	IP      string
	KeyType string
	PubKey  string
}

func (h *Host) getNameIP(value string) {
	name := strings.Split(value, ",")

	switch len(name) {
	case 1:
		if net.ParseIP(name[0]) == nil {
			h.Name = name[0]
		} else {
			h.IP = name[0]
		}
	case 2:
		h.Name = name[0]
		h.IP = name[1]
	default:
	}
}

// DisplayName returns a human-readable host identifier combining name and IP.
func (h *Host) DisplayName() string {
	if h.Name != "" && h.IP != "" {
		return h.Name + ", " + h.IP
	}
	if h.Name != "" {
		return h.Name
	}
	return h.IP
}

// NewHost parses a single known_hosts line into a Host struct.
// The line must have exactly 3 space-separated fields: hostpart keytype pubkey.
func NewHost(input string) (host Host, err error) {
	keys := strings.Split(input, " ")
	if len(keys) != 3 {
		return host, fmt.Errorf("invalid host: '%s'", input)
	}

	host.getNameIP(keys[0])
	host.KeyType = keys[1]
	host.PubKey = keys[2]

	return host, nil
}
