package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Timeline renders a progress bar with event markers spanning full terminal width.
// It shows playback position, timestamps, and note/tackle markers.
func Timeline(timePos, duration float64, items []ListItem, width int) string {
	if width < 20 {
		return ""
	}

	// Styles
	filledStyle := lipgloss.NewStyle().Foreground(styles.BrightPurple)
	unfilledStyle := lipgloss.NewStyle().Foreground(styles.Purple)
	timeStyle := lipgloss.NewStyle().Foreground(styles.LightLavender).Bold(true)
	markerStyle := lipgloss.NewStyle().Foreground(styles.Cyan)
	posStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)

	// Format timestamps
	currentStr := timeutil.FormatTime(timePos)
	totalStr := timeutil.FormatTime(duration)
	timeDisplay := fmt.Sprintf(" %s / %s", currentStr, totalStr)
	timeDisplayWidth := lipgloss.Width(timeDisplay)

	// Bar width = total width minus time display and spacing
	barWidth := width - timeDisplayWidth - 2 // 1 space left margin + 1 space before time
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

	// Build the bar line
	barLine := " " + barBuilder.String() + " " + timeStyle.Render(timeDisplay)

	// Build position indicator line below the bar
	var indicatorBuilder strings.Builder
	indicatorBuilder.WriteString(" ")
	for i := 0; i < barWidth; i++ {
		if i == fillPos {
			indicatorBuilder.WriteString(posStyle.Render("▲"))
		} else {
			indicatorBuilder.WriteString(" ")
		}
	}

	// Apply background style to both lines
	bgStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Width(width)

	return bgStyle.Render(barLine) + "\n" + bgStyle.Render(indicatorBuilder.String())
}
