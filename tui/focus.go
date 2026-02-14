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

// cycleFocus cycles focus between panels.
// forward=true: Video -> Search -> Notes -> Video
// forward=false: Video -> Notes -> Search -> Video
func (m *Model) cycleFocus(forward bool) {
	if forward {
		switch m.focus {
		case FocusVideo:
			m.focus = FocusSearch
		case FocusSearch:
			m.focus = FocusNotes
		case FocusNotes:
			m.focus = FocusVideo
		}
	} else {
		switch m.focus {
		case FocusVideo:
			m.focus = FocusNotes
		case FocusSearch:
			m.focus = FocusVideo
		case FocusNotes:
			m.focus = FocusSearch
		}
	}
}
