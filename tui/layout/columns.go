package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// Responsive layout constants.
const (
	Col1Width         = 30  // fixed width for column 1
	Col2TargetWidth   = 80  // target width for column 2
	Col3TargetWidth   = 60  // target width for column 3
	ColMinWidth       = 30  // minimum width before a column is hidden
	Col4Width         = 30  // fixed width for column 4 (controls)
	Col4ShowThreshold = 170 // show column 4 when terminal width >= this
)

// ComputeColumnWidths calculates responsive column widths based on terminal width.
// Returns individual column widths and whether columns 2, 3, and 4 should be shown.
// Column 1 is always fixed at Col1Width (30). Column 4 is fixed at Col4Width (30).
// Hide order: Col4 first (below 170), then Col3 (below 30 cells), then Col2 (below 30 cells).
// Col1 is always visible at any terminal width.
func ComputeColumnWidths(termWidth int) (col1, col2, col3, col4 int, showCol2, showCol3, showCol4 bool) {
	col1 = Col1Width

	// Step 1: Determine if col4 is shown
	showCol4 = termWidth >= Col4ShowThreshold

	// Step 2: Calculate available space for col2 and col3
	borders := 0
	fixedUsed := col1
	if showCol4 {
		col4 = Col4Width
		fixedUsed += col4
		borders = 3 // col1|col2|col3|col4
	}

	// Try 3-column layout (col1 + col2 + col3 [+ col4])
	if !showCol4 {
		borders = 2 // col1|col2|col3
	}
	usable := termWidth - fixedUsed - borders
	if usable >= ColMinWidth*2 {
		// Enough room for both col2 and col3
		showCol2 = true
		showCol3 = true
		// Distribute proportionally based on targets
		totalTarget := Col2TargetWidth + Col3TargetWidth
		col2 = usable * Col2TargetWidth / totalTarget
		col3 = usable - col2
		// Clamp: if col3 would be below min, give space to col2
		if col3 < ColMinWidth {
			col3 = ColMinWidth
			col2 = usable - col3
		}
		if col2 < ColMinWidth {
			col2 = ColMinWidth
			col3 = usable - col2
		}
		return
	}

	// Try 2-column layout (col1 + col2 [+ col4])
	showCol3 = false
	col3 = 0
	if showCol4 {
		borders = 2 // col1|col2|col4
	} else {
		borders = 1 // col1|col2
	}
	usable = termWidth - fixedUsed - borders
	if usable >= ColMinWidth {
		showCol2 = true
		col2 = usable
		return
	}

	// Only col1 [+ col4] visible
	showCol2 = false
	col2 = 0
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
