// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Control represents a single control with its display info.
type Control struct {
	Emoji    string
	Name     string
	Shortcut string
}

// ControlGroup represents a group of related controls.
type ControlGroup struct {
	Name     string
	Controls []Control
}

// GetControlGroups returns the control groups for display.
func GetControlGroups() []ControlGroup {
	return []ControlGroup{
		// Playback controls
		{
			Name: "Playback",
			Controls: []Control{
				{Emoji: "\u23ea", Name: "Back", Shortcut: "H/\u2190"},
				{Emoji: "\u23e9", Name: "Fwd", Shortcut: "L/\u2192"},
				{Emoji: "\u23ef\ufe0f", Name: "Play", Shortcut: "Space"},
				{Emoji: "\U0001F3AC", Name: "Frame-", Shortcut: "C-h"},
				{Emoji: "\U0001F3AC", Name: "Frame+", Shortcut: "C-l"},
				{Emoji: "\u23ea", Name: "Speed-", Shortcut: "[/{"},
				{Emoji: "\u23e9", Name: "Speed+", Shortcut: "]/}"},
				{Emoji: "\U0001F504", Name: "Speed1x", Shortcut: "\\"},
			},
		},
		// Navigation controls
		{
			Name: "Navigation",
			Controls: []Control{
				{Emoji: "\u23ee", Name: "Prev", Shortcut: "J/\u2191"},
				{Emoji: "\u23ed", Name: "Next", Shortcut: "K/\u2193"},
				{Emoji: "\U0001F507", Name: "Mute", Shortcut: "M"},
			},
		},
		// Step/overlay controls
		{
			Name: "Step / Overlay",
			Controls: []Control{
				{Emoji: "\u2796", Name: "Step-", Shortcut: ",/<"},
				{Emoji: "\u2795", Name: "Step+", Shortcut: "./>"},
				{Emoji: "\U0001F4DD", Name: "Overlay", Shortcut: "O"},
			},
		},
		// View controls
		{
			Name: "Views",
			Controls: []Control{
				{Emoji: "\U0001F4CA", Name: "Stats", Shortcut: "S"},
				{Emoji: "\u2753", Name: "Help", Shortcut: "?"},
				{Emoji: "\U0001F6AA", Name: "Quit", Shortcut: "Ctrl+C"},
			},
		},
	}
}

// ControlsDisplay renders the controls display component.
// It shows all available controls grouped by function with emoji, name, and shortcut key.
func ControlsDisplay(width int) string {
	groups := GetControlGroups()

	// Style for control items
	controlStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	shortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	// Build control strings for each group
	var groupStrings []string
	for _, group := range groups {
		var controlStrs []string
		for _, ctrl := range group.Controls {
			// Format: emoji Name [shortcut]
			ctrlStr := ctrl.Emoji + " " + ctrl.Name + " " + shortcutStyle.Render("["+ctrl.Shortcut+"]")
			controlStrs = append(controlStrs, controlStyle.Render(ctrlStr))
		}
		groupStrings = append(groupStrings, strings.Join(controlStrs, "  "))
	}

	// Join all groups with separator
	allControls := strings.Join(groupStrings, "   ")

	// Center the controls
	controlsWidth := lipgloss.Width(allControls)
	padding := (width - controlsWidth) / 2
	if padding < 0 {
		padding = 0
	}

	paddingStr := strings.Repeat(" ", padding)

	// Container style
	containerStyle := lipgloss.NewStyle().
		Background(styles.DeepPurple).
		Width(width)

	return containerStyle.Render(paddingStr + allControls)
}
