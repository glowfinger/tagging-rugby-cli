package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
	"github.com/user/tagging-rugby-cli/tui/components"
	"github.com/user/tagging-rugby-cli/tui/layout"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// renderColumn1 renders Column 1: Playback status, selected tag detail.
func (m *Model) renderColumn1(width, height int) string {
	var lines []string

	// Video status card
	videoBox := components.RenderVideoBox(m.statusBar, width, false)
	lines = append(lines, strings.Split(videoBox, "\n")...)

	// Summary counts box
	noteCount := 0
	tackleCount := 0
	for _, item := range m.notesList.Items {
		if item.Type == components.ItemTypeNote {
			noteCount++
		} else {
			tackleCount++
		}
	}
	summaryStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	summaryLines := []string{
		summaryStyle.Render(fmt.Sprintf(" Notes:   %d", noteCount)),
		summaryStyle.Render(fmt.Sprintf(" Tackles: %d", tackleCount)),
		summaryStyle.Render(fmt.Sprintf(" Total:   %d", noteCount+tackleCount)),
	}
	summaryBox := components.RenderInfoBox("Summary", summaryLines, width)
	lines = append(lines, strings.Split(summaryBox, "\n")...)

	// Current tag detail card (selected item) — bordered box
	item := m.notesList.GetSelectedItem()
	if item != nil {
		detailStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
		dimStyle := lipgloss.NewStyle().Foreground(styles.Lavender)

		// Inner width = column width - 4 (2 border chars + 2 padding/space)
		innerW := width - 4
		if innerW < 10 {
			innerW = 10
		}

		typeStr := "Note"
		if item.Type == components.ItemTypeTackle {
			typeStr = "Tackle"
		}
		starStr := ""
		if item.Starred {
			starStr = " ★"
		}

		var contentLines []string
		contentLines = append(contentLines, detailStyle.Render(fmt.Sprintf(" #%d %s%s", item.ID, typeStr, starStr)))
		contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" @ %s", timeutil.FormatTime(item.TimestampSeconds))))
		if item.Category != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" [%s]", item.Category)))
		}
		if item.Player != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" Player: %s", item.Player)))
		}
		if item.Team != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" Team: %s", item.Team)))
		}
		if item.Text != "" {
			text := item.Text
			if len(text) > innerW {
				text = text[:innerW-3] + "..."
			}
			contentLines = append(contentLines, detailStyle.Render(" "+text))
		}

		infoBox := components.RenderInfoBox("Selected Tag", contentLines, width)
		lines = append(lines, strings.Split(infoBox, "\n")...)
	}

	return layout.Container{Width: width, Height: height}.Render(strings.Join(lines, "\n"))
}

// renderColumn2 renders Column 2: Scrollable list of all tags/events.
func (m *Model) renderColumn2(width, height int) string {
	// Reduce height by 2 for InfoBox top+bottom border lines
	innerHeight := height - 2
	if innerHeight < 3 {
		innerHeight = 3
	}

	// Render notes list with reduced width (InfoBox adds 2 border chars)
	notesOutput := components.NotesList(m.notesList, width-2, innerHeight, m.statusBar.TimePos)
	notesLines := strings.Split(notesOutput, "\n")

	infoBox := components.RenderInfoBox("Notes", notesLines, width)
	return layout.Container{Width: width, Height: height}.Render(infoBox)
}

// renderColumn3 renders Column 3: Live stats summary, bar graph, top players leaderboard.
func (m *Model) renderColumn3(width, height int) string {
	return layout.Container{Width: width, Height: height}.Render(
		components.StatsPanel(m.statsView.Stats, m.notesList.Items, width, height))
}

// renderColumn4 renders Column 4: Keybinding control groups (Playback, Navigation, Views).
func (m *Model) renderColumn4(width, height int) string {
	var lines []string

	groups := components.GetControlGroups()
	for i, group := range groups {
		box := components.RenderControlBox(group, width)
		lines = append(lines, box)

		// 1 blank line gap between bordered containers
		if i < len(groups)-1 {
			lines = append(lines, "")
		}
	}

	return layout.Container{Width: width, Height: height}.Render(strings.Join(lines, "\n"))
}

// formatStepSize formats the step size for display.
func formatStepSize(stepSize float64) string {
	if stepSize < 1 {
		return fmt.Sprintf("%.1fs", stepSize)
	}
	return fmt.Sprintf("%.0fs", stepSize)
}
