/*
Package main implements an utility for manage the ssh known_hosts file.
A clone of https://github.com/markmcconachie/known_hosts write in Go.
Aimming to support Windows/Linux/macOS.
*/
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type opts struct {
	operation string
	host      string
	dryRun    bool
}

const (
	cmdRemove = "rm"
	cmdList   = "ls"
	cmdHelp   = "help"
	cmdSearch = "search"
	cmdTUI    = "tui"
)

// validateHost validates host parameter
func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if strings.ContainsAny(host, "\n\r") {
		return fmt.Errorf("host cannot contain newline characters")
	}
	if len(host) > 1024 {
		return fmt.Errorf("host too long (max 1024 characters)")
	}
	return nil
}

func checkArgs(num int) {
	if len(os.Args) != num {
		fmt.Println("Invalid parameter")
		printUsage()
		os.Exit(1)
	}
}

func parseRemoveArgs(args []string) (host string, dryRun bool, err error) {
	if len(args) < 1 || len(args) > 2 {
		return "", false, fmt.Errorf("rm requires a host and supports optional --dry-run")
	}

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			if dryRun {
				return "", false, fmt.Errorf("duplicate --dry-run flag")
			}
			dryRun = true
		default:
			if host != "" {
				return "", false, fmt.Errorf("rm accepts exactly one host")
			}
			host = arg
		}
	}

	if err := validateHost(host); err != nil {
		return "", false, err
	}

	return host, dryRun, nil
}

func parseArgs() (opt opts) {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case cmdRemove:
		host, dryRun, err := parseRemoveArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opt.operation = cmdRemove
		opt.host = host
		opt.dryRun = dryRun
	case cmdList:
		checkArgs(2)
		opt.operation = cmdList
	case cmdSearch:
		checkArgs(3)
		if err := validateHost(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opt.operation = cmdSearch
		opt.host = os.Args[2]
	case cmdTUI:
		checkArgs(2)
		opt.operation = cmdTUI
	case cmdHelp:
		printUsage()
		os.Exit(0) // help is successful exit
	default:
		fmt.Println("Invalid parameter")
		printUsage()
		os.Exit(1)
	}

	return opt
}

func displayHostIdentifier(line string) string {
	host, err := NewHost(line)
	if err == nil {
		if host.Name != "" && host.IP != "" {
			return host.Name + ", " + host.IP
		}
		if host.Name != "" {
			return host.Name
		}
		if host.IP != "" {
			return host.IP
		}
	}

	parts := strings.Fields(line)
	if len(parts) > 0 {
		return parts[0]
	}

	return line
}

func previewDelete(hosts []string, host string) {
	_, removed := deleteMatches(hosts, host)
	if len(removed) == 0 {
		fmt.Println("Dry run: no matching hosts would be removed for:", host)
		return
	}

	fmt.Printf("Dry run: would remove %d entr", len(removed))
	if len(removed) == 1 {
		fmt.Println("y:")
	} else {
		fmt.Println("ies:")
	}
	for _, line := range removed {
		fmt.Printf("- %s\n", displayHostIdentifier(line))
	}
}

func deleteHost(hosts []string, host string) {
	fmt.Println("Removing host:", host)
	remaining, _ := deleteMatches(hosts, host)
	if err := SaveFile(remaining); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to delete host: %v\n", err)
		os.Exit(1)
	}
}

func searchHost(hosts []string, host string) {
	newHosts := Search(hosts, host)
	listHost(newHosts)
}

func listHost(hosts []string) {
	fmt.Println("Current known hosts:")

	for _, v := range hosts {
		if v == "" {
			continue
		}

		host, err := NewHost(v)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if host.Name == "" {
			fmt.Printf("%s\n", host.IP)
		} else if host.IP == "" {
			fmt.Printf("%s\n", host.Name)
		} else {
			fmt.Printf("%s, %s\n", host.Name, host.IP)
		}
	}
}

func printUsage() {
	fmt.Println(`
usage: known_hosts command [host]
  commands:
    ls      - List all known hosts
    rm      - Remove a host (supports --dry-run)
    search  - Search host in known hosts
    tui     - Interactive terminal UI
    help    - Show this message
    `)

}

func runTUI(hosts []string) {
	p := tea.NewProgram(
		Model{
			hosts:    hosts,
			filtered: hosts,
			mode:     viewList,
		},
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func ensureKnownHostsExists() error {
	if Exists() {
		return nil
	}

	return fmt.Errorf("known_hosts file not found in ~/.ssh/known_hosts; connect to a host first or create it manually")
}

func main() {
	opt := parseArgs()
	if err := ensureKnownHostsExists(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	hosts, err := ReadFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch opt.operation {
	case cmdRemove:
		if opt.dryRun {
			previewDelete(hosts, opt.host)
			return
		}
		deleteHost(hosts, opt.host)
	case cmdList:
		listHost(hosts)
	case cmdSearch:
		searchHost(hosts, opt.host)
	case cmdTUI:
		runTUI(hosts)
	}
}
