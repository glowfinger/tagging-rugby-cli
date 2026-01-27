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

// NotesList renders the notes list component.
// It displays a scrollable list of notes and tackles sorted by timestamp.
func NotesList(state NotesListState, width, height int) string {
	if len(state.Items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Lavender).
			Italic(true).
			Padding(1, 2)
		return emptyStyle.Render("No notes or tackles for this video")
	}

	// Calculate visible area (leave room for header)
	visibleHeight := height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	// Adjust scroll offset to keep selected item visible
	if state.SelectedIndex < state.ScrollOffset {
		state.ScrollOffset = state.SelectedIndex
	} else if state.SelectedIndex >= state.ScrollOffset+visibleHeight {
		state.ScrollOffset = state.SelectedIndex - visibleHeight + 1
	}

	// Build the list
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Bold(true).
		Width(width)
	lines = append(lines, headerStyle.Render(fmt.Sprintf(" Notes & Tackles (%d items)", len(state.Items))))

	// Render visible items
	for i := state.ScrollOffset; i < len(state.Items) && i < state.ScrollOffset+visibleHeight; i++ {
		item := state.Items[i]
		isSelected := i == state.SelectedIndex
		lines = append(lines, renderListItem(item, isSelected, width))
	}

	return strings.Join(lines, "\n")
}

// renderListItem renders a single list item.
func renderListItem(item ListItem, selected bool, width int) string {
	// Timestamp
	timeStr := formatTime(item.TimestampSeconds)

	// Type badge
	var typeBadge string
	var badgeStyle lipgloss.Style
	if item.Type == ItemTypeNote {
		badgeStyle = lipgloss.NewStyle().
			Background(styles.Cyan).
			Foreground(styles.DeepPurple).
			Bold(true).
			Padding(0, 1)
		typeBadge = badgeStyle.Render("note")
	} else {
		badgeStyle = lipgloss.NewStyle().
			Background(styles.Pink).
			Foreground(styles.DeepPurple).
			Bold(true).
			Padding(0, 1)
		typeBadge = badgeStyle.Render("tackle")
	}

	// Star symbol for starred items
	starStr := ""
	if item.Starred {
		starStr = " ★"
	}

	// Build prefix with timestamp and badge
	prefix := fmt.Sprintf(" %s %s%s ", timeStr, typeBadge, starStr)
	prefixWidth := lipgloss.Width(prefix)

	// Calculate remaining width for text preview
	textWidth := width - prefixWidth - 2 // 2 for padding
	if textWidth < 10 {
		textWidth = 10
	}

	// Truncate text if needed
	text := item.Text
	if len(text) > textWidth {
		text = text[:textWidth-3] + "..."
	}

	// Full line content
	content := prefix + text

	// Apply style based on selection
	var lineStyle lipgloss.Style
	if selected {
		lineStyle = lipgloss.NewStyle().
			Background(styles.BrightPurple).
			Foreground(styles.LightLavender).
			Bold(true).
			Width(width)
	} else {
		lineStyle = lipgloss.NewStyle().
			Foreground(styles.LightLavender).
			Width(width)
	}

	return lineStyle.Render(content)
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
