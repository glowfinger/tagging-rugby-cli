package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Responsive layout constants.
const (
	MinTerminalWidth  = 80 // minimum terminal width for multi-column layout
	Col3HideThreshold = 90 // below this width, hide column 3 entirely
	Col3MinWidth      = 18 // minimum width for column 3 before hiding
)

// ComputeColumnWidths calculates responsive column widths based on terminal width.
// Returns individual column widths and whether column 3 should be shown.
// At >=120 width: equal thirds. At 90-119: min-width col3, rest split evenly.
// At 80-89: two-column layout (col3 hidden).
func ComputeColumnWidths(termWidth int) (col1, col2, col3 int, showCol3 bool) {
	showCol3 = termWidth >= Col3HideThreshold

	if showCol3 {
		// Three-column layout: account for 2 border characters
		usableWidth := termWidth - 2

		if termWidth >= 120 {
			// Wide: equal thirds
			col1 = usableWidth / 3
			col2 = usableWidth / 3
			col3 = usableWidth - col1 - col2
		} else {
			// Medium: column 3 gets minimum, rest splits between 1 and 2
			col3 = Col3MinWidth
			remaining := usableWidth - col3
			col1 = remaining / 2
			col2 = remaining - col1
		}
	} else {
		// Two-column layout: 1 border character
		usableWidth := termWidth - 1
		col1 = usableWidth / 2
		col2 = usableWidth - col1
		col3 = 0
	}

	return
}

// JoinColumns joins pre-rendered column strings side by side with purple border separators.
// Each column is normalized to the given height and padded to its width.
func JoinColumns(columns []string, widths []int, height int) string {
	borderStr := lipgloss.NewStyle().
		Foreground(styles.Purple).
		Render("│")

	// Split each column into lines and normalize to height
	colLines := make([][]string, len(columns))
	for i, col := range columns {
		colLines[i] = NormalizeLines(strings.Split(col, "\n"), height)
	}

	var rows []string
	for row := 0; row < height; row++ {
		var parts []string
		for i, lines := range colLines {
			parts = append(parts, PadToWidth(lines[row], widths[i]))
		}
		rows = append(rows, strings.Join(parts, borderStr))
	}

	return strings.Join(rows, "\n")
}
