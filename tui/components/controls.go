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
				{Name: "Back", Shortcut: "H/\u2190"},
				{Name: "Fwd", Shortcut: "L/\u2192"},
				{Name: "Play", Shortcut: "Space"},
				{Name: "Frame-", Shortcut: "C-h"},
				{Name: "Frame+", Shortcut: "C-l"},
				{Name: "Speed-", Shortcut: "[/{"},
				{Name: "Speed+", Shortcut: "]/}"},
				{Name: "Speed1x", Shortcut: "\\"},
			},
		},
		// Navigation controls
		{
			Name: "Navigation",
			Controls: []Control{
				{Name: "Prev", Shortcut: "J/\u2191"},
				{Name: "Next", Shortcut: "K/\u2193"},
				{Name: "Mute", Shortcut: "M"},
			},
		},
		// Step/overlay controls
		{
			Name: "Step / Overlay",
			Controls: []Control{
				{Name: "Step-", Shortcut: ",/<"},
				{Name: "Step+", Shortcut: "./>"},
				{Name: "Overlay", Shortcut: "O"},
			},
		},
		// View controls
		{
			Name: "Views",
			Controls: []Control{
				{Name: "Stats", Shortcut: "S"},
				{Name: "Help", Shortcut: "?"},
				{Name: "Quit", Shortcut: "Ctrl+C"},
			},
		},
	}
}

// RenderInfoBox renders a generic bordered box with a tab-style header and content lines.
// Content lines are rendered as-is (caller handles styling). The box uses the same
// box-drawing characters as RenderMiniPlayer.
func RenderInfoBox(title string, contentLines []string, width int) string {
	if width < 4 {
		return ""
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	borderColor := styles.Purple

	// Tab header: ╭─ Title ─────╮
	headerText := headerStyle.Render(" " + title + " ")
	headerTextWidth := lipgloss.Width(headerText)

	topBorderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topLeft := topBorderStyle.Render("╭─")
	topRight := topBorderStyle.Render("╮")
	topLeftWidth := 2
	topRightWidth := 1
	fillWidth := innerWidth - topLeftWidth - headerTextWidth - topRightWidth + 2
	if fillWidth < 0 {
		fillWidth = 0
	}
	topFill := strings.Repeat("─", fillWidth)
	topLine := topLeft + headerText + topBorderStyle.Render(topFill) + topRight

	sideStyle := lipgloss.NewStyle().Foreground(borderColor)
	var renderedLines []string
	renderedLines = append(renderedLines, topLine)

	for _, line := range contentLines {
		lineWidth := lipgloss.Width(line)
		pad := innerWidth - lineWidth
		if pad < 0 {
			pad = 0
		}
		renderedLines = append(renderedLines, sideStyle.Render("│")+line+strings.Repeat(" ", pad)+sideStyle.Render("│"))
	}

	// Bottom border: ╰──────────────╯
	bottomLine := topBorderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	renderedLines = append(renderedLines, bottomLine)

	return strings.Join(renderedLines, "\n")
}

// ControlsDisplay renders the controls display component.
// It shows all available controls grouped by function with [Key] Label format.
func ControlsDisplay(width int) string {
	groups := GetControlGroups()

	shortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)

	nameStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	// Build control strings for each group
	var groupStrings []string
	for _, group := range groups {
		var controlStrs []string
		for _, ctrl := range group.Controls {
			// Format: [Key] Label
			ctrlStr := shortcutStyle.Render("["+ctrl.Shortcut+"]") + " " + nameStyle.Render(ctrl.Name)
			controlStrs = append(controlStrs, ctrlStr)
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
