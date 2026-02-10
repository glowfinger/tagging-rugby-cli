// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

// RenderControlBox renders a control group inside a bordered box with tab header
// and horizontal dividers between sub-groups.
//
// Layout matches wireframe/playback.txt:
//
//	 ┌──────────┐
//	┌┤ Playback ├┐
//	│└──────────┘└────────────┐
//	│ Play    [ Space ]       │
//	├─────────────────────────┤
//	│ Step -  [ , / < ]       │
//	└─────────────────────────┘
func RenderControlBox(group ControlGroup, width int) string {
	if width < 6 {
		return ""
	}

	borderStyle := lipgloss.NewStyle().Foreground(styles.Purple)
	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	shortcutStyle := lipgloss.NewStyle().Foreground(styles.Cyan).Bold(true)

	// Box-drawing characters
	const (
		hBar = "─"
		vBar = "│"
		tl   = "┌"
		tr   = "┐"
		bl   = "└"
		br   = "┘"
		teeL = "├"
		teeR = "┤"
	)

	// Inner width is total width minus 2 border chars
	innerW := width - 2

	// Build tab header (3 lines)
	// Line 1:  ┌──<name>──┐
	// Line 2: ┌┤ <name> ├┐
	// Line 3: │└──<name>──┘└───...───┐
	tabLabel := " " + group.Name + " "
	tabW := lipgloss.Width(tabLabel) + 2 // +2 for ┤ and ├ (but they're border chars on tab)

	// Line 1: space + ┌ + ─ repeated for tab inner width + ┐
	tabInnerW := lipgloss.Width(tabLabel)
	line1 := " " + borderStyle.Render(tl+strings.Repeat(hBar, tabInnerW)+tr)

	// Line 2: ┌┤ Name ├┐  (but the outer ┐ is at the right edge if tab is short)
	// Actually looking at wireframe more carefully:
	// Line 2: ┌┤ Playback ├┐
	// This is just the tab bracket line, no extension to the right
	_ = tabW
	line2 := borderStyle.Render(tl+teeR) + headerStyle.Render(tabLabel) + borderStyle.Render(teeL+tr)

	// Line 3: │└──────────┘└────────────┐
	// Left border │, then tab bottom └─...─┘, then extension └─...─┐
	tabBottomW := tabInnerW // width of ─ inside └...┘
	remainW := innerW - tabBottomW - 3 // -3 for └, ┘, └ between tab bottom and right extension
	if remainW < 0 {
		remainW = 0
	}
	line3 := borderStyle.Render(vBar+bl+strings.Repeat(hBar, tabBottomW)+br+bl+strings.Repeat(hBar, remainW)+tr)

	var lines []string
	lines = append(lines, line1, line2, line3)

	// Find max control name width for alignment
	maxNameW := 0
	for _, sg := range group.SubGroups {
		for _, c := range sg {
			if len(c.Name) > maxNameW {
				maxNameW = len(c.Name)
			}
		}
	}

	// Render control rows
	for si, subGroup := range group.SubGroups {
		for _, c := range subGroup {
			// Format: │ Name    [ Shortcut ] │
			// Left-align name, right-align shortcut bracket
			namePart := nameStyle.Render(fmt.Sprintf("%-*s", maxNameW, c.Name))
			shortcutPart := shortcutStyle.Render("[ " + c.Shortcut + " ]")

			// Calculate padding between name and shortcut
			contentStr := namePart + "  " + shortcutPart
			contentVisW := lipgloss.Width(contentStr)
			padRight := innerW - 2 - contentVisW // -2 for leading and trailing space
			if padRight < 0 {
				padRight = 0
			}

			row := borderStyle.Render(vBar) + " " + contentStr + strings.Repeat(" ", padRight) + " " + borderStyle.Render(vBar)
			// Truncate to width if needed
			if lipgloss.Width(row) > width {
				row = ansi.Truncate(row, width, "")
			}
			lines = append(lines, row)
		}

		// Horizontal divider between sub-groups (not after the last)
		if si < len(group.SubGroups)-1 {
			divider := borderStyle.Render(teeL + strings.Repeat(hBar, innerW) + teeR)
			lines = append(lines, divider)
		}
	}

	// Bottom border
	bottom := borderStyle.Render(bl + strings.Repeat(hBar, innerW) + br)
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
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
