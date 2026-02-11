package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// RenderMiniPlayer renders a bordered mini player card showing playback status.
// When fixedWidth > 0, the card uses that exact width instead of auto-sizing.
// When showWarning is true, a warning line is shown (e.g. for disconnected state).
func RenderMiniPlayer(state StatusBarState, fixedWidth int, showWarning bool) string {
	// Play/pause icon
	playIcon := "▶"
	if state.Paused {
		playIcon = "⏸"
	}

	// Speed display
	speed := state.Speed
	if speed == 0 {
		speed = 1.0
	}
	speedStr := fmt.Sprintf("%gx", speed)

	// Step size display
	stepStr := formatStepSize(state.StepSize)

	// Time display
	timeStr := formatTime(state.TimePos)
	durationStr := formatTime(state.Duration)

	// Build content lines
	var contentLines []string

	// Line 1: play state + step + speed
	line1 := fmt.Sprintf(" %s  Step: %s  Speed: %s", playIcon, stepStr, speedStr)
	contentLines = append(contentLines, line1)

	// Line 2: time position / duration
	line2 := fmt.Sprintf(" %s / %s", timeStr, durationStr)
	contentLines = append(contentLines, line2)

	// Line 3: mute label (only when muted)
	if state.Muted {
		contentLines = append(contentLines, " MUTED")
	}

	// Warning line (e.g. "mpv not connected")
	if showWarning {
		warnStyle := lipgloss.NewStyle().Foreground(styles.Red)
		contentLines = append(contentLines, warnStyle.Render(" ! Not connected"))
	}

	// Determine card width
	cardWidth := fixedWidth
	if cardWidth <= 0 {
		// Auto-size: find widest line + border padding
		maxW := 0
		for _, line := range contentLines {
			w := lipgloss.Width(line)
			if w > maxW {
				maxW = w
			}
		}
		cardWidth = maxW + 4 // 2 for borders + 2 for padding
	}

	// Inner width (subtract border chars)
	innerWidth := cardWidth - 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Styles
	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	borderColor := styles.Purple
	contentStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)

	// Build the bordered box with tab-style header
	// Tab header: ╭─ Playback ─────╮
	headerText := headerStyle.Render(" Playback ")
	headerTextWidth := lipgloss.Width(headerText)

	topBorderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topLeft := topBorderStyle.Render("╭─")
	topRight := topBorderStyle.Render("╮")
	topLeftWidth := 2 // ╭─
	topRightWidth := 1 // ╮
	fillWidth := innerWidth - topLeftWidth - headerTextWidth - topRightWidth + 2 // +2 for border chars counted in innerWidth
	if fillWidth < 0 {
		fillWidth = 0
	}
	topFill := ""
	for i := 0; i < fillWidth; i++ {
		topFill += "─"
	}
	topLine := topLeft + headerText + topBorderStyle.Render(topFill) + topRight

	// Content lines with side borders
	sideStyle := lipgloss.NewStyle().Foreground(borderColor)
	var renderedLines []string
	renderedLines = append(renderedLines, topLine)

	for _, line := range contentLines {
		styledLine := contentStyle.Render(line)
		lineWidth := lipgloss.Width(styledLine)
		pad := innerWidth - lineWidth
		if pad < 0 {
			pad = 0
		}
		padding := ""
		for i := 0; i < pad; i++ {
			padding += " "
		}
		renderedLines = append(renderedLines, sideStyle.Render("│")+styledLine+padding+sideStyle.Render("│"))
	}

	// Bottom border: ╰──────────────╯
	bottomFill := ""
	for i := 0; i < innerWidth; i++ {
		bottomFill += "─"
	}
	bottomLine := topBorderStyle.Render("╰"+bottomFill+"╯")
	renderedLines = append(renderedLines, bottomLine)

	result := ""
	for i, line := range renderedLines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}

	return result
}
