// Package components provides reusable TUI components.
package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// CommandInputState holds the state for the command input component.
type CommandInputState struct {
	// Active indicates if command mode is active
	Active bool
	// Input is the current command input buffer
	Input string
	// CursorPos is the cursor position within the input
	CursorPos int
	// Result is the result message to display (success or error)
	Result string
	// IsError indicates if the result is an error message
	IsError bool
}

// CommandInput renders the command input component.
// When active, it shows a ':' prompt with the current input.
// When not active but there's a result, it shows the result message.
// Otherwise, it shows a help hint.
func CommandInput(state CommandInputState, width int) string {
	if state.Active {
		// Command mode active - show prompt and input
		promptStyle := lipgloss.NewStyle().
			Foreground(styles.Cyan).
			Bold(true)

		inputStyle := lipgloss.NewStyle().
			Foreground(styles.LightLavender)

		// Build the input line with cursor
		input := state.Input
		cursor := "_"

		// Insert cursor at position
		var displayInput string
		if state.CursorPos >= len(input) {
			displayInput = input + cursor
		} else {
			displayInput = input[:state.CursorPos] + cursor + input[state.CursorPos:]
		}

		content := promptStyle.Render(":") + inputStyle.Render(displayInput)

		// Apply background to full width
		lineStyle := lipgloss.NewStyle().
			Background(styles.DarkPurple).
			Width(width)

		return lineStyle.Render(content)
	}

	if state.Result != "" {
		// Show result message
		var resultStyle lipgloss.Style
		if state.IsError {
			resultStyle = lipgloss.NewStyle().
				Foreground(styles.Pink).
				Bold(true)
		} else {
			resultStyle = lipgloss.NewStyle().
				Foreground(styles.Cyan).
				Bold(true)
		}

		lineStyle := lipgloss.NewStyle().
			Background(styles.DarkPurple).
			Width(width)

		return lineStyle.Render(" " + resultStyle.Render(state.Result))
	}

	// Default: empty bar
	lineStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Width(width)

	return lineStyle.Render(" ")
}

// InsertChar inserts a character at the current cursor position.
func (s *CommandInputState) InsertChar(c rune) {
	if s.CursorPos >= len(s.Input) {
		s.Input += string(c)
	} else {
		s.Input = s.Input[:s.CursorPos] + string(c) + s.Input[s.CursorPos:]
	}
	s.CursorPos++
}

// Backspace deletes the character before the cursor.
func (s *CommandInputState) Backspace() {
	if s.CursorPos > 0 && len(s.Input) > 0 {
		if s.CursorPos >= len(s.Input) {
			s.Input = s.Input[:len(s.Input)-1]
		} else {
			s.Input = s.Input[:s.CursorPos-1] + s.Input[s.CursorPos:]
		}
		s.CursorPos--
	}
}

// Delete deletes the character at the cursor.
func (s *CommandInputState) Delete() {
	if s.CursorPos < len(s.Input) {
		s.Input = s.Input[:s.CursorPos] + s.Input[s.CursorPos+1:]
	}
}

// MoveCursorLeft moves the cursor left.
func (s *CommandInputState) MoveCursorLeft() {
	if s.CursorPos > 0 {
		s.CursorPos--
	}
}

// MoveCursorRight moves the cursor right.
func (s *CommandInputState) MoveCursorRight() {
	if s.CursorPos < len(s.Input) {
		s.CursorPos++
	}
}

// Clear clears the input buffer and deactivates command mode.
func (s *CommandInputState) Clear() {
	s.Input = ""
	s.CursorPos = 0
	s.Active = false
}

// GetCommand returns the current command and clears the input.
func (s *CommandInputState) GetCommand() string {
	cmd := s.Input
	s.Clear()
	return cmd
}

// SetResult sets the result message.
func (s *CommandInputState) SetResult(msg string, isError bool) {
	s.Result = msg
	s.IsError = isError
}

// ClearResult clears the result message.
func (s *CommandInputState) ClearResult() {
	s.Result = ""
	s.IsError = false
}
