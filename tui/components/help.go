// Package components provides reusable TUI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// HelpOverlay renders the help overlay showing all keybindings.
// The overlay is styled with the palette colors and grouped by function.
func HelpOverlay(width, height int) string {
	// Define keybinding groups
	groups := []struct {
		title    string
		bindings []struct {
			key  string
			desc string
		}
	}{
		{
			title: "Playback",
			bindings: []struct {
				key  string
				desc string
			}{
				{"Space", "Toggle play/pause"},
				{"M", "Toggle mute"},
				{"H / Left", "Step backward (by step size)"},
				{"L / Right", "Step forward (by step size)"},
				{"Ctrl+H", "Frame step backward"},
				{"Ctrl+L", "Frame step forward"},
				{", / <", "Decrease step size"},
				{". / >", "Increase step size"},
			},
		},
		{
			title: "Navigation",
			bindings: []struct {
				key  string
				desc string
			}{
				{"J / Up", "Select previous item"},
				{"K / Down", "Select next item"},
				{"Enter", "Jump to selected item"},
				{"E", "Edit selected tackle"},
				{"X", "Delete selected item"},
				{"Ctrl+E", "Export player clips"},
			},
		},
		{
			title: "Views",
			bindings: []struct {
				key  string
				desc string
			}{
				{"?", "Show/hide this help"},
				{"S", "Open stats view"},
				{"O", "Toggle overlay on video"},
				{"N", "Quick add note"},
				{"T", "Quick add tackle"},
				{"Backspace", "Return to main view"},
				{"/ (stats)", "Filter players by name/initials"},
				{"Esc (stats)", "Clear player filters"},
			},
		},
		{
			title: "Commands",
			bindings: []struct {
				key  string
				desc string
			}{
				{":", "Enter command mode"},
				{"Esc", "Cancel command mode"},
				{"Ctrl+C", "Quit application"},
			},
		},
		{
			title: "Shorthand Commands",
			bindings: []struct {
				key  string
				desc string
			}{
				{":nn", "Quick note (or :nn <text>)"},
				{":nt", "Quick tackle (or :nt <p> <t> <a> <o>)"},
				{":cs", "Clip start"},
				{":ce <desc>", "Clip end with description"},
			},
		},
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true).
		Padding(0, 1)

	// Group header style
	groupHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true).
		MarginTop(1)

	// Key style
	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Bold(true).
		Width(12)

	// Description style
	descStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	// Build help content
	var lines []string

	// Title
	lines = append(lines, titleStyle.Render("Keybindings"))
	lines = append(lines, "")

	// Render each group
	for _, group := range groups {
		lines = append(lines, groupHeaderStyle.Render(group.title))
		for _, binding := range group.bindings {
			line := "  " + keyStyle.Render(binding.key) + descStyle.Render(binding.desc)
			lines = append(lines, line)
		}
	}

	// Footer
	lines = append(lines, "")
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true)
	lines = append(lines, footerStyle.Render("Press any key to close"))

	content := strings.Join(lines, "\n")

	// Calculate content dimensions
	contentLines := strings.Split(content, "\n")
	contentHeight := len(contentLines)
	contentWidth := 0
	for _, line := range contentLines {
		w := lipgloss.Width(line)
		if w > contentWidth {
			contentWidth = w
		}
	}

	// Add padding
	paddedWidth := contentWidth + 4
	paddedHeight := contentHeight + 2

	// Center the overlay
	marginLeft := (width - paddedWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}
	marginTop := (height - paddedHeight) / 2
	if marginTop < 0 {
		marginTop = 0
	}

	// Overlay panel style with border
	panelStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BrightPurple).
		Padding(1, 2)

	// Apply panel style
	panel := panelStyle.Render(content)

	// Create positioning by adding margin
	positionedStyle := lipgloss.NewStyle().
		MarginLeft(marginLeft).
		MarginTop(marginTop)

	return positionedStyle.Render(panel)
}
