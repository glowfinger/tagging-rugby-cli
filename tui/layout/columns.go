package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Responsive layout constants.
const (
	Col1Width         = 30  // fixed width for column 1
	Col3HideThreshold = 90  // below this width, hide column 3 entirely
	Col3MinWidth      = 18  // minimum width for column 3 before hiding
	Col4Width         = 30  // fixed width for column 4 (controls)
	Col4ShowThreshold = 160 // show column 4 when terminal width >= this
)

// ComputeColumnWidths calculates responsive column widths based on terminal width.
// Returns individual column widths and whether columns 3 and 4 should be shown.
// Column 1 is always fixed at Col1Width (30). Column 4 is fixed at Col4Width (30).
// At >=Col4ShowThreshold: 4-column layout. At >=Col3HideThreshold: 3-column layout.
// At <Col3HideThreshold: 2-column layout.
func ComputeColumnWidths(termWidth int) (col1, col2, col3, col4 int, showCol3, showCol4 bool) {
	showCol4 = termWidth >= Col4ShowThreshold
	showCol3 = termWidth >= Col3HideThreshold
	col1 = Col1Width

	if showCol4 {
		// Four-column layout: account for 3 border characters
		usableWidth := termWidth - 3
		col4 = Col4Width
		remaining := usableWidth - col1 - col4
		col2 = remaining / 2
		col3 = remaining - col2
	} else if showCol3 {
		// Three-column layout: account for 2 border characters
		usableWidth := termWidth - 2
		remaining := usableWidth - col1
		col2 = remaining / 2
		col3 = remaining - col2
	} else {
		// Two-column layout: 1 border character
		usableWidth := termWidth - 1
		col2 = usableWidth - col1
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
