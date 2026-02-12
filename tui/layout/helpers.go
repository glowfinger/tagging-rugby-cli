package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// PadToWidth pads or truncates a string to exactly the specified width.
// Uses ansi.Truncate for ANSI-aware, grapheme-aware truncation that correctly
// handles double-width characters (emoji, East-Asian).
func PadToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	currentWidth := lipgloss.Width(s)
	if currentWidth == width {
		return s
	}
	if currentWidth > width {
		s = ansi.Truncate(s, width, "")
		currentWidth = lipgloss.Width(s)
	}
	if currentWidth < width {
		return s + strings.Repeat(" ", width-currentWidth)
	}
	return s
}

// NormalizeLines pads or truncates a slice of strings to exactly the given height.
func NormalizeLines(lines []string, height int) []string {
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}
