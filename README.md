# Known Hosts

A clone of [know_hosts](https://github.com/markmcconachie/known_hosts) write in [Go](https://golang.org)

Which supports multiple platforms, tested on Windows, Linux and macOS.

## Installation

```bash
go install github.com/jsfaint/known_hosts@latest
```

## Usage

```bash
$ known_hosts

usage: known_hosts command [host]
  commands:
    ls      - List all known hosts
    rm      - Remove a host (supports --dry-run)
    search  - Search host in known hosts
    tui     - Interactive terminal UI
    help    - Show this message
```

Dry-run example:

```bash
known_hosts rm github.com --dry-run
```

Special thanks to [markmcconachie](https://github.com/markmcconachie) about the original utility.
