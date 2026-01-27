// Package components provides reusable TUI components.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// NoteInputField represents which field is currently active in the note input.
type NoteInputField int

const (
	// NoteInputFieldText is the text input field.
	NoteInputFieldText NoteInputField = iota
	// NoteInputFieldCategory is the category input field.
	NoteInputFieldCategory
	// NoteInputFieldPlayer is the player input field.
	NoteInputFieldPlayer
	// NoteInputFieldTeam is the team input field.
	NoteInputFieldTeam
)

// NoteInputState holds the state for the quick note input component.
type NoteInputState struct {
	// Active indicates if the note input prompt is visible
	Active bool
	// CurrentField is the currently focused field
	CurrentField NoteInputField
	// Text is the note text input
	Text string
	// Category is the optional category
	Category string
	// Player is the optional player name
	Player string
	// Team is the optional team name
	Team string
	// Timestamp is the timestamp when the prompt was opened
	Timestamp float64
}

// NoteInput renders the quick note input component.
// Shows timestamp, text field, and optional category/player/team fields.
func NoteInput(state NoteInputState, width int, timestamp float64) string {
	// Format timestamp as MM:SS
	totalSeconds := int(timestamp)
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	timestampStr := fmt.Sprintf("%d:%02d", mins, secs)

	// Header with timestamp
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	header := headerStyle.Render(fmt.Sprintf("Add Note @ %s", timestampStr))

	// Field styles
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Width(10)

	activeInputStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender).
		Background(styles.Purple)

	inactiveInputStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	cursor := "_"

	// Build fields
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	// Text field (required)
	textLabel := labelStyle.Render("Text:")
	textValue := state.Text
	if state.CurrentField == NoteInputFieldText {
		textValue = activeInputStyle.Render(textValue + cursor)
	} else {
		if textValue == "" {
			textValue = inactiveInputStyle.Render("(empty)")
		} else {
			textValue = inactiveInputStyle.Render(textValue)
		}
	}
	lines = append(lines, textLabel+textValue)

	// Category field (optional)
	categoryLabel := labelStyle.Render("Category:")
	categoryValue := state.Category
	if state.CurrentField == NoteInputFieldCategory {
		categoryValue = activeInputStyle.Render(categoryValue + cursor)
	} else {
		if categoryValue == "" {
			categoryValue = inactiveInputStyle.Render("(optional)")
		} else {
			categoryValue = inactiveInputStyle.Render(categoryValue)
		}
	}
	lines = append(lines, categoryLabel+categoryValue)

	// Player field (optional)
	playerLabel := labelStyle.Render("Player:")
	playerValue := state.Player
	if state.CurrentField == NoteInputFieldPlayer {
		playerValue = activeInputStyle.Render(playerValue + cursor)
	} else {
		if playerValue == "" {
			playerValue = inactiveInputStyle.Render("(optional)")
		} else {
			playerValue = inactiveInputStyle.Render(playerValue)
		}
	}
	lines = append(lines, playerLabel+playerValue)

	// Team field (optional)
	teamLabel := labelStyle.Render("Team:")
	teamValue := state.Team
	if state.CurrentField == NoteInputFieldTeam {
		teamValue = activeInputStyle.Render(teamValue + cursor)
	} else {
		if teamValue == "" {
			teamValue = inactiveInputStyle.Render("(optional)")
		} else {
			teamValue = inactiveInputStyle.Render(teamValue)
		}
	}
	lines = append(lines, teamLabel+teamValue)

	// Footer with instructions
	lines = append(lines, "")
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true)
	lines = append(lines, footerStyle.Render("Tab: next field | Enter: save | Esc: cancel"))

	// Join lines
	content := ""
	for i, line := range lines {
		if i > 0 {
			content += "\n"
		}
		content += line
	}

	// Apply background style to full width
	lineStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Width(width).
		Padding(0, 1)

	return lineStyle.Render(content)
}

// Clear resets the note input state.
func (s *NoteInputState) Clear() {
	s.Active = false
	s.CurrentField = NoteInputFieldText
	s.Text = ""
	s.Category = ""
	s.Player = ""
	s.Team = ""
	s.Timestamp = 0
}

// NextField moves to the next field (cycles back to text).
func (s *NoteInputState) NextField() {
	s.CurrentField = (s.CurrentField + 1) % 4
}

// PrevField moves to the previous field (cycles back to team).
func (s *NoteInputState) PrevField() {
	if s.CurrentField == 0 {
		s.CurrentField = NoteInputFieldTeam
	} else {
		s.CurrentField--
	}
}

// InsertChar inserts a character into the current field.
func (s *NoteInputState) InsertChar(c rune) {
	switch s.CurrentField {
	case NoteInputFieldText:
		s.Text += string(c)
	case NoteInputFieldCategory:
		s.Category += string(c)
	case NoteInputFieldPlayer:
		s.Player += string(c)
	case NoteInputFieldTeam:
		s.Team += string(c)
	}
}

// Backspace deletes the last character from the current field.
func (s *NoteInputState) Backspace() {
	switch s.CurrentField {
	case NoteInputFieldText:
		if len(s.Text) > 0 {
			s.Text = s.Text[:len(s.Text)-1]
		}
	case NoteInputFieldCategory:
		if len(s.Category) > 0 {
			s.Category = s.Category[:len(s.Category)-1]
		}
	case NoteInputFieldPlayer:
		if len(s.Player) > 0 {
			s.Player = s.Player[:len(s.Player)-1]
		}
	case NoteInputFieldTeam:
		if len(s.Team) > 0 {
			s.Team = s.Team[:len(s.Team)-1]
		}
	}
}

// GetNote returns the note data from the input state.
func (s *NoteInputState) GetNote() (text, category, player, team string, timestamp float64) {
	return s.Text, s.Category, s.Player, s.Team, s.Timestamp
}
