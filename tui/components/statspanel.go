// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// CategoryCount holds the count for a single event category.
type CategoryCount struct {
	Name  string
	Count int
}

// StatsPanel renders a live stats panel for column 3 of the three-column layout.
// It shows: stats summary, bar graph of event distribution, and top players leaderboard.
func StatsPanel(tackleStats []PlayerStats, items []ListItem, width, height int) string {
	if width < 5 {
		return ""
	}

	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)
	sections = append(sections, titleStyle.Render("Live Stats"))
	sections = append(sections, "")

	// --- Stats summary ---
	summaryHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)
	sections = append(sections, summaryHeaderStyle.Render("Summary"))

	noteCount := 0
	tackleCount := 0
	for _, item := range items {
		if item.Type == ItemTypeNote {
			noteCount++
		} else {
			tackleCount++
		}
	}

	infoStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	sections = append(sections, infoStyle.Render(fmt.Sprintf(" Notes:   %d", noteCount)))
	sections = append(sections, infoStyle.Render(fmt.Sprintf(" Tackles: %d", tackleCount)))
	sections = append(sections, infoStyle.Render(fmt.Sprintf(" Total:   %d", noteCount+tackleCount)))
	sections = append(sections, "")

	// --- Bar graph of event distribution ---
	sections = append(sections, summaryHeaderStyle.Render("Event Distribution"))

	// Count categories from items
	catCounts := make(map[string]int)
	for _, item := range items {
		cat := item.Category
		if item.Type == ItemTypeTackle {
			cat = "tackle"
		}
		if cat == "" {
			cat = "other"
		}
		catCounts[cat]++
	}

	// Sort by count descending
	var categories []CategoryCount
	for name, count := range catCounts {
		categories = append(categories, CategoryCount{Name: name, Count: count})
	}
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].Count != categories[j].Count {
			return categories[i].Count > categories[j].Count
		}
		return categories[i].Name < categories[j].Name
	})

	// Render bar graph (max 6 categories)
	maxDisplay := 6
	if len(categories) < maxDisplay {
		maxDisplay = len(categories)
	}

	if len(categories) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(styles.Purple).Italic(true)
		sections = append(sections, dimStyle.Render(" No events yet"))
	} else {
		maxCount := categories[0].Count
		barMaxWidth := width - 16 // label (8) + count (4) + padding (4)
		if barMaxWidth < 5 {
			barMaxWidth = 5
		}

		labelStyle := lipgloss.NewStyle().Foreground(styles.Lavender)
		barStyle := lipgloss.NewStyle().Foreground(styles.BrightPurple)
		countStyle := lipgloss.NewStyle().Foreground(styles.Cyan)

		for i := 0; i < maxDisplay; i++ {
			cat := categories[i]
			label := truncateStr(cat.Name, 8)
			barLen := 1
			if maxCount > 0 {
				barLen = (cat.Count * barMaxWidth) / maxCount
				if barLen < 1 {
					barLen = 1
				}
			}
			bar := strings.Repeat("█", barLen)
			sections = append(sections, fmt.Sprintf(" %s %s %s",
				labelStyle.Render(fmt.Sprintf("%-8s", label)),
				barStyle.Render(bar),
				countStyle.Render(fmt.Sprintf("%d", cat.Count)),
			))
		}
	}
	sections = append(sections, "")

	// --- Top Players Leaderboard ---
	sections = append(sections, summaryHeaderStyle.Render("Top Players"))

	if len(tackleStats) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(styles.Purple).Italic(true)
		sections = append(sections, dimStyle.Render(" No tackle data"))
	} else {
		// Sort by total tackles descending for leaderboard
		sorted := make([]PlayerStats, len(tackleStats))
		copy(sorted, tackleStats)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Total != sorted[j].Total {
				return sorted[i].Total > sorted[j].Total
			}
			return sorted[i].Player < sorted[j].Player
		})

		// Show top 5 players
		maxPlayers := 5
		if len(sorted) < maxPlayers {
			maxPlayers = len(sorted)
		}

		nameStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
		numStyle := lipgloss.NewStyle().Foreground(styles.Cyan).Bold(true)
		pctStyle := lipgloss.NewStyle().Foreground(styles.Lavender)
		medalIcons := []string{"#1", "#2", "#3", "#4", "#5"}

		nameWidth := width - 18 // medal(2) + total(4) + pct(6) + padding(6)
		if nameWidth < 6 {
			nameWidth = 6
		}

		for i := 0; i < maxPlayers; i++ {
			p := sorted[i]
			medal := medalIcons[i]
			name := truncateStr(p.Player, nameWidth)
			pctStr := "-"
			if p.Completed+p.Missed > 0 {
				pctStr = fmt.Sprintf("%.0f%%", p.Percentage)
			}
			sections = append(sections, fmt.Sprintf(" %s %s %s %s",
				medal,
				nameStyle.Render(fmt.Sprintf("%-*s", nameWidth, name)),
				numStyle.Render(fmt.Sprintf("%3d", p.Total)),
				pctStyle.Render(fmt.Sprintf("%4s", pctStr)),
			))
		}
	}

	content := strings.Join(sections, "\n")
	return content
}
