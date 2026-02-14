package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ExportProgressState holds the state for the export progress display.
type ExportProgressState struct {
	Active      bool
	Total       int
	Completed   int
	Errors      int
	CurrentFile string
}

// ExportProgress renders a bordered info box showing export progress.
// It displays a progress bar, percentage, clip counter, current file, and error count.
func ExportProgress(state ExportProgressState, width int) string {
	if !state.Active || width < 10 {
		return ""
	}

	greenStyle := lipgloss.NewStyle().Foreground(styles.Green)
	amberStyle := lipgloss.NewStyle().Foreground(styles.Amber)
	redStyle := lipgloss.NewStyle().Foreground(styles.Red)
	textStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)

	// Inner width for content (box border = 2, plus 1 space padding each side)
	innerW := width - 4
	if innerW < 6 {
		innerW = 6
	}

	var contentLines []string

	// Progress bar
	var pct int
	if state.Total > 0 {
		pct = state.Completed * 100 / state.Total
	}

	// Bar width: innerW minus " XXX%" label (5 chars) minus 1 space padding
	barWidth := innerW - 6
	if barWidth < 4 {
		barWidth = 4
	}

	filled := 0
	if state.Total > 0 {
		filled = barWidth * state.Completed / state.Total
	}
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := greenStyle.Render(strings.Repeat("█", filled)) + amberStyle.Render(strings.Repeat("░", empty))
	pctLabel := fmt.Sprintf(" %3d%%", pct)
	contentLines = append(contentLines, " "+bar+textStyle.Render(pctLabel))

	// Clip counter
	counterLine := fmt.Sprintf(" %d/%d clips", state.Completed, state.Total)
	if state.Errors > 0 {
		counterLine += "  " + redStyle.Render(fmt.Sprintf("%d errors", state.Errors))
	}
	contentLines = append(contentLines, textStyle.Render(counterLine))

	// Completion or current file
	if state.Completed == state.Total && state.Total > 0 {
		contentLines = append(contentLines, " "+greenStyle.Render("Export complete"))
	} else if state.CurrentFile != "" {
		// Truncate filename to fit
		maxFileW := innerW - 2 // 1 space padding each side
		fileDisplay := state.CurrentFile
		if lipgloss.Width(fileDisplay) > maxFileW {
			fileDisplay = ansi.Truncate(fileDisplay, maxFileW-3, "...")
		}
		contentLines = append(contentLines, " "+textStyle.Render(fileDisplay))
	}

	return RenderInfoBox("Export", contentLines, width)
}
