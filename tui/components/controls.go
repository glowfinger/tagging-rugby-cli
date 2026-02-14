// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

// RenderMiniPlayer renders a compact playback card for narrow terminals.
// Uses RenderInfoBox for consistent styling across all containers.
func RenderMiniPlayer(state StatusBarState, termWidth int, showWarning bool) string {
	textStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	warningStyle := lipgloss.NewStyle().Foreground(styles.Lavender).Italic(true)

	// Build content lines
	playState := "▶ Playing"
	if state.Paused {
		playState = "⏸ Paused"
	}

	stepStr := formatStepSize(state.StepSize)
	statusLine := playState + "      Step: " + stepStr
	if state.Muted {
		statusLine += "  🔇"
	}

	timeLine := fmt.Sprintf("Time: %s / %s",
		timeutil.FormatTime(state.TimePos),
		timeutil.FormatTime(state.Duration))

	overlayLine := "Overlay: off"
	if state.OverlayEnabled {
		overlayLine = "Overlay: on"
	}

	videoLine := "Video: Closed"
	if state.VideoOpen {
		videoLine = "Video: Open"
	}

	contentLines := []string{
		textStyle.Render(statusLine),
		textStyle.Render(timeLine),
		textStyle.Render(overlayLine),
		textStyle.Render(videoLine),
	}

	// Card width: fit the widest content line + 4 (2 border chars + 2 padding spaces)
	contentW := lipgloss.Width(statusLine)
	if lipgloss.Width(timeLine) > contentW {
		contentW = lipgloss.Width(timeLine)
	}
	if lipgloss.Width(overlayLine) > contentW {
		contentW = lipgloss.Width(overlayLine)
	}
	if lipgloss.Width(videoLine) > contentW {
		contentW = lipgloss.Width(videoLine)
	}
	cardWidth := contentW + 4 // │ + space + content + space + │

	// Ensure card is at least wide enough for the tab header
	minCardW := lipgloss.Width(" Playback ") + 5 // tab overhead: ╭─ + ─╮
	if cardWidth < minCardW {
		cardWidth = minCardW
	}

	card := RenderInfoBox("Playback", contentLines, cardWidth)

	// Center the card horizontally if terminal is wider than card
	if termWidth > cardWidth {
		padding := (termWidth - cardWidth) / 2
		padStr := strings.Repeat(" ", padding)
		var centeredLines []string
		for _, l := range strings.Split(card, "\n") {
			centeredLines = append(centeredLines, padStr+l)
		}
		card = strings.Join(centeredLines, "\n")
	}

	if !showWarning {
		return card
	}

	// Warning line below card
	warning := warningStyle.Render("Mini player mode - resize for full view")
	warnW := lipgloss.Width(warning)
	warnPad := (termWidth - warnW) / 2
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
