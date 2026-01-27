// Package components provides reusable TUI components.
package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// StatusBarState holds the current playback state for the status bar.
type StatusBarState struct {
	// Paused indicates if playback is paused
	Paused bool
	// Muted indicates if audio is muted
	Muted bool
	// TimePos is the current playback position in seconds
	TimePos float64
	// Duration is the total video duration in seconds
	Duration float64
	// StepSize is the current seek step size in seconds
	StepSize float64
	// OverlayEnabled indicates if the mpv overlay is enabled
	OverlayEnabled bool
}

// StatusBar renders the status bar component.
// The status bar displays play/pause icon, current timestamp, duration, step size,
// mute icon when muted, and overlay icon when enabled.
func StatusBar(state StatusBarState, width int) string {
	// Play/pause icon
	var playIcon string
	if state.Paused {
		playIcon = "⏸"
	} else {
		playIcon = "▶"
	}

	// Format timestamps as MM:SS
	timeStr := formatTime(state.TimePos)
	durationStr := formatTime(state.Duration)

	// Step size display
	stepStr := formatStepSize(state.StepSize)

	// Mute icon (only shown when muted)
	var muteIcon string
	if state.Muted {
		muteIcon = " 🔇"
	}

	// Overlay icon (only shown when enabled)
	var overlayIcon string
	if state.OverlayEnabled {
		overlayIcon = " 📺"
	}

	// Build the status bar content
	leftContent := fmt.Sprintf(" %s %s / %s", playIcon, timeStr, durationStr)
	rightContent := fmt.Sprintf("Step: %s%s%s ", stepStr, muteIcon, overlayIcon)

	// Calculate padding between left and right content
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)
	padding := width - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	// Create padded middle section
	middle := ""
	for i := 0; i < padding; i++ {
		middle += " "
	}

	// Build the full status bar
	content := leftContent + middle + rightContent

	// Apply style
	statusBarStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Foreground(styles.LightLavender).
		Bold(true).
		Width(width)

	return statusBarStyle.Render(content)
}

// formatTime formats seconds as MM:SS.
func formatTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalSeconds := int(seconds)
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

// formatStepSize formats the step size for display.
// Shows decimal for values less than 1, otherwise whole number.
func formatStepSize(stepSize float64) string {
	if stepSize < 1 {
		return fmt.Sprintf("%.1fs", stepSize)
	}
	return fmt.Sprintf("%.0fs", stepSize)
}
