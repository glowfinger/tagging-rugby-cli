// Package components provides reusable TUI components.
//
// Deprecated: TackleInput is replaced by tui/forms/tackleform.go using huh forms.
// This file is kept for reference and will be removed in a future cleanup.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// TackleInputField represents which field is currently active in the tackle input.
type TackleInputField int

const (
	// TackleInputFieldPlayer is the player input field.
	TackleInputFieldPlayer TackleInputField = iota
	// TackleInputFieldTeam is the team input field.
	TackleInputFieldTeam
	// TackleInputFieldAttempt is the attempt number input field.
	TackleInputFieldAttempt
	// TackleInputFieldOutcome is the outcome selection field.
	TackleInputFieldOutcome
	// TackleInputFieldFollowed is the followed input field (optional).
	TackleInputFieldFollowed
	// TackleInputFieldNotes is the notes input field (optional).
	TackleInputFieldNotes
	// TackleInputFieldZone is the zone input field (optional).
	TackleInputFieldZone
)

// TackleOutcome represents the valid outcomes for a tackle.
var TackleOutcomes = []string{"completed", "missed", "possible", "other"}

// TackleInputState holds the state for the quick tackle input component.
type TackleInputState struct {
	// Active indicates if the tackle input prompt is visible
	Active bool
	// CurrentField is the currently focused field
	CurrentField TackleInputField
	// Player is the player name (required)
	Player string
	// Team is the team name (required)
	Team string
	// Attempt is the attempt number (required)
	Attempt string
	// OutcomeIndex is the index of the selected outcome in TackleOutcomes
	OutcomeIndex int
	// Followed is who followed up on the tackle (optional)
	Followed string
	// Notes is additional notes about the tackle (optional)
	Notes string
	// Zone is the field zone (optional)
	Zone string
	// Star indicates if this tackle is starred
	Star bool
	// Timestamp is the timestamp when the prompt was opened
	Timestamp float64
}

// TackleInput renders the quick tackle input component.
// Shows timestamp, required fields (player, team, attempt, outcome), and optional fields.
func TackleInput(state TackleInputState, width int, timestamp float64) string {
	// Format timestamp as MM:SS
	totalSeconds := int(timestamp)
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	timestampStr := fmt.Sprintf("%d:%02d", mins, secs)

	// Header with timestamp and star indicator
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)

	starStr := ""
	if state.Star {
		starStr = " " + lipgloss.NewStyle().Foreground(styles.Cyan).Render("★")
	}
	header := headerStyle.Render(fmt.Sprintf("Add Tackle @ %s", timestampStr)) + starStr

	// Field styles
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Width(10)

	requiredLabelStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Width(10)

	activeInputStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender).
		Background(styles.Purple)

	inactiveInputStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	cursor := "_"

	// Build fields
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	// Player field (required)
	playerLabel := requiredLabelStyle.Render("Player*:")
	playerValue := state.Player
	if state.CurrentField == TackleInputFieldPlayer {
		playerValue = activeInputStyle.Render(playerValue + cursor)
	} else {
		if playerValue == "" {
			playerValue = inactiveInputStyle.Render("(required)")
		} else {
			playerValue = inactiveInputStyle.Render(playerValue)
		}
	}
	lines = append(lines, playerLabel+playerValue)

	// Team field (required)
	teamLabel := requiredLabelStyle.Render("Team*:")
	teamValue := state.Team
	if state.CurrentField == TackleInputFieldTeam {
		teamValue = activeInputStyle.Render(teamValue + cursor)
	} else {
		if teamValue == "" {
			teamValue = inactiveInputStyle.Render("(required)")
		} else {
			teamValue = inactiveInputStyle.Render(teamValue)
		}
	}
	lines = append(lines, teamLabel+teamValue)

	// Attempt field (required)
	attemptLabel := requiredLabelStyle.Render("Attempt*:")
	attemptValue := state.Attempt
	if state.CurrentField == TackleInputFieldAttempt {
		attemptValue = activeInputStyle.Render(attemptValue + cursor)
	} else {
		if attemptValue == "" {
			attemptValue = inactiveInputStyle.Render("(required)")
		} else {
			attemptValue = inactiveInputStyle.Render(attemptValue)
		}
	}
	lines = append(lines, attemptLabel+attemptValue)

	// Outcome field (required) - selection style
	outcomeLabel := requiredLabelStyle.Render("Outcome*:")
	var outcomeParts []string
	for i, outcome := range TackleOutcomes {
		if state.CurrentField == TackleInputFieldOutcome {
			// In active field, highlight the selected outcome
			if i == state.OutcomeIndex {
				outcomeParts = append(outcomeParts, activeInputStyle.Render("["+outcome+"]"))
			} else {
				outcomeParts = append(outcomeParts, unselectedStyle.Render(" "+outcome+" "))
			}
		} else {
			// Not in this field, show only selected outcome
			if i == state.OutcomeIndex {
				outcomeParts = append(outcomeParts, selectedStyle.Render(outcome))
			}
		}
	}
	outcomeValue := ""
	if state.CurrentField == TackleInputFieldOutcome {
		for _, part := range outcomeParts {
			outcomeValue += part + " "
		}
	} else {
		outcomeValue = inactiveInputStyle.Render(TackleOutcomes[state.OutcomeIndex])
	}
	lines = append(lines, outcomeLabel+outcomeValue)

	// Separator for optional fields
	lines = append(lines, "")
	optionalHeader := lipgloss.NewStyle().Foreground(styles.Lavender).Italic(true).Render("Optional:")
	lines = append(lines, optionalHeader)

	// Followed field (optional)
	followedLabel := labelStyle.Render("Followed:")
	followedValue := state.Followed
	if state.CurrentField == TackleInputFieldFollowed {
		followedValue = activeInputStyle.Render(followedValue + cursor)
	} else {
		if followedValue == "" {
			followedValue = inactiveInputStyle.Render("-")
		} else {
			followedValue = inactiveInputStyle.Render(followedValue)
		}
	}
	lines = append(lines, followedLabel+followedValue)

	// Notes field (optional)
	notesLabel := labelStyle.Render("Notes:")
	notesValue := state.Notes
	if state.CurrentField == TackleInputFieldNotes {
		notesValue = activeInputStyle.Render(notesValue + cursor)
	} else {
		if notesValue == "" {
			notesValue = inactiveInputStyle.Render("-")
		} else {
			notesValue = inactiveInputStyle.Render(notesValue)
		}
	}
	lines = append(lines, notesLabel+notesValue)

	// Zone field (optional)
	zoneLabel := labelStyle.Render("Zone:")
	zoneValue := state.Zone
	if state.CurrentField == TackleInputFieldZone {
		zoneValue = activeInputStyle.Render(zoneValue + cursor)
	} else {
		if zoneValue == "" {
			zoneValue = inactiveInputStyle.Render("-")
		} else {
			zoneValue = inactiveInputStyle.Render(zoneValue)
		}
	}
	lines = append(lines, zoneLabel+zoneValue)

	// Footer with instructions
	lines = append(lines, "")
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true)
	lines = append(lines, footerStyle.Render("Tab/Shift+Tab: fields | Enter: save | */S: star | Esc: cancel"))

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

