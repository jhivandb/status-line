package api

// Section defines a part of the status line that displays some information.
// It needs to be self contained and be able to fetch its own data
type Section interface {
	// Render returns the final output to be displayed
	Render() string
}
