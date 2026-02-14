package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// SearchInputState holds the state for the search input component.
type SearchInputState struct {
	// Input is the current search/command input buffer
	Input string
	// CursorPos is the cursor position within the input
	CursorPos int
	// Mode is either "search" or "command"
	Mode string
	// Matches holds indices of matching items
	Matches []int
	// CurrentMatch is the index into Matches of the current match
	CurrentMatch int
}

// SearchInput renders the search input component inside a RenderInfoBox.
// When focused, the box border is pink; otherwise purple.
func SearchInput(state SearchInputState, width int, focused bool) string {
	promptStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	// Determine prefix based on mode
	prefix := "/"
	if state.Mode == "command" {
		prefix = ":"
	}

	// Build input display with cursor
	input := state.Input
	var displayInput string
	if focused {
		cursor := "_"
		if state.CursorPos >= len(input) {
			displayInput = input + cursor
		} else {
			displayInput = input[:state.CursorPos] + cursor + input[state.CursorPos:]
		}
	} else {
		displayInput = input
	}

	content := " " + promptStyle.Render(prefix) + inputStyle.Render(displayInput)

	// Match indicator right-aligned
	if len(state.Matches) > 0 {
		indicator := fmt.Sprintf("[%d/%d]", state.CurrentMatch+1, len(state.Matches))
		indicatorStyled := lipgloss.NewStyle().Foreground(styles.Lavender).Render(indicator)

		innerW := width - 4 // InfoBox inner width
		contentW := lipgloss.Width(content)
		indicatorW := lipgloss.Width(indicatorStyled)
		pad := innerW - contentW - indicatorW
		if pad < 1 {
			pad = 1
		}
		content = content + strings.Repeat(" ", pad) + indicatorStyled
	}

	return RenderInfoBox("Search", []string{content}, width, focused)
}

// InsertChar inserts a character at the current cursor position.
func (s *SearchInputState) InsertChar(c rune) {
	if s.CursorPos >= len(s.Input) {
		s.Input += string(c)
	} else {
		s.Input = s.Input[:s.CursorPos] + string(c) + s.Input[s.CursorPos:]
	}
	s.CursorPos++
}

// Backspace deletes the character before the cursor.
func (s *SearchInputState) Backspace() {
	if s.CursorPos > 0 && len(s.Input) > 0 {
		if s.CursorPos >= len(s.Input) {
			s.Input = s.Input[:len(s.Input)-1]
		} else {
			s.Input = s.Input[:s.CursorPos-1] + s.Input[s.CursorPos:]
		}
		s.CursorPos--
	}
}

// MoveCursorLeft moves the cursor left.
func (s *SearchInputState) MoveCursorLeft() {
	if s.CursorPos > 0 {
		s.CursorPos--
	}
}

// MoveCursorRight moves the cursor right.
func (s *SearchInputState) MoveCursorRight() {
	if s.CursorPos < len(s.Input) {
		s.CursorPos++
	}
}

// Clear resets the search input to empty state.
func (s *SearchInputState) Clear() {
	s.Input = ""
	s.CursorPos = 0
	s.Mode = "search"
	s.Matches = nil
	s.CurrentMatch = 0
}