// Clear resets the tackle input state.
func (s *TackleInputState) Clear() {
	s.Active = false
	s.CurrentField = TackleInputFieldPlayer
	s.Player = ""
	s.Team = ""
	s.Attempt = ""
	s.OutcomeIndex = 0
	s.Followed = ""
	s.Notes = ""
	s.Zone = ""
	s.Star = false
	s.Timestamp = 0
}

// NextField moves to the next field (cycles through all fields).
func (s *TackleInputState) NextField() {
	s.CurrentField = (s.CurrentField + 1) % 7
}

// PrevField moves to the previous field (cycles back to last field).
func (s *TackleInputState) PrevField() {
	if s.CurrentField == 0 {
		s.CurrentField = TackleInputFieldZone
	} else {
		s.CurrentField--
	}
}

// NextOutcome cycles to the next outcome in the list.
func (s *TackleInputState) NextOutcome() {
	s.OutcomeIndex = (s.OutcomeIndex + 1) % len(TackleOutcomes)
}

// PrevOutcome cycles to the previous outcome in the list.
func (s *TackleInputState) PrevOutcome() {
	if s.OutcomeIndex == 0 {
		s.OutcomeIndex = len(TackleOutcomes) - 1
	} else {
		s.OutcomeIndex--
	}
}

// ToggleStar toggles the star flag.
func (s *TackleInputState) ToggleStar() {
	s.Star = !s.Star
}

// InsertChar inserts a character into the current field.
func (s *TackleInputState) InsertChar(c rune) {
	switch s.CurrentField {
	case TackleInputFieldPlayer:
		s.Player += string(c)
	case TackleInputFieldTeam:
		s.Team += string(c)
	case TackleInputFieldAttempt:
		// Only allow digits for attempt
		if c >= '0' && c <= '9' {
			s.Attempt += string(c)
		}
	case TackleInputFieldFollowed:
		s.Followed += string(c)
	case TackleInputFieldNotes:
		s.Notes += string(c)
	case TackleInputFieldZone:
		s.Zone += string(c)
		// For outcome field, left/right arrows are handled separately
	}
}

// Backspace deletes the last character from the current field.
func (s *TackleInputState) Backspace() {
	switch s.CurrentField {
	case TackleInputFieldPlayer:
		if len(s.Player) > 0 {
			s.Player = s.Player[:len(s.Player)-1]
		}
	case TackleInputFieldTeam:
		if len(s.Team) > 0 {
			s.Team = s.Team[:len(s.Team)-1]
		}
	case TackleInputFieldAttempt:
		if len(s.Attempt) > 0 {
			s.Attempt = s.Attempt[:len(s.Attempt)-1]
		}
	case TackleInputFieldFollowed:
		if len(s.Followed) > 0 {
			s.Followed = s.Followed[:len(s.Followed)-1]
		}
	case TackleInputFieldNotes:
		if len(s.Notes) > 0 {
			s.Notes = s.Notes[:len(s.Notes)-1]
		}
	case TackleInputFieldZone:
		if len(s.Zone) > 0 {
			s.Zone = s.Zone[:len(s.Zone)-1]
		}
	}
}

// GetTackle returns the tackle data from the input state.
// Returns player, team, attempt, outcome, followed, notes, zone, star, timestamp.
func (s *TackleInputState) GetTackle() (player, team, attempt, outcome, followed, notes, zone string, star bool, timestamp float64) {
	return s.Player, s.Team, s.Attempt, TackleOutcomes[s.OutcomeIndex], s.Followed, s.Notes, s.Zone, s.Star, s.Timestamp
}

// IsValid returns true if all required fields are filled.
func (s *TackleInputState) IsValid() bool {
	return s.Player != "" && s.Team != "" && s.Attempt != ""
}

// ValidationError returns an error message if any required field is missing.
func (s *TackleInputState) ValidationError() string {
	if s.Player == "" {
		return "Player is required"
	}
	if s.Team == "" {
		return "Team is required"
	}
	if s.Attempt == "" {
		return "Attempt number is required"
	}
	return ""
}
