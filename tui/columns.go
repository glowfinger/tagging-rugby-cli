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

	// Playback status card
	statusHeader := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)
	lines = append(lines, statusHeader.Render(" Playback"))

	infoStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)

	playState := "▶ Playing"
	if m.statusBar.Paused {
		playState = "⏸ Paused"
	}
	lines = append(lines, infoStyle.Render(" "+playState))
	lines = append(lines, infoStyle.Render(fmt.Sprintf(" Time: %s / %s",
		timeutil.FormatTime(m.statusBar.TimePos),
		timeutil.FormatTime(m.statusBar.Duration))))
	lines = append(lines, infoStyle.Render(fmt.Sprintf(" Step: %s", formatStepSize(m.statusBar.StepSize))))

	if m.statusBar.Muted {
		lines = append(lines, infoStyle.Render(" 🔇 Muted"))
	}
	if m.statusBar.OverlayEnabled {
		lines = append(lines, infoStyle.Render(" Overlay: on"))
	} else {
		lines = append(lines, infoStyle.Render(" Overlay: off"))
	}
	lines = append(lines, "")

	// Current tag detail card (selected item)
	item := m.notesList.GetSelectedItem()
	if item != nil {
		detailHeader := lipgloss.NewStyle().
			Foreground(styles.Pink).
			Bold(true)
		lines = append(lines, detailHeader.Render(" Selected Tag"))

		detailStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
		dimStyle := lipgloss.NewStyle().Foreground(styles.Lavender)

		typeStr := "Note"
		if item.Type == components.ItemTypeTackle {
			typeStr = "Tackle"
		}
		starStr := ""
		if item.Starred {
			starStr = " ★"
		}
		lines = append(lines, detailStyle.Render(fmt.Sprintf(" #%d %s%s", item.ID, typeStr, starStr)))
		lines = append(lines, dimStyle.Render(fmt.Sprintf(" @ %s", timeutil.FormatTime(item.TimestampSeconds))))
		if item.Category != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" [%s]", item.Category)))
		}
		if item.Player != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" Player: %s", item.Player)))
		}
		if item.Team != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" Team: %s", item.Team)))
		}
		if item.Text != "" {
			text := item.Text
			maxTextW := width - 3
			if maxTextW < 10 {
				maxTextW = 10
			}
			if len(text) > maxTextW {
				text = text[:maxTextW-3] + "..."
			}
			lines = append(lines, detailStyle.Render(" "+text))
		}
		lines = append(lines, "")
	}

	return layout.Container{Width: width, Height: height}.Render(strings.Join(lines, "\n"))
}

// renderColumn2 renders Column 2: Scrollable list of all tags/events.
func (m *Model) renderColumn2(width, height int) string {
	// Use a taller list that fills the column
	listHeight := height
	if listHeight < 3 {
		listHeight = 3
	}

	return layout.Container{Width: width, Height: height}.Render(
		components.NotesList(m.notesList, width, listHeight, m.statusBar.TimePos))
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
