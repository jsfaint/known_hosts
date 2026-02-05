# Contributing to known_hosts

Thank you for your interest in contributing to `known_hosts`!

---

## Development Setup

### Prerequisites

- Go 1.16 or higher
- Git
- `golangci-lint` for linting

### Clone and Build

```bash
git clone git@github.com:jsfaint/known_hosts.git
cd known_hosts
go build
```

### Running Tests

```bash
go test -v ./...
```

---

## Code Standards

### Project Characteristics

This is a **simple Go CLI tool** with:
- Single package structure (`package main`)
- File-based storage (operates on `~/.ssh/known_hosts`)
- Cross-platform support (Windows, Linux, macOS)
- Direct output via `fmt.Println()` (no logging framework)

### File Organization

```
known_hosts/
├── main.go              # CLI entry point, argument parsing
├── file.go              # File I/O operations
├── host.go              # Data structures and parsing
├── *_test.go            # Tests
└── .github/workflows/   # CI/CD
```

### Code Style

**Must pass**:
- `golangci-lint run` (enforced by CI)
- `go test -v ./...` (enforced by CI)

**Conventions**:
- **Exported functions**: PascalCase with comments
- **Unexported functions**: camelCase
- **Package**: `package main` (single binary)
- **Errors**: Last return value, never panic
- **Comments**: Exported functions must have documentation

### Cross-Platform Compatibility

**CRITICAL**: This tool runs on Windows, Linux, and macOS.

**Do's**:
```go
// ✅ CORRECT - Cross-platform paths
path := filepath.Join(home, ".ssh", "known_hosts")

// ✅ CORRECT - Cross-platform home directory
home, err := os.UserHomeDir()

// ✅ CORRECT - Platform detection
if runtime.GOOS == "windows" {
    // Windows-specific code
}
```

**Don'ts**:
```go
// ❌ WRONG - Hardcoded Unix path separator
path := home + "/.ssh/known_hosts"

// ❌ WRONG - Unix-only home directory
home := os.Getenv("HOME")
```

### Line Endings

Handle `\r\n` (Windows) vs `\n` (Unix):

```go
func getLinebreak() string {
    if runtime.GOOS == "windows" {
        return "\r\n"
    }
    return "\n"
}
```

---

## Testing

### Test Requirements

- All functions must have corresponding tests
- Aim for >80% coverage on core logic
- Test both success and error paths

### Example Test

```go
func TestNewHost(t *testing.T) {
    // Test valid input
    host, err := NewHost("github.com ssh-rsa AAAAB3...")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }

    // Test invalid input
    _, err = NewHost("invalid")
    if err == nil {
        t.Error("Expected error for invalid input")
    }
}
```

---

## Error Handling

### Philosophy

**Fail fast and exit**:
- Print errors immediately
- Exit with non-zero status
- No silent failures
- No error recovery/retry logic

### Patterns

```go
// ✅ Return errors, don't panic
func SaveFile(input []string) error {
    // ...
    return ioutil.WriteFile(name, []byte(str), 0644)
}

// ❌ Don't panic for expected errors
func SaveFile(input []string) {
    if err != nil {
        panic(err)  // WRONG
    }
}
```

---

## Adding New Features

### When to Add New Files

**Create a new `.go` file when**:
- Adding a new data structure (like `host.go`)
- Adding new operations (like `file.go`)
- Separating concerns for better testability

**Keep in `main.go`**:
- CLI argument parsing
- Command routing
- Usage information

### Example: Adding a New Command

1. Add command constant to `main.go`:
```go
const (
    cmdList   = "ls"
    cmdRemove = "rm"
    cmdAdd    = "add"  // New command
)
```

2. Add operation logic to `file.go` or new file
3. Add tests in `{file}_test.go`
4. Update `printUsage()` with command description

---

## Commit Messages

Follow conventional commits:

```
type(scope): description

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- test: Tests
- chore: Maintenance

Example:
feat(file): Add search function for known hosts
fix: Handle Windows line endings correctly
```

---

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linter:
   ```bash
   go test -v ./...
   golangci-lint run
   ```
5. Commit your changes (conventional commits)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

**CI Checks**:
- All PRs must pass `golangci-lint`
- All PRs must pass tests

---

## Questions?

Feel free to open an issue for questions or discussions.
