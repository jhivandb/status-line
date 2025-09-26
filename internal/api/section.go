package api

// Section defines a part of the status line that displays some information.
// It needs to be self contained and be able to fetch its own data
type Section interface {
	// getStyles returns the color for that segment
	getStyles(in InputData) string
	// getText returns the output including symbols and text
	getText(in InputData) string
	// Render returns the final output to be displayed
	Render() string
}
