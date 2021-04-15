/*
Package main implements an utility for manage the ssh known_hosts file.
A clone of https://github.com/markmcconachie/known_hosts write in Go.
Aimming to support Windows/Linux/macOS.
*/
package main

import (
	"fmt"
	"os"
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
)

func checkArgs(num int) {
	if len(os.Args) != num {
		fmt.Println("Invalid parameter")
		printUsage()
	}
}

func parseArgs() (opt opts) {
	if len(os.Args) < 2 {
		printUsage()
	}

	switch os.Args[1] {
	case cmdRemove:
		checkArgs(3)
		opt.operation = cmdRemove
		opt.host = os.Args[2]
	case cmdList:
		checkArgs(2)
		opt.operation = cmdList
	case cmdSearch:
		checkArgs(3)
		opt.operation = cmdSearch
		opt.host = os.Args[2]
	case cmdHelp:
		printUsage()
	default:
		fmt.Println("Invalid parameter")
		printUsage()
	}

	return opt
}

func deleteHost(hosts []string, host string) {
	fmt.Println("Removing host: ", host)
	hosts = Delete(hosts, host)
	if err := SaveFile(hosts); err != nil {
		fmt.Println("Delete host fail")
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
    help    - Show this message
    `)

	os.Exit(1)
}

func main() {
	if !Exists() {
		return
	}

	opt := parseArgs()

	hosts := ReadFile()

	switch opt.operation {
	case cmdRemove:
		deleteHost(hosts, opt.host)
	case cmdList:
		listHost(hosts)
	case cmdSearch:
		searchHost(hosts, opt.host)
	}
}
