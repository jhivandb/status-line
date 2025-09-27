## Commands

### Build and Test

```bash
# Build the status line executable
go build

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestFormatSize

# Format code
go fmt ./...

# Clean module cache if needed
go clean -modcache
```

### Testing the Application

```bash
# Test with sample input
cat interface/input.json | ./status-line

# Build and test in one command
go build && cat interface/input.json | ./status-line
```

## Architecture

This is a Go CLI tool that generates status line information for Claude Code by processing JSON input from stdin.

### Core Architecture

- **Entry Point**: `main.go` reads JSON from stdin, instantiates sections, and orchestrates rendering
- **API Layer**: `internal/api/` contains shared contracts:
  - `types.go`: Input/output structures matching Claude Code's hook format
  - `section.go`: Interface that all status sections must implement
  - `theme.go`: ANSI color constants for consistent theming
- **Sections**: `internal/sections/` contains modular status components:
  - `context.go`: Token usage calculation and formatting
  - `git.go`: Git branch detection with multiple fallback strategies
  - `path.go`: Current directory display with home directory shortening

### Section Interface Pattern

All status line components implement the `Section` interface:

```go
type Section interface {
    Render() string  // Returns formatted output with colors
}
```

Each section is self-contained and responsible for:

- Fetching its own data based on input
- Applying appropriate ANSI colors
- Rendering its portion of the status line

### Development Notes

**Current Refactoring State**: The project has duplicate functionality during migration. When adding features, prefer implementing in the new sections-based architecture in `internal/sections/`.

**Module Info**:

- Module: `github.com/jhivandb/status-line`
- Go version: 1.24.0
- Dependencies: Go standard library only

# important-instruction-reminders

Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.
