// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Control represents a single control with its display info.
type Control struct {
	Name     string
	Shortcut string
}

// ControlGroup represents a group of related controls with sub-group support.
// SubGroups allows the renderer to place horizontal dividers between sub-groups.
type ControlGroup struct {
	Name      string
	SubGroups [][]Control
}

// GetControlGroups returns the control groups for display.
func GetControlGroups() []ControlGroup {
	return []ControlGroup{
		// Playback controls — three sub-groups separated by dividers
		{
			Name: "Playback",
			SubGroups: [][]Control{
				{
					{Name: "Play", Shortcut: "Space"},
					{Name: "Back", Shortcut: "H / \u2190"},
					{Name: "Fwd", Shortcut: "L / \u2192"},
				},
				{
					{Name: "Step -", Shortcut: ", / <"},
					{Name: "Step +", Shortcut: ". / >"},
				},
				{
					{Name: "Frame -", Shortcut: "Ctrl+h"},
					{Name: "Frame +", Shortcut: "Ctrl+l"},
				},
			},
		},
		// Navigation controls — single sub-group, no dividers
		{
			Name: "Navigation",
			SubGroups: [][]Control{
				{
					{Name: "Prev", Shortcut: "J / \u2191"},
					{Name: "Next", Shortcut: "K / \u2193"},
					{Name: "Mute", Shortcut: "M"},
					{Name: "Overlay", Shortcut: "O"},
				},
			},
		},
		// View controls — single sub-group, no dividers
		{
			Name: "Views",
			SubGroups: [][]Control{
				{
					{Name: "Stats", Shortcut: "S"},
					{Name: "Help", Shortcut: "?"},
					{Name: "Quit", Shortcut: "Ctrl+C"},
				},
			},
		},
	}
}

// ControlsDisplay renders the controls display component as a horizontal bar.
// It shows all available controls grouped by function with Name [Shortcut] format.
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
		for _, subGroup := range group.SubGroups {
			for _, ctrl := range subGroup {
				ctrlStr := ctrl.Name + " " + shortcutStyle.Render("["+ctrl.Shortcut+"]")
				controlStrs = append(controlStrs, controlStyle.Render(ctrlStr))
			}
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
