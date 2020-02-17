package main

import (
	"errors"
	"net"
	"strings"
)

//Host defines struct for host string in known_hosts file
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

/*
NewHost create host struct from string, the input string format as below:
[name],<ip> <key type> <public key>
*/
func NewHost(input string) (host Host, err error) {
	keys := strings.Split(input, " ")
	if len(keys) != 3 {
		return host, errors.New("invalid input string")
	}

	host.getNameIP(keys[0])
	host.KeyType = keys[1]
	host.PubKey = keys[2]

	return host, nil
}
