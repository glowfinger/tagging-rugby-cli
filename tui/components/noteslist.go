// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
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

// NotesList renders the notes list component as a dynamically-sized table.
// It displays notes and tackles sorted by timestamp.
// The visible row count is derived from the height parameter (height - 1 for the header row).
// The currentTimePos parameter is used to auto-scroll to show notes near the current video timestamp.
func NotesList(state NotesListState, width, height int, currentTimePos float64, matches []int, currentMatch int, query string) string {
	// Compute visible rows from height (subtract 1 for header row)
	visibleRows := height - 1
	if visibleRows <= 0 {
		return ""
	}

	// Build the table
	var lines []string

	// Table header
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Bold(true).
		Underline(true)

	// Column widths (#: 5, ID: 6, Timestamp: 9 for H:MM:SS, Category: 12, Text: rest)
	rowWidth := 5
	idWidth := 6
	timeWidth := 9
	catWidth := 12
	textWidth := width - rowWidth - idWidth - timeWidth - catWidth - 10 // 10 for spacing/borders
	if textWidth < 10 {
		textWidth = 10
	}

	// Build header row
	header := fmt.Sprintf(" %*s %-*s %-*s %-*s %-*s",
		rowWidth, "Row",
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
		for i := 1; i < visibleRows; i++ {
			lines = append(lines, "")
		}
		return strings.Join(lines, "\n")
	}

	// Auto-scroll to show notes near current video timestamp
	state.scrollToCurrentTime(currentTimePos, visibleRows)

	// Adjust scroll offset to keep selected item visible within visible rows
	if state.SelectedIndex < state.ScrollOffset {
		state.ScrollOffset = state.SelectedIndex
	} else if state.SelectedIndex >= state.ScrollOffset+visibleRows {
		state.ScrollOffset = state.SelectedIndex - visibleRows + 1
	}

	// Ensure scroll offset doesn't go negative or beyond items
	if state.ScrollOffset < 0 {
		state.ScrollOffset = 0
	}
	maxOffset := len(state.Items) - visibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if state.ScrollOffset > maxOffset {
		state.ScrollOffset = maxOffset
	}

	// Build a set of match indices for O(1) lookup
	matchSet := make(map[int]bool, len(matches))
	for _, idx := range matches {
		matchSet[idx] = true
	}
	currentMatchIdx := -1
	if len(matches) > 0 && currentMatch >= 0 && currentMatch < len(matches) {
		currentMatchIdx = matches[currentMatch]
	}

	// Render visible rows
	for row := 0; row < visibleRows; row++ {
		itemIndex := state.ScrollOffset + row
		if itemIndex < len(state.Items) {
			item := state.Items[itemIndex]
			isSelected := itemIndex == state.SelectedIndex
			rowNum := itemIndex + 1
			isMatch := matchSet[itemIndex]
			isCurrentMatch := itemIndex == currentMatchIdx
			lines = append(lines, renderTableRow(item, isSelected, isMatch, isCurrentMatch, rowNum, rowWidth, idWidth, timeWidth, catWidth, textWidth, width, query))
		} else {
			// Empty row
			lines = append(lines, "")
		}
	}

	return strings.Join(lines, "\n")
}

// scrollToCurrentTime adjusts the scroll offset to show notes near the current timestamp.
func (s *NotesListState) scrollToCurrentTime(currentTimePos float64, visibleRows int) {
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
	if s.SelectedIndex < s.ScrollOffset || s.SelectedIndex >= s.ScrollOffset+visibleRows {
		// Only auto-scroll if selection is out of view - position nearest item in upper third
		targetOffset := nearestIndex - visibleRows/3
		if targetOffset < 0 {
			targetOffset = 0
		}
		maxOffset := len(s.Items) - visibleRows
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
// When query is non-empty and the row is a match, the matching substring is highlighted
// inline rather than coloring the whole row. Matched rows get a subtle background.
func renderTableRow(item ListItem, selected, isMatch, isCurrentMatch bool, rowNum, rowWidth, idWidth, timeWidth, catWidth, textWidth, fullWidth int, query string) string {
	// Format row number: right-aligned, no # prefix (e.g., "  1", " 12", "123")
	rowStr := fmt.Sprintf("%*d", rowWidth, rowNum)

	// Format ID with star symbol if starred
	idStr := fmt.Sprintf("%d", item.ID)
	if item.Starred {
		idStr = "★" + idStr
	}

	// Format timestamp
	timeStr := timeutil.FormatTime(item.TimestampSeconds)

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

	// Determine highlight color for matching substrings
	var highlightBg lipgloss.Color
	if isCurrentMatch {
		highlightBg = styles.Pink
	} else if isMatch {
		highlightBg = styles.Amber
	}

	// Choose base text style based on state
	var baseStyle lipgloss.Style
	if selected && !isMatch && !isCurrentMatch {
		baseStyle = lipgloss.NewStyle().
			Background(styles.BrightPurple).
			Foreground(styles.LightLavender).
			Bold(true)
	} else if isMatch || isCurrentMatch {
		// Matched rows get subtle background
		baseStyle = lipgloss.NewStyle().
			Background(styles.MatchBg).
			Foreground(styles.LightLavender)
	} else {
		baseStyle = lipgloss.NewStyle().
			Foreground(styles.LightLavender)
	}

	// Helper to render a field with inline query highlighting
	renderField := func(s string, fieldWidth int) string {
		truncated := truncateStr(s, fieldWidth)
		padded := fmt.Sprintf("%-*s", fieldWidth, truncated)
		if query != "" && (isMatch || isCurrentMatch) {
			return highlightSubstring(padded, query, baseStyle, highlightBg)
		}
		return baseStyle.Render(padded)
	}

	// Build row with inline highlighting per field
	space := baseStyle.Render(" ")
	row := space +
		renderField(rowStr, rowWidth) + space +
		renderField(idStr, idWidth) + space +
		renderField(timeStr, timeWidth) + space +
		renderField(catStr, catWidth) + space +
		renderField(text, textWidth)

	// Pad to full width
	rowVisW := lipgloss.Width(row)
	if rowVisW < fullWidth {
		row += baseStyle.Render(strings.Repeat(" ", fullWidth-rowVisW))
	}

	return row
}

// highlightSubstring renders a string with the query substring highlighted using
// the given highlight background color. The rest uses the base style.
func highlightSubstring(s, query string, baseStyle lipgloss.Style, highlightBg lipgloss.Color) string {
	if query == "" {
		return baseStyle.Render(s)
	}

	lower := strings.ToLower(s)
	lowerQuery := strings.ToLower(query)

	highlightStyle := baseStyle.
		Background(highlightBg).
		Foreground(styles.LightLavender).
		Bold(true)

	var result strings.Builder
	pos := 0
	for {
		idx := strings.Index(lower[pos:], lowerQuery)
		if idx < 0 {
			// No more matches — render the rest
			result.WriteString(baseStyle.Render(s[pos:]))
			break
		}
		// Render text before the match
		if idx > 0 {
			result.WriteString(baseStyle.Render(s[pos : pos+idx]))
		}
		// Render the matching substring
		matchEnd := pos + idx + len(lowerQuery)
		result.WriteString(highlightStyle.Render(s[pos+idx : matchEnd]))
		pos = matchEnd
	}

	return result.String()
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
