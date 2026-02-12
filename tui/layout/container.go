package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Container wraps content into an exact Width x Height bounding box.
// Lines are truncated/padded to Width and the line count is padded/truncated to Height.
// When content is truncated vertically, the last visible line shows a scroll indicator.
type Container struct {
	Width  int
	Height int
}

// Render returns the content constrained to exactly Width columns and Height lines.
func (c Container) Render(content string) string {
	lines := strings.Split(content, "\n")

	if len(lines) > c.Height {
		lines = lines[:c.Height]
		// Replace the last visible line with a scroll indicator
		indicator := lipgloss.NewStyle().Foreground(styles.Purple).Render("↓ More...")
		lines[c.Height-1] = PadToWidth(indicator, c.Width)
	}

	for len(lines) < c.Height {
		lines = append(lines, "")
	}

	for i, line := range lines {
		lines[i] = PadToWidth(line, c.Width)
	}

	return strings.Join(lines, "\n")
}
