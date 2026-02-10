package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Timeline renders a progress bar with event markers in a bordered container.
// It shows playback position, timestamps, and note/tackle markers.
// Total output height is 6 lines: 3 tab header + 2 content rows + 1 bottom border.
func Timeline(timePos, duration float64, items []ListItem, width int) string {
	if width < 20 {
		return ""
	}

	// Inner width = width - 4 (2 border chars + 2 padding spaces)
	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Styles
	filledStyle := lipgloss.NewStyle().Foreground(styles.BrightPurple)
	unfilledStyle := lipgloss.NewStyle().Foreground(styles.Purple)
	timeStyle := lipgloss.NewStyle().Foreground(styles.LightLavender).Bold(true)
	markerStyle := lipgloss.NewStyle().Foreground(styles.Cyan)
	posStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)

	// Format timestamps
	currentStr := formatTime(timePos)
	totalStr := formatTime(duration)
	timeDisplay := fmt.Sprintf(" %s / %s", currentStr, totalStr)
	timeDisplayWidth := lipgloss.Width(timeDisplay)

	// Bar width = inner width minus time display and spacing
	barWidth := innerWidth - timeDisplayWidth - 2 // 1 space left margin + 1 space before time
	if barWidth < 10 {
		barWidth = 10
	}

	// Calculate fill position
	var fillPos int
	if duration > 0 {
		fillPos = int(math.Round(float64(barWidth) * timePos / duration))
	}
	if fillPos < 0 {
		fillPos = 0
	}
	if fillPos > barWidth {
		fillPos = barWidth
	}

	// Build the bar with event markers
	barChars := make([]rune, barWidth)
	markerPositions := make([]bool, barWidth)

	// Place event markers
	if duration > 0 {
		for _, item := range items {
			pos := int(math.Round(float64(barWidth-1) * item.TimestampSeconds / duration))
			if pos >= 0 && pos < barWidth {
				markerPositions[pos] = true
			}
		}
	}

	// Fill bar characters
	for i := 0; i < barWidth; i++ {
		if markerPositions[i] {
			barChars[i] = '◆'
		} else if i < fillPos {
			barChars[i] = '━'
		} else if i == fillPos {
			barChars[i] = '╸'
		} else {
			barChars[i] = '─'
		}
	}

	// Render the bar with appropriate colors per character
	var barBuilder strings.Builder
	for i, ch := range barChars {
		s := string(ch)
		if markerPositions[i] {
			barBuilder.WriteString(markerStyle.Render(s))
		} else if i < fillPos {
			barBuilder.WriteString(filledStyle.Render(s))
		} else if i == fillPos {
			barBuilder.WriteString(posStyle.Render(s))
		} else {
			barBuilder.WriteString(unfilledStyle.Render(s))
		}
	}

	// Build the bar line (content row 1)
	barLine := " " + barBuilder.String() + " " + timeStyle.Render(timeDisplay)

	// Build position indicator line (content row 2)
	var indicatorBuilder strings.Builder
	indicatorBuilder.WriteString(" ")
	for i := 0; i < barWidth; i++ {
		if i == fillPos {
			indicatorBuilder.WriteString(posStyle.Render("▲"))
		} else {
			indicatorBuilder.WriteString(" ")
		}
	}

	// Build bordered box with tab-style "Timeline" header
	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	borderColor := styles.Purple
	topBorderStyle := lipgloss.NewStyle().Foreground(borderColor)
	sideStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Box inner width (just border chars, content has its own padding)
	boxInner := width - 2

	// Tab header: ╭─ Timeline ─────╮
	headerText := headerStyle.Render(" Timeline ")
	headerTextWidth := lipgloss.Width(headerText)
	topLeft := topBorderStyle.Render("╭─")
	topRight := topBorderStyle.Render("╮")
	fillWidth := boxInner - 2 - headerTextWidth - 1 + 2
	if fillWidth < 0 {
		fillWidth = 0
	}
	topLine := topLeft + headerText + topBorderStyle.Render(strings.Repeat("─", fillWidth)) + topRight

	// Wrap content lines in side borders
	wrapLine := func(content string) string {
		contentWidth := lipgloss.Width(content)
		pad := boxInner - contentWidth
		if pad < 0 {
			pad = 0
		}
		return sideStyle.Render("│") + content + strings.Repeat(" ", pad) + sideStyle.Render("│")
	}

	// Empty line for inner padding
	emptyLine := wrapLine("")

	// Bottom border: ╰──────────────╯
	bottomLine := topBorderStyle.Render("╰" + strings.Repeat("─", boxInner) + "╯")

	// Total: 6 lines — top border (1) + padding (2) + bar (3) + indicator (4) + padding (5) + bottom (6)
	return topLine + "\n" + emptyLine + "\n" + wrapLine(barLine) + "\n" + wrapLine(indicatorBuilder.String()) + "\n" + emptyLine + "\n" + bottomLine
}
