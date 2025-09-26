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

		{2000, "2.0K"},
		{1536, "1536"},
	}

	for _, test := range tests {
		result := formatSize(test.input)
		if result != test.expected {
			t.Errorf("formatSize(%d) = %s; expected %s", test.input, result, test.expected)
		}
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

	// The new format is: "<color>path<reset> <color>branch<reset>"
	// For non-existent directory, path should be the directory and branch should be "No Git"
	// Since context is "Unknown", no context section should be added
	if !strings.Contains(result, "/non/existent/directory") {
		t.Errorf("Result should contain directory path, got: %s", result)
	}
	if !strings.Contains(result, "No Git") {
		t.Errorf("Result should contain 'No Git', got: %s", result)
	}

	// Verify ANSI color codes are present
	if !strings.Contains(result, "\033[") {
		t.Errorf("Result should contain ANSI color codes, got: %s", result)
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

func TestContextColorLogic(t *testing.T) {
	tests := []struct {
		name                  string
		contextBytes          int
		expectedColorContains string
		description           string
	}{
		{"Low context", 100000, ColorLightGreen, "Should be green for low token count"},
		{"Just under 90k tokens", 359999, ColorLightGreen, "Should be green just under 90k tokens"}, // 89999.75 tokens
		{"Exactly 90k tokens", 360000, ColorLightGreen, "Should be green at exactly 90k tokens"},    // exactly 90000 tokens
		{"Just over 90k tokens", 360001, ColorPink, "Should be pink just over 90k tokens"},          // 90000.25 tokens
		{"High context", 500000, ColorPink, "Should be pink for high token count"},                  // 125k tokens
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a temp file with the specified byte count
			tmpFile, err := os.CreateTemp("", "test-context-*.json")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write content to achieve the desired byte count
			content := strings.Repeat("x", test.contextBytes)
			if _, err := tmpFile.WriteString(content); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Create status line with this file
			input := InputData{
				TranscriptPath: tmpFile.Name(),
				Workspace:      Workspace{CurrentDir: "/tmp"},
			}
			statusLine := NewStatusLine(input)
			result := statusLine.Generate()

			// Check if the expected color is present
			if !strings.Contains(result, test.expectedColorContains) {
				t.Errorf("%s: Expected color %s not found in result: %s",
					test.description, test.expectedColorContains, result)
			}

			// Additional verification: calculate token count to ensure our test is correct
			tokens := test.contextBytes / 4
			t.Logf("Test %s: %d bytes = %d tokens", test.name, test.contextBytes, tokens)
		})
	}
}

func TestCalculateTokenUsage(t *testing.T) {
	tests := []struct {
		name           string
		transcriptData []string
		expectedTokens int
		description    string
	}{
		{
			name: "No assistant messages",
			transcriptData: []string{
				`{"type": "user", "message": {"content": "Hello"}}`,
				`{"type": "system", "message": {"content": "System message"}}`,
			},
			expectedTokens: 0,
			description:    "Should return 0 when no assistant messages found",
		},
		{
			name: "Assistant message without usage",
			transcriptData: []string{
				`{"type": "assistant", "message": {"content": "Hello"}}`,
			},
			expectedTokens: 0,
			description:    "Should return 0 when assistant message has no usage info",
		},
		{
			name: "Assistant message with usage",
			transcriptData: []string{
				`{"type": "assistant", "message": {"usage": {"input_tokens": 100, "output_tokens": 50}}}`,
			},
			expectedTokens: 150,
			description:    "Should return sum of input and output tokens",
		},
		{
			name: "Assistant message with all token types",
			transcriptData: []string{
				`{"type": "assistant", "message": {"usage": {"input_tokens": 100, "cache_creation_input_tokens": 25, "cache_read_input_tokens": 75, "output_tokens": 50}}}`,
			},
			expectedTokens: 250,
			description:    "Should return sum of all token types",
		},
		{
			name: "Multiple assistant messages, uses most recent",
			transcriptData: []string{
				`{"type": "assistant", "message": {"usage": {"input_tokens": 100, "output_tokens": 50}}}`,
				`{"type": "user", "message": {"content": "Another message"}}`,
				`{"type": "assistant", "message": {"usage": {"input_tokens": 200, "output_tokens": 75}}}`,
			},
			expectedTokens: 275, // Most recent assistant message
			description:    "Should use the most recent assistant message with usage",
		},
		{
			name: "Invalid JSON mixed with valid",
			transcriptData: []string{
				`{"type": "assistant", "message": {"usage": {"input_tokens": 100, "output_tokens": 50}}}`,
				`{invalid json}`,
				`{"type": "assistant", "message": {"usage": {"input_tokens": 200, "output_tokens": 75}}}`,
			},
			expectedTokens: 275,
			description:    "Should skip invalid JSON and find valid usage",
		},
		{
			name: "Empty lines and whitespace",
			transcriptData: []string{
				`{"type": "assistant", "message": {"usage": {"input_tokens": 100, "output_tokens": 50}}}`,
				`   `,
				``,
				`  {"type": "assistant", "message": {"usage": {"input_tokens": 200, "output_tokens": 75}}}  `,
			},
			expectedTokens: 275,
			description:    "Should handle empty lines and whitespace correctly",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create temporary file with transcript data
			tmpFile, err := os.CreateTemp("", "test-transcript-*.jsonl")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write transcript data line by line
			for _, line := range test.transcriptData {
				if _, err := tmpFile.WriteString(line + "\n"); err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}
			}
			tmpFile.Close()

			// Create status line and test token calculation
			input := InputData{TranscriptPath: tmpFile.Name()}
			statusLine := NewStatusLine(input)
			result := statusLine.CalculateTokenUsage()

			if result != test.expectedTokens {
				t.Errorf("%s: Expected %d tokens, got %d",
					test.description, test.expectedTokens, result)
			}
		})
	}
}

func TestCalculateTokenUsageEdgeCases(t *testing.T) {
	// Test with non-existent file
	input := InputData{TranscriptPath: "/non/existent/file.jsonl"}
	statusLine := NewStatusLine(input)
	result := statusLine.CalculateTokenUsage()
	if result != 0 {
		t.Errorf("Expected 0 tokens for non-existent file, got %d", result)
	}

	// Test with empty transcript path
	input = InputData{TranscriptPath: ""}
	statusLine = NewStatusLine(input)
	result = statusLine.CalculateTokenUsage()
	if result != 0 {
		t.Errorf("Expected 0 tokens for empty transcript path, got %d", result)
	}

	// Test with empty file
	tmpFile, err := os.CreateTemp("", "test-empty-transcript-*.jsonl")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	input = InputData{TranscriptPath: tmpFile.Name()}
	statusLine = NewStatusLine(input)
	result = statusLine.CalculateTokenUsage()
	if result != 0 {
		t.Errorf("Expected 0 tokens for empty file, got %d", result)
	}
}
