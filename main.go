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

func parseArgs() (opt opts) {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case cmdRemove:
		checkArgs(3)
		if err := validateHost(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opt.operation = cmdRemove
		opt.host = os.Args[2]
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

func deleteHost(hosts []string, host string) {
	fmt.Println("Removing host: ", host)
	hosts = Delete(hosts, host)
	if err := SaveFile(hosts); err != nil {
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
    rm      - Remove a host
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

func main() {
	if !Exists() {
		return
	}

	opt := parseArgs()

	hosts, err := ReadFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch opt.operation {
	case cmdRemove:
		deleteHost(hosts, opt.host)
	case cmdList:
		listHost(hosts)
	case cmdSearch:
		searchHost(hosts, opt.host)
	case cmdTUI:
		runTUI(hosts)
	}
}
