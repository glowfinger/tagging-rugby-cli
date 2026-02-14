// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
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
					{Name: "Back", Shortcut: "H"},
					{Name: "Fwd", Shortcut: "L"},
				},
				{
					{Name: "Step -", Shortcut: ", / <"},
					{Name: "Step +", Shortcut: ". / >"},
				},
				{
					{Name: "Frame -", Shortcut: "Ctrl+h"},
					{Name: "Frame +", Shortcut: "Ctrl+l"},
				},
				{
					{Name: "Speed -", Shortcut: "[ / {"},
					{Name: "Speed +", Shortcut: "] / }"},
					{Name: "Speed 1x", Shortcut: "\\"},
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
					{Name: "Sort", Shortcut: "X"},
					{Name: "Help", Shortcut: "?"},
					{Name: "Quit", Shortcut: "Ctrl+C"},
				},
			},
		},
	}
}

// RenderInfoBox renders a generic bordered box with a tab-style header and content lines.
// Content lines are rendered as-is (caller handles styling). The box uses the same
// box-drawing characters as RenderVideoBox.
func RenderInfoBox(title string, contentLines []string, width int, focused bool) string {
	if width < 4 {
		return ""
	}

	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	borderColor := styles.Purple
	if focused {
		borderColor = styles.Pink
	}

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

// controlGroupLines converts a ControlGroup into content lines for RenderInfoBox.
// Each control renders as a styled line with left-aligned name and right-aligned shortcut bracket.
// Sub-groups are separated by a blank line.
func ControlGroupLines(group ControlGroup, innerWidth int) []string {
	nameStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	shortcutStyle := lipgloss.NewStyle().Foreground(styles.Cyan).Bold(true)

	// Find max control name width for alignment
	maxNameW := 0
	for _, sg := range group.SubGroups {
		for _, c := range sg {
			if len(c.Name) > maxNameW {
				maxNameW = len(c.Name)
			}
		}
	}

	var lines []string
	for si, subGroup := range group.SubGroups {
		for _, c := range subGroup {
			namePart := nameStyle.Render(fmt.Sprintf("%-*s", maxNameW, c.Name))
			shortcutPart := shortcutStyle.Render("[ " + c.Shortcut + " ]")

			contentStr := " " + namePart + "  " + shortcutPart
			contentVisW := lipgloss.Width(contentStr)
			padRight := innerWidth - contentVisW
			if padRight < 0 {
				padRight = 0
			}
			line := contentStr + strings.Repeat(" ", padRight)
			if lipgloss.Width(line) > innerWidth {
				line = ansi.Truncate(line, innerWidth, "")
			}
			lines = append(lines, line)
		}

		// Blank line between sub-groups (not after the last)
		if si < len(group.SubGroups)-1 {
			lines = append(lines, "")
		}
	}

	return lines
}

// Deprecated: Use RenderInfoBox with ControlGroupLines instead.
//
// RenderControlBox renders a control group inside a bordered box with tab header
// and horizontal dividers between sub-groups.
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

// RenderVideoBox renders the video status card at full column width.
// Uses RenderInfoBox for consistent styling across all containers.
func RenderVideoBox(state StatusBarState, width int, showWarning bool, focused bool) string {
	textStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	warningStyle := lipgloss.NewStyle().Foreground(styles.Lavender).Italic(true)

	// Build first line: left-align play state, right-align step size + mute
	playState := "▶ Playing"
	if state.Paused {
		playState = "⏸ Paused"
	}

	stepStr := formatStepSize(state.StepSize)
	leftPart := " " + playState
	rightPart := "Step: " + stepStr
	if state.Muted {
		rightPart += "  🔇"
	}

	innerW := width - 4 // 2 border chars + 2 padding spaces from InfoBox
	if innerW < 1 {
		innerW = 1
	}
	padding := innerW - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if padding < 1 {
		padding = 1
	}
	statusLine := leftPart + strings.Repeat(" ", padding) + rightPart

	timeLine := fmt.Sprintf(" Time: %s / %s",
		timeutil.FormatTime(state.TimePos),
		timeutil.FormatTime(state.Duration))

	overlayLine := " Overlay: off"
	if state.OverlayEnabled {
		overlayLine = " Overlay: on"
	}

	videoLine := " Video: Closed"
	if state.VideoOpen {
		videoLine = " Video: Open"
	}

	contentLines := []string{
		textStyle.Render(statusLine),
		textStyle.Render(timeLine),
		textStyle.Render(overlayLine),
		textStyle.Render(videoLine),
	}

	card := RenderInfoBox("Video", contentLines, width, focused)

	if !showWarning {
		return card
	}

	// Warning line below card
	warning := warningStyle.Render("Mini player mode - resize for full view")
	warnW := lipgloss.Width(warning)
	warnPad := (width - warnW) / 2
	if warnPad < 0 {
		warnPad = 0
	}

	return card + "\n" + strings.Repeat(" ", warnPad) + warning
}

// ControlsDisplay renders the controls display component as a horizontal bar.
// It shows all available controls grouped by function with Name [Shortcut] format.
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
		for _, subGroup := range group.SubGroups {
			for _, ctrl := range subGroup {
				ctrlStr := nameStyle.Render(ctrl.Name) + " " + shortcutStyle.Render("["+ctrl.Shortcut+"]")
				controlStrs = append(controlStrs, ctrlStr)
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
