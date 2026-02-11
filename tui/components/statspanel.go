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
// It shows: stats summary, bar graph of event distribution, and tackle stats table.
func StatsPanel(tackleStats []PlayerStats, items []ListItem, width, height int, sortColumn SortColumn) string {
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
		return categories[i].Count > categories[j].Count
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

	// --- Tackle Stats Table ---
	tackleTable := renderTackleStatsTable(tackleStats, width, sortColumn)
	sections = append(sections, tackleTable)

	content := strings.Join(sections, "\n")
	return content
}

// renderTackleStatsTable renders the tackle stats table in a bordered box.
func renderTackleStatsTable(tackleStats []PlayerStats, width int, sortColumn SortColumn) string {
	var tableLines []string

	if len(tackleStats) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(styles.Purple).Italic(true)
		tableLines = append(tableLines, dimStyle.Render(" No tackle data"))
		return RenderInfoBox("Tackle Stats", tableLines, width)
	}

	// Sort by the specified column
	sorted := sortTackleStats(tackleStats, sortColumn)

	// Calculate totals
	var totalTackles, totalCompleted, totalMissed int
	for _, p := range tackleStats {
		totalTackles += p.Total
		totalCompleted += p.Completed
		totalMissed += p.Missed
	}
	totalPct := 0.0
	if totalCompleted+totalMissed > 0 {
		totalPct = (float64(totalCompleted) / float64(totalCompleted+totalMissed)) * 100
	}

	// Column widths (fixed for numeric columns, flexible for Player name)
	// Format: Player | Tot | Comp | Miss | Pct
	// Minimum widths: Tot(3), Comp(4), Miss(4), Pct(3) = 14 chars + spaces + borders
	// Header:  "Player | Tot | Comp | Miss | Pct"

	contentWidth := width - 2 // excluding borders from RenderInfoBox
	if contentWidth < 16 {
		contentWidth = 16 // minimum viable width
	}

	// Fixed column widths
	totWidth := 3
	compWidth := 4
	missWidth := 4
	pctWidth := 3

	// Calculate name column width (what's left)
	// Need: name + " | " + tot + " | " + comp + " | " + miss + " | " + pct
	// Separators: " | " appears 4 times = 12 chars
	separatorSpace := 12
	nameWidth := contentWidth - separatorSpace - totWidth - compWidth - missWidth - pctWidth
	if nameWidth < 4 {
		nameWidth = 4
	}

	// Styles
	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	activeHeaderStyle := lipgloss.NewStyle().Foreground(styles.Amber).Bold(true)
	totalsStyle := lipgloss.NewStyle().Foreground(styles.Cyan).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
	numStyle := lipgloss.NewStyle().Foreground(styles.Cyan)

	// Header row with sort indicators
	playerHeader := "Player"
	totHeader := "Tot"
	compHeader := "Comp"
	missHeader := "Miss"
	pctHeader := "Pct"

	switch sortColumn {
	case SortByPlayer:
		playerHeader = "Player↑"
	case SortByTotal:
		totHeader = "Tot↓"
	case SortByCompleted:
		compHeader = "Comp↓"
	case SortByMissed:
		missHeader = "Miss↓"
	case SortByPercentage:
		pctHeader = "Pct↓"
	}

	// Build header with conditional styling for active column
	styleFor := func(col SortColumn) lipgloss.Style {
		if col == sortColumn {
			return activeHeaderStyle
		}
		return headerStyle
	}
	header := styleFor(SortByPlayer).Render(fmt.Sprintf("%-*s", nameWidth, playerHeader)) +
		" | " +
		styleFor(SortByTotal).Render(fmt.Sprintf("%*s", totWidth, totHeader)) +
		" | " +
		styleFor(SortByCompleted).Render(fmt.Sprintf("%*s", compWidth, compHeader)) +
		" | " +
		styleFor(SortByMissed).Render(fmt.Sprintf("%*s", missWidth, missHeader)) +
		" | " +
		styleFor(SortByPercentage).Render(fmt.Sprintf("%*s", pctWidth, pctHeader))
	tableLines = append(tableLines, header)

	// Totals row (immediately below header)
	totalPctStr := "-"
	if totalCompleted+totalMissed > 0 {
		totalPctStr = fmt.Sprintf("%.0f%%", totalPct)
	}
	totalsRow := fmt.Sprintf("%-*s | %*d | %*d | %*d | %*s",
		nameWidth, "TOTAL",
		totWidth, totalTackles,
		compWidth, totalCompleted,
		missWidth, totalMissed,
		pctWidth, totalPctStr,
	)
	tableLines = append(tableLines, totalsStyle.Render(totalsRow))

	// Player rows
	for _, p := range sorted {
		name := truncateStr(p.Player, nameWidth)
		pctStr := "-"
		if p.Completed+p.Missed > 0 {
			pctStr = fmt.Sprintf("%.0f%%", p.Percentage)
		}
		row := fmt.Sprintf("%s | %s | %s | %s | %s",
			nameStyle.Render(fmt.Sprintf("%-*s", nameWidth, name)),
			numStyle.Render(fmt.Sprintf("%*d", totWidth, p.Total)),
			numStyle.Render(fmt.Sprintf("%*d", compWidth, p.Completed)),
			numStyle.Render(fmt.Sprintf("%*d", missWidth, p.Missed)),
			numStyle.Render(fmt.Sprintf("%*s", pctWidth, pctStr)),
		)
		tableLines = append(tableLines, row)
	}

	// Render in bordered box
	return RenderInfoBox("Tackle Stats", tableLines, width)
}

// sortTackleStats sorts tackle stats by the specified column.
// Returns a new sorted slice without mutating the input.
func sortTackleStats(stats []PlayerStats, sortColumn SortColumn) []PlayerStats {
	sorted := make([]PlayerStats, len(stats))
	copy(sorted, stats)

	switch sortColumn {
	case SortByTotal:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Total > sorted[j].Total
		})
	case SortByCompleted:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Completed > sorted[j].Completed
		})
	case SortByMissed:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Missed > sorted[j].Missed
		})
	case SortByPercentage:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Percentage > sorted[j].Percentage
		})
	case SortByPlayer:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Player < sorted[j].Player
		})
	default:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Total > sorted[j].Total
		})
	}

	return sorted
}
