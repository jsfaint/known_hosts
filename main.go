package main

import (
	"fmt"
	"os"

	"github.com/jsfaint/known_hosts/cli"
	"github.com/jsfaint/known_hosts/knownhosts"
)

func ensureKnownHostsExists() error {
	if knownhosts.Exists() {
		return nil
	}

	path, err := knownhosts.FilePath()
	if err != nil {
		return fmt.Errorf("cannot determine known_hosts path: %w", err)
	}
	return fmt.Errorf("known_hosts file not found at %s; connect to a host first or create it manually", path)
}

func main() {
	if err := ensureKnownHostsExists(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	hosts, err := knownhosts.ReadFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if err := cli.Run(os.Stdout, os.Args, hosts, knownhosts.SaveFile); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
