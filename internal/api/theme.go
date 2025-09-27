package api

type theme struct{}

// ANSI color codes based on theme
const (
	ColorReset      = "\033[0m"
	ColorLightGreen = "\033[38;2;69;241;194m" // #45F1C2
	ColorPink       = "\033[38;2;205;66;119m" // #CD4277
	ColorBlue       = "\033[38;2;12;160;216m" // #0CA0D8
	ColorTeal       = "\033[38;2;20;165;174m" // #14A5AE
)
