// Package cli provides the CLI argument parsing, command dispatch, and output rendering.
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/jsfaint/known_hosts/host"
	"github.com/jsfaint/known_hosts/knownhosts"
	"github.com/jsfaint/known_hosts/tui"
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

// Run parses CLI arguments and executes the requested command.
// Output is written to w, not os.Stdout directly. saveFn is used to persist changes.
func Run(w io.Writer, args []string, hosts []string, saveFn func([]string) error) error {
	opt, err := parseArgs(args)
	if err != nil {
		printUsage(w)
		return err
	}

	switch opt.operation {
	case cmdRemove:
		if opt.dryRun {
			previewDelete(w, hosts, opt.host)
			return nil
		}
		return deleteHost(w, hosts, opt.host, saveFn)
	case cmdList:
		listHost(w, hosts)
	case cmdSearch:
		searchHost(w, hosts, opt.host)
	case cmdTUI:
		return tui.Run(hosts)
	case cmdHelp:
		printUsage(w)
	}
	return nil
}

// validateHost validates the host parameter.
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

func parseArgs(args []string) (opts, error) {
	if len(args) < 2 {
		return opts{}, fmt.Errorf("insufficient arguments")
	}

	switch args[1] {
	case cmdRemove:
		host, dryRun, err := parseRemoveArgs(args[2:])
		if err != nil {
			return opts{}, err
		}
		return opts{operation: cmdRemove, host: host, dryRun: dryRun}, nil
	case cmdList:
		if len(args) != 2 {
			return opts{}, fmt.Errorf("invalid parameter")
		}
		return opts{operation: cmdList}, nil
	case cmdSearch:
		if len(args) != 3 {
			return opts{}, fmt.Errorf("invalid parameter")
		}
		if err := validateHost(args[2]); err != nil {
			return opts{}, err
		}
		return opts{operation: cmdSearch, host: args[2]}, nil
	case cmdTUI:
		if len(args) != 2 {
			return opts{}, fmt.Errorf("invalid parameter")
		}
		return opts{operation: cmdTUI}, nil
	case cmdHelp:
		return opts{operation: cmdHelp}, nil
	default:
		return opts{}, fmt.Errorf("invalid parameter")
	}
}

// listHost prints all known hosts to w.
func listHost(w io.Writer, hosts []string) {
	_, _ = fmt.Fprintln(w, "Current known hosts:")

	for _, v := range hosts {
		if v == "" {
			continue
		}

		h, err := host.NewHost(v)
		if err != nil {
			_, _ = fmt.Fprintln(w, err)
			continue
		}

		_, _ = fmt.Fprintln(w, h.DisplayName())
	}
}

// searchHost searches hosts by pattern and prints matching entries.
func searchHost(w io.Writer, hosts []string, hostPattern string) {
	newHosts := knownhosts.Search(hosts, hostPattern)
	listHost(w, newHosts)
}

// deleteHost removes a host and persists the change via saveFn.
func deleteHost(w io.Writer, hosts []string, hostPattern string, saveFn func([]string) error) error {
	_, _ = fmt.Fprintln(w, "Removing host:", hostPattern)
	remaining := knownhosts.Delete(hosts, hostPattern)
	if err := saveFn(remaining); err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}
	return nil
}

// previewDelete shows what would be deleted without actually removing.
func previewDelete(w io.Writer, hosts []string, hostPattern string) {
	_, removed := knownhosts.DeleteMatches(hosts, hostPattern)
	if len(removed) == 0 {
		_, _ = fmt.Fprintln(w, "Dry run: no matching hosts would be removed for:", hostPattern)
		return
	}

	plural := "y"
	if len(removed) > 1 {
		plural = "ies"
	}
	_, _ = fmt.Fprintf(w, "Dry run: would remove %d entr%s:\n", len(removed), plural)
	for _, line := range removed {
		h, err := host.NewHost(line)
		if err == nil {
			_, _ = fmt.Fprintf(w, "- %s\n", h.DisplayName())
		} else {
			_, _ = fmt.Fprintf(w, "- %s\n", line)
		}
	}
}

// printUsage prints the help text to w.
func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `
usage: known_hosts command [host]
  commands:
    ls      - List all known hosts
    rm      - Remove a host (supports --dry-run)
    search  - Search host in known hosts
    tui     - Interactive terminal UI
    help    - Show this message
    `)
}
