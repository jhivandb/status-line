# Project Overview

This is a Go-based status line extension for Claude Code that processes JSON input from stdin and generates formatted status information including context size and git branch details.

## Build and Test Commands

```bash
# Build the project
go build

# Run tests
go test

# Run tests with verbose output
go test -v

# Run a specific test
go test -run TestFormatSize

# Format code
go fmt ./...

# Clean module cache
go clean -modcache
```

## Architecture

The project follows a clean architecture pattern:

### Core Components

- **`main.go`**: Entry point and primary business logic
  - `StatusLine` struct handles calculation and formatting
  - JSON input parsing from stdin
  - Context size calculation from transcript files
  - Git branch detection with fallback methods

- **`internal/api/`**: API contracts and data structures
  - `types.go`: Complete JSON input structure matching Claude Code's hook format
  - `section.go`: Interface definitions for future extensibility (currently stub)

### Input Processing

The application expects JSON input via stdin matching the structure defined in `internal/api/types.go`. Example input format is available in `interface/input.json`.

Key input fields:

- `transcript_path`: Path to Claude Code transcript file for context size calculation
- `workspace.current_dir` or `cwd`: Working directory for git operations
- Additional metadata: session_id, model info, cost tracking, etc.

### Git Integration

Git branch detection uses multiple fallback strategies:

1. `git branch --show-current` (primary)
2. `git symbolic-ref --short HEAD` (fallback)
3. `git describe --tags --exact-match` (for tagged commits)
4. `git rev-parse --short HEAD` (detached HEAD)

Returns "No Git" if directory doesn't exist or isn't a git repository.

### File Size Formatting

Context size calculation reads entire transcript files and formats byte counts:

- < 1KB: Shows exact bytes (e.g., "500B")
- < 1MB: Shows kilobytes (e.g., "1.5K")
- >= 1MB: Shows megabytes (e.g., "2.0M")

## Testing Strategy

Test files follow Go conventions with comprehensive coverage:

- Unit tests for all public methods
- Temporary file creation for integration testing
- Edge case handling (non-existent files, empty directories)
- Format validation for output strings

## Recent Architecture Changes

The project recently refactored from `internal/types` to `internal/api` package structure. When working with imports, ensure you're using:

- `github.com/jhivandb/status-line/internal/api` (current)
- Not `github.com/jhivandb/status-line/internal/types` (deprecated)

## Module Information

- Module: `github.com/jhivandb/status-line`
- Go version: 1.24.0
- No external dependencies (uses only Go standard library)

### DO NOT WRITE CODE, ONLY Assist with advice and snippets ###
