package main

import (
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
	HookEventName   string `json:"hook_event_name"`
	SessionID       string `json:"session_id"`
	TranscriptPath  string `json:"transcript_path"`
	CWD             string `json:"cwd"`
	Model           Model  `json:"model"`
	Workspace       Workspace `json:"workspace"`
	Version         string `json:"version"`
	OutputStyle     OutputStyle `json:"output_style"`
	Cost            Cost   `json:"cost"`
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
	TotalCostUSD         float64 `json:"total_cost_usd"`
	TotalDurationMS      int64   `json:"total_duration_ms"`
	TotalAPIDurationMS   int64   `json:"total_api_duration_ms"`
	TotalLinesAdded      int     `json:"total_lines_added"`
	TotalLinesRemoved    int     `json:"total_lines_removed"`
}

// StatusLine handles the calculation and formatting of status information
type StatusLine struct {
	input InputData
}

// NewStatusLine creates a new StatusLine instance
func NewStatusLine(input InputData) *StatusLine {
	return &StatusLine{input: input}
}

// CalculateContextSize reads the transcript file and calculates its character count
func (s *StatusLine) CalculateContextSize() string {
	if s.input.TranscriptPath == "" {
		return "Unknown"
	}

	// Check if file exists and is readable
	if _, err := os.Stat(s.input.TranscriptPath); os.IsNotExist(err) {
		return "Unknown"
	}

	file, err := os.Open(s.input.TranscriptPath)
	if err != nil {
		return "Unknown"
	}
	defer file.Close()

	// Read the entire file to count characters
	data, err := io.ReadAll(file)
	if err != nil {
		return "Unknown"
	}

	size := len(data)
	return formatSize(size)
}

// formatSize formats byte count into human-readable units
func formatSize(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
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

	return branch
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

// Generate creates the formatted status line
func (s *StatusLine) Generate() string {
	contextSize := s.CalculateContextSize()
	gitBranch := s.GetGitBranch()

	return fmt.Sprintf("Context: %s | Branch: %s", contextSize, gitBranch)
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