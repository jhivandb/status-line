package sections

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jhivandb/status-line/internal/api"
)

type Context struct {
	config string
	In     api.InputData
}

func (ctx *Context) Render() string {
	tokens := calculateTokenUsage(ctx.In)
	contextColor := api.ColorLightGreen
	if tokens > 90000 {
		contextColor = api.ColorPink
	}

	return fmt.Sprintf("%s%s", contextColor, formatSize(tokens))
}

// calculateTokenUsage reads the transcript file from end to beginning,
// looking for the most recent assistant message with token usage information.
// Returns the total token count (input + cache creation + cache read + output tokens).
func calculateTokenUsage(in api.InputData) int {
	if in.TranscriptPath == "" {
		return 0
	}

	file, err := os.Open(in.TranscriptPath)
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

		var entry api.TranscriptEntry
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
