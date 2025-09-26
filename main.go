package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InputData represents the JSON input structure from Claude Code
type InputData struct {
	HookEventName  string      `json:"hook_event_name"`
	SessionID      string      `json:"session_id"`
	TranscriptPath string      `json:"transcript_path"`
	CWD            string      `json:"cwd"`
	Model          Model       `json:"model"`
	Workspace      Workspace   `json:"workspace"`
	Version        string      `json:"version"`
	OutputStyle    OutputStyle `json:"output_style"`
	Cost           Cost        `json:"cost"`
}

type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMS    int64   `json:"total_duration_ms"`
	TotalAPIDurationMS int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

// TranscriptEntry represents a line in the transcript file
type TranscriptEntry struct {
	Type    string  `json:"type"`
	Message Message `json:"message"`
}

// Message represents the message content in a transcript entry
type Message struct {
	Usage Usage `json:"usage"`
}

// Usage represents token usage information in assistant messages
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

// StatusLine handles the calculation and formatting of status information
type StatusLine struct {
	input InputData
}

// ANSI color codes based on theme
const (
	ColorReset      = "\033[0m"
	ColorLightGreen = "\033[38;2;69;241;194m" // #45F1C2
	ColorPink       = "\033[38;2;205;66;119m" // #CD4277
	ColorBlue       = "\033[38;2;12;160;216m" // #0CA0D8
	ColorTeal       = "\033[38;2;20;165;174m" // #14A5AE
)

// NewStatusLine creates a new StatusLine instance
func NewStatusLine(input InputData) *StatusLine {
	return &StatusLine{input: input}
}

// CalculateTokenUsage reads the transcript file from end to beginning,
// looking for the most recent assistant message with token usage information.
// Returns the total token count (input + cache creation + cache read + output tokens).
func (s *StatusLine) CalculateTokenUsage() int {
	if s.input.TranscriptPath == "" {
		return 0
	}

	file, err := os.Open(s.input.TranscriptPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Read all lines into a slice
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return 0
	}

	// Iterate from last line to first line
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Check if this line contains token usage information
		if entry.Type == "assistant" {
			usage := entry.Message.Usage
			// Check if usage information is present (at least one field is non-zero)
			if usage.InputTokens > 0 || usage.CacheCreationInputTokens > 0 ||
				usage.CacheReadInputTokens > 0 || usage.OutputTokens > 0 {
				return usage.InputTokens + usage.CacheCreationInputTokens +
					usage.CacheReadInputTokens + usage.OutputTokens
			}
		}
	}

	return 0
}

func formatSize(size int) string {
	const (
		K = 1000
	)

	switch {
	case size >= K+1000:
		return fmt.Sprintf("%.1fK", float64(size)/K)
	default:
		return fmt.Sprintf("%d", size)
	}
}

// GetGitBranch executes git command to get the current branch
func (s *StatusLine) GetGitBranch() string {
	workDir := s.input.Workspace.CurrentDir
	if workDir == "" {
		workDir = s.input.CWD
	}

	if workDir == "" {
		return "No Git"
	}

	// Check if directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return "No Git"
	}

	// Check if it's a git repository by looking for .git directory
	gitDir := filepath.Join(workDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "No Git"
	}

	// Execute git command to get current branch
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err != nil {
		// Try alternative method for detached HEAD or older git versions
		return s.getGitBranchFallback(workDir)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return s.getGitBranchFallback(workDir)
	}

	return fmt.Sprintf("\uf09b \ue0a0%s", branch)
}

// getGitBranchFallback tries alternative methods to get git branch info
func (s *StatusLine) getGitBranchFallback(workDir string) string {
	// Try git symbolic-ref for current branch
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = workDir

	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch
		}
	}

	// Try git describe for detached HEAD
	cmd = exec.Command("git", "describe", "--tags", "--exact-match")
	cmd.Dir = workDir

	output, err = cmd.Output()
	if err == nil {
		tag := strings.TrimSpace(string(output))
		if tag != "" {
			return fmt.Sprintf("tag:%s", tag)
		}
	}

	// Try git rev-parse for commit hash
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = workDir

	output, err = cmd.Output()
	if err == nil {
		hash := strings.TrimSpace(string(output))
		if hash != "" {
			return fmt.Sprintf("detached:%s", hash)
		}
	}

	return "No Git"
}

// GetCurrentPath returns the working directory path, formatted like terminal
func (s *StatusLine) GetCurrentPath() string {
	workDir := s.input.Workspace.CurrentDir
	if workDir == "" {
		workDir = s.input.CWD
	}

	if workDir == "" {
		return "~"
	}

	// Get home directory to replace with ~
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(workDir, homeDir) {
		workDir = "\uf07b ~" + workDir[len(homeDir):]
	}

	return workDir
}

// Generate creates the formatted status line matching terminal style
func (s *StatusLine) Generate() string {
	actualTokens := s.CalculateTokenUsage()
	gitBranch := s.GetGitBranch()
	currentPath := s.GetCurrentPath()

	// Use actual token usage if available, otherwise fall back to file size estimation

	// Choose context color based on token count
	contextColor := ColorLightGreen
	if actualTokens > 90000 {
		contextColor = ColorPink
	}

	// Format like terminal: ~/path/to/dir main [context]
	statusLine := fmt.Sprintf("%s %s%s%s %s%s%s",
		ColorReset,
		ColorBlue, currentPath, ColorReset,
		ColorTeal, gitBranch, ColorReset)

	// Add context info with appropriate coloring

	statusLine += fmt.Sprintf(" %s%s/90k%s",
		contextColor, formatSize(actualTokens), ColorReset)

	return statusLine
}

func main() {
	// Read JSON input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON input
	var inputData InputData
	if err := json.Unmarshal(input, &inputData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Create status line and generate output
	statusLine := NewStatusLine(inputData)
	output := statusLine.Generate()

	fmt.Println(output)
}
