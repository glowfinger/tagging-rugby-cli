package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ModeIndicator renders the mode indicator showing current focus and input mode.
// focusName is one of "Video", "Search", "Notes".
// mode is one of "Normal", "Search", "Command".
func ModeIndicator(focusName, mode string, width int) string {
	textStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)

	left := " Focus: " + focusName
	right := mode + " "

	innerW := width - 4 // InfoBox inner width (2 border + 2 padding)
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := innerW - leftW - rightW
	if pad < 1 {
		pad = 1
	}

	line := textStyle.Render(left + strings.Repeat(" ", pad) + right)
	return RenderInfoBox("Mode", []string{line}, width, false)
}
