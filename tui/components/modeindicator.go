package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ModeIndicator renders the mode indicator showing current focus and input mode.
// focusName is one of "Video", "Search", "Notes".
// mode is one of "Normal", "Search", "Command".
// Focus and Mode are displayed on separate lines within the InfoBox.
func ModeIndicator(focusName, mode string, width int) string {
	labelStyle := lipgloss.NewStyle().Foreground(styles.Lavender)
	valueStyle := lipgloss.NewStyle().Foreground(styles.LightLavender).Bold(true)

	innerW := width - 4 // InfoBox inner width (2 border + 2 padding)

	// Line 1: " Focus:  <value>" — label left, value right
	focusLabel := " Focus:"
	focusValue := focusName + " "
	focusPad := innerW - lipgloss.Width(focusLabel) - lipgloss.Width(focusValue)
	if focusPad < 1 {
		focusPad = 1
	}
	focusLine := labelStyle.Render(focusLabel) + strings.Repeat(" ", focusPad) + valueStyle.Render(focusValue)

	// Line 2: " Mode:   <value>" — label left, value right
	modeLabel := " Mode:"
	modeValue := mode + " "
	modePad := innerW - lipgloss.Width(modeLabel) - lipgloss.Width(modeValue)
	if modePad < 1 {
		modePad = 1
	}
	modeLine := labelStyle.Render(modeLabel) + strings.Repeat(" ", modePad) + valueStyle.Render(modeValue)

	return RenderInfoBox("Mode", []string{focusLine, modeLine}, width, false)
}
