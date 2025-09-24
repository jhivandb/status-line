package main

import (
	"os"
	"strings"
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1572864, "1.5M"},
		{2097152, "2.0M"},
	}

	for _, test := range tests {
		result := formatSize(test.input)
		if result != test.expected {
			t.Errorf("formatSize(%d) = %s; expected %s", test.input, result, test.expected)
		}
	}
}

func TestStatusLineCalculateContextSize(t *testing.T) {
	// Test with non-existent file
	input := InputData{TranscriptPath: "/non/existent/file.json"}
	statusLine := NewStatusLine(input)
	result := statusLine.CalculateContextSize()
	if result != "Unknown" {
		t.Errorf("Expected 'Unknown' for non-existent file, got %s", result)
	}

	// Test with empty path
	input = InputData{TranscriptPath: ""}
	statusLine = NewStatusLine(input)
	result = statusLine.CalculateContextSize()
	if result != "Unknown" {
		t.Errorf("Expected 'Unknown' for empty path, got %s", result)
	}

	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-transcript-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testContent := `{"test": "content", "size": "small"}`
	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test with actual file
	input = InputData{TranscriptPath: tmpFile.Name()}
	statusLine = NewStatusLine(input)
	result = statusLine.CalculateContextSize()
	expected := formatSize(len(testContent))
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestStatusLineGetGitBranch(t *testing.T) {
	// Test with non-existent directory
	input := InputData{
		Workspace: Workspace{CurrentDir: "/non/existent/directory"},
	}
	statusLine := NewStatusLine(input)
	result := statusLine.GetGitBranch()
	if result != "No Git" {
		t.Errorf("Expected 'No Git' for non-existent directory, got %s", result)
	}

	// Test with empty workspace
	input = InputData{}
	statusLine = NewStatusLine(input)
	result = statusLine.GetGitBranch()
	if result != "No Git" {
		t.Errorf("Expected 'No Git' for empty workspace, got %s", result)
	}

	// Create temporary directory without git
	tmpDir, err := os.MkdirTemp("", "test-no-git-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	input = InputData{
		Workspace: Workspace{CurrentDir: tmpDir},
	}
	statusLine = NewStatusLine(input)
	result = statusLine.GetGitBranch()
	if result != "No Git" {
		t.Errorf("Expected 'No Git' for directory without git, got %s", result)
	}
}

func TestStatusLineGenerate(t *testing.T) {
	input := InputData{
		TranscriptPath: "/non/existent/file.json",
		Workspace:      Workspace{CurrentDir: "/non/existent/directory"},
	}
	statusLine := NewStatusLine(input)
	result := statusLine.Generate()

	expected := "Context: Unknown | Branch: No Git"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Verify format structure
	if !strings.Contains(result, "Context:") || !strings.Contains(result, "Branch:") {
		t.Errorf("Result does not contain expected format markers: %s", result)
	}
}

func TestNewStatusLine(t *testing.T) {
	input := InputData{
		TranscriptPath: "/test/path",
		Workspace:      Workspace{CurrentDir: "/test/workspace"},
	}

	statusLine := NewStatusLine(input)
	if statusLine == nil {
		t.Fatal("NewStatusLine returned nil")
	}

	if statusLine.input.TranscriptPath != "/test/path" {
		t.Errorf("Expected TranscriptPath '/test/path', got %s", statusLine.input.TranscriptPath)
	}

	if statusLine.input.Workspace.CurrentDir != "/test/workspace" {
		t.Errorf("Expected CurrentDir '/test/workspace', got %s", statusLine.input.Workspace.CurrentDir)
	}
}