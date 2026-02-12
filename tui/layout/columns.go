package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Responsive layout constants.
const (
	MinTerminalWidth  = 80 // minimum terminal width for multi-column layout
	Col1Width         = 30 // fixed width for column 1
	Col3HideThreshold = 90 // below this width, hide column 3 entirely
	Col3MinWidth      = 18 // minimum width for column 3 before hiding
)

// ComputeColumnWidths calculates responsive column widths based on terminal width.
// Returns individual column widths and whether column 3 should be shown.
// Column 1 is always fixed at Col1Width (30). Remaining space is distributed
// between columns 2 and 3. At <Col3HideThreshold: two-column layout (col3 hidden).
func ComputeColumnWidths(termWidth int) (col1, col2, col3 int, showCol3 bool) {
	showCol3 = termWidth >= Col3HideThreshold
	col1 = Col1Width

	if showCol3 {
		// Three-column layout: account for 2 border characters
		remaining := termWidth - 2 - col1
		col2 = remaining / 2
		col3 = remaining - col2
	} else {
		// Two-column layout: 1 border character
		col2 = termWidth - 1 - col1
		col3 = 0
	}

	return
}

// JoinColumns joins pre-rendered column strings side by side with purple border separators.
// Columns should already be containerized (exact Width x Height) via Container.Render.
// NormalizeLines/PadToWidth are still applied as a safety net.
func JoinColumns(columns []string, widths []int, height int) string {
	borderStr := lipgloss.NewStyle().
		Foreground(styles.Purple).
		Render("│")

	colLines := make([][]string, len(columns))
	for i, col := range columns {
		colLines[i] = strings.Split(col, "\n")
	}

	var rows []string
	for row := 0; row < height; row++ {
		var parts []string
		for i, lines := range colLines {
			if row < len(lines) {
				parts = append(parts, lines[row])
			} else {
				parts = append(parts, PadToWidth("", widths[i]))
			}
		}
		rows = append(rows, strings.Join(parts, borderStr))
	}

	return strings.Join(rows, "\n")
}
