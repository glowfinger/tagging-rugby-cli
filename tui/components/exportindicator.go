// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ExportIndicatorState holds the current export progress counts.
type ExportIndicatorState struct {
	TotalTackles   int
	CompletedClips int
	PendingClips   int
	ErrorClips     int
}

// ExportStatus returns a human-readable status string based on the current counts.
// Priority: Completed > Processing > Error > Ready
func (s ExportIndicatorState) ExportStatus() string {
	switch {
	case s.TotalTackles > 0 && s.CompletedClips == s.TotalTackles && s.PendingClips == 0 && s.ErrorClips == 0:
		return "Completed"
	case s.PendingClips > 0:
		return "Processing"
	case s.ErrorClips > 0 && s.PendingClips == 0:
		return "Error"
	default:
		return "Ready"
	}
}

// ExportIndicator renders a 5-line InfoBox showing the current export progress.
func ExportIndicator(state ExportIndicatorState, width int) string {
	if width < 4 {
		return ""
	}

	// innerWidth is the usable content area: border (2) + padding spaces (2)
	innerWidth := width - 4

	// --- Row 1: Status ---
	status := state.ExportStatus()
	var statusColor lipgloss.Color
	switch status {
	case "Ready":
		statusColor = styles.Lavender
	case "Processing":
		statusColor = styles.Amber
	case "Error":
		statusColor = styles.Red
	case "Completed":
		statusColor = styles.Green
	}
	statusValue := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	statusValueWidth := lipgloss.Width(statusValue)
	statusLabel := "Status:"
	statusLabelWidth := len(statusLabel)
	statusPad := innerWidth - statusLabelWidth - statusValueWidth
	if statusPad < 0 {
		statusPad = 0
	}
	statusLine := " " + statusLabel + strings.Repeat(" ", statusPad) + statusValue + " "

	// --- Row 2: Clips count ---
	clipsValue := lipgloss.NewStyle().Foreground(styles.LightLavender).Render(
		fmt.Sprintf("%d/%d", state.CompletedClips, state.TotalTackles),
	)
	clipsValueWidth := lipgloss.Width(clipsValue)
	clipsLabel := "Clips:"
	clipsLabelWidth := len(clipsLabel)
	clipsPad := innerWidth - clipsLabelWidth - clipsValueWidth
	if clipsPad < 0 {
		clipsPad = 0
	}
	clipsLine := " " + clipsLabel + strings.Repeat(" ", clipsPad) + clipsValue + " "

	// --- Row 3: Progress bar ---
	fraction := 0.0
	if state.TotalTackles > 0 {
		fraction = float64(state.CompletedClips) / float64(state.TotalTackles)
		if fraction > 1.0 {
			fraction = 1.0
		}
		if fraction < 0.0 {
			fraction = 0.0
		}
	}
	filled := int(fraction * float64(innerWidth))
	empty := innerWidth - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	barLine := " " + lipgloss.NewStyle().Foreground(styles.Cyan).Render(bar) + " "

	contentLines := []string{statusLine, clipsLine, barLine}
	return RenderInfoBox("Export", contentLines, width, false)
}
