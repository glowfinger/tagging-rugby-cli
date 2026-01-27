// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ListItemType represents the type of list item.
type ListItemType int

const (
	// ItemTypeNote represents a note item.
	ItemTypeNote ListItemType = iota
	// ItemTypeTackle represents a tackle item.
	ItemTypeTackle
)

// ListItem represents a note or tackle in the list.
type ListItem struct {
	// ID is the database ID of the item
	ID int64
	// Type is either note or tackle
	Type ListItemType
	// TimestampSeconds is the position in the video
	TimestampSeconds float64
	// Text is the note text or tackle description
	Text string
	// Starred indicates if this is a starred item (tackles only)
	Starred bool
	// Category is the optional category
	Category string
	// Player is the optional player name
	Player string
	// Team is the optional team name
	Team string
}

// NotesListState holds the state for the notes list component.
type NotesListState struct {
	// Items is the list of notes and tackles
	Items []ListItem
	// SelectedIndex is the currently selected item index
	SelectedIndex int
	// ScrollOffset is the scroll position
	ScrollOffset int
}

// tableRows is the fixed number of rows in the table (excluding header).
const tableRows = 10

// NotesList renders the notes list component as a fixed 10-row table at the bottom.
// It displays notes and tackles sorted by timestamp.
// The currentTimePos parameter is used to auto-scroll to show notes near the current video timestamp.
func NotesList(state NotesListState, width, height int, currentTimePos float64) string {
	// Build the table
	var lines []string

	// Table header
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Bold(true).
		Underline(true)

	// Column widths (ID: 6, Timestamp: 8, Category: 12, Text: rest)
	idWidth := 6
	timeWidth := 8
	catWidth := 12
	textWidth := width - idWidth - timeWidth - catWidth - 8 // 8 for spacing/borders
	if textWidth < 10 {
		textWidth = 10
	}

	// Build header row
	header := fmt.Sprintf(" %-*s %-*s %-*s %-*s",
		idWidth, "ID",
		timeWidth, "Time",
		catWidth, "Category",
		textWidth, "Text")
	lines = append(lines, headerStyle.Render(header))

	if len(state.Items) == 0 {
		// Empty state - show placeholder rows
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Purple).
			Italic(true)
		emptyRow := emptyStyle.Render(fmt.Sprintf(" %-*s", width-2, "No notes or tackles for this video"))
		lines = append(lines, emptyRow)
		// Fill remaining rows with empty space
		for i := 1; i < tableRows; i++ {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}

	// Auto-scroll to show notes near current video timestamp
	state.scrollToCurrentTime(currentTimePos)

	// Adjust scroll offset to keep selected item visible within the 10 rows
	if state.SelectedIndex < state.ScrollOffset {
		state.ScrollOffset = state.SelectedIndex
	} else if state.SelectedIndex >= state.ScrollOffset+tableRows {
		state.ScrollOffset = state.SelectedIndex - tableRows + 1
	}

	// Ensure scroll offset doesn't go negative or beyond items
	if state.ScrollOffset < 0 {
		state.ScrollOffset = 0
	}
	maxOffset := len(state.Items) - tableRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if state.ScrollOffset > maxOffset {
		state.ScrollOffset = maxOffset
	}

	// Render exactly 10 rows
	for row := 0; row < tableRows; row++ {
		itemIndex := state.ScrollOffset + row
		if itemIndex < len(state.Items) {
			item := state.Items[itemIndex]
			isSelected := itemIndex == state.SelectedIndex
			lines = append(lines, renderTableRow(item, isSelected, idWidth, timeWidth, catWidth, textWidth, width))
		} else {
			// Empty row
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// scrollToCurrentTime adjusts the scroll offset to show notes near the current timestamp.
func (s *NotesListState) scrollToCurrentTime(currentTimePos float64) {
	if len(s.Items) == 0 {
		return
	}

	// Find the first item that is at or after the current time
	nearestIndex := 0
	for i, item := range s.Items {
		if item.TimestampSeconds >= currentTimePos {
			nearestIndex = i
			break
		}
		nearestIndex = i // Keep track of the last item if all are before current time
	}

	// Center the view around the nearest item if not already selected
	if s.SelectedIndex < s.ScrollOffset || s.SelectedIndex >= s.ScrollOffset+tableRows {
		// Only auto-scroll if selection is out of view - position nearest item in upper third
		targetOffset := nearestIndex - tableRows/3
		if targetOffset < 0 {
			targetOffset = 0
		}
		maxOffset := len(s.Items) - tableRows
		if maxOffset < 0 {
			maxOffset = 0
		}
		if targetOffset > maxOffset {
			targetOffset = maxOffset
		}
		s.ScrollOffset = targetOffset
	}
}

// renderTableRow renders a single table row.
func renderTableRow(item ListItem, selected bool, idWidth, timeWidth, catWidth, textWidth, fullWidth int) string {
	// Format ID with star symbol if starred
	idStr := fmt.Sprintf("%d", item.ID)
	if item.Starred {
		idStr = "★" + idStr
	}

	// Format timestamp
	timeStr := formatTime(item.TimestampSeconds)

	// Get category (or type badge for tackles)
	catStr := item.Category
	if item.Type == ItemTypeTackle && catStr == "" {
		catStr = "tackle"
	}

	// Truncate text if needed
	text := item.Text
	if len(text) > textWidth {
		text = text[:textWidth-3] + "..."
	}

	// Build row content
	content := fmt.Sprintf(" %-*s %-*s %-*s %-*s",
		idWidth, truncateStr(idStr, idWidth),
		timeWidth, timeStr,
		catWidth, truncateStr(catStr, catWidth),
		textWidth, text)

	// Apply style based on selection
	var lineStyle lipgloss.Style
	if selected {
		lineStyle = lipgloss.NewStyle().
			Background(styles.BrightPurple).
			Foreground(styles.LightLavender).
			Bold(true).
			Width(fullWidth)
	} else {
		lineStyle = lipgloss.NewStyle().
			Foreground(styles.LightLavender).
			Width(fullWidth)
	}

	return lineStyle.Render(content)
}

// truncateStr truncates a string to maxLen characters.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// MoveUp moves the selection up in the list.
func (s *NotesListState) MoveUp() {
	if s.SelectedIndex > 0 {
		s.SelectedIndex--
	}
}

// MoveDown moves the selection down in the list.
func (s *NotesListState) MoveDown() {
	if s.SelectedIndex < len(s.Items)-1 {
		s.SelectedIndex++
	}
}

// GetSelectedItem returns the currently selected item, or nil if list is empty.
func (s *NotesListState) GetSelectedItem() *ListItem {
	if len(s.Items) == 0 || s.SelectedIndex < 0 || s.SelectedIndex >= len(s.Items) {
		return nil
	}
	return &s.Items[s.SelectedIndex]
}
