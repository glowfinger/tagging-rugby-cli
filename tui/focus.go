package tui

// FocusTarget represents which panel currently has focus.
type FocusTarget int

const (
	// FocusVideo focuses the video panel.
	FocusVideo FocusTarget = iota
	// FocusSearch focuses the search input.
	FocusSearch
	// FocusNotes focuses the notes list.
	FocusNotes
)
