package api

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
