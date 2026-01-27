// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// SortColumn represents which column the stats table is sorted by.
type SortColumn int

const (
	// SortByPlayer sorts by player name (alphabetically)
	SortByPlayer SortColumn = iota
	// SortByTotal sorts by total tackles
	SortByTotal
	// SortByCompleted sorts by completed tackles
	SortByCompleted
	// SortByMissed sorts by missed tackles
	SortByMissed
	// SortByPossible sorts by possible tackles
	SortByPossible
	// SortByPercentage sorts by completion percentage
	SortByPercentage
	// SortByStarred sorts by starred count
	SortByStarred
)

// PlayerStats holds tackle statistics for a single player.
type PlayerStats struct {
	// Player is the player name
	Player string
	// Total is the total number of tackles
	Total int
	// Completed is the number of completed tackles
	Completed int
	// Missed is the number of missed tackles
	Missed int
	// Possible is the number of possible tackles
	Possible int
	// Other is the number of other tackles
	Other int
	// Starred is the number of starred tackles
	Starred int
	// Percentage is the completion percentage (Completed / (Completed + Missed) * 100)
	Percentage float64
}

// StatsViewState holds the state for the stats view component.
type StatsViewState struct {
	// Active indicates if the stats view is currently displayed
	Active bool
	// Stats is the list of player statistics
	Stats []PlayerStats
	// SortColumn is the current sort column
	SortColumn SortColumn
	// AllVideos indicates if showing stats for all videos (true) or current video only (false)
	AllVideos bool
	// SelectedIndex is the currently selected row
	SelectedIndex int
	// ScrollOffset is the scroll position
	ScrollOffset int
	// FilterMode indicates if filter input mode is active
	FilterMode bool
	// FilterInput is the current filter text being typed
	FilterInput string
	// FilteredPlayers is a set of player names that are currently filtered (highlighted)
	FilteredPlayers map[string]bool
}

// SortStats sorts the stats by the current sort column.
func (s *StatsViewState) SortStats() {
	switch s.SortColumn {
	case SortByPlayer:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Player < s.Stats[j].Player
		})
	case SortByTotal:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Total > s.Stats[j].Total
		})
	case SortByCompleted:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Completed > s.Stats[j].Completed
		})
	case SortByMissed:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Missed > s.Stats[j].Missed
		})
	case SortByPossible:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Possible > s.Stats[j].Possible
		})
	case SortByPercentage:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Percentage > s.Stats[j].Percentage
		})
	case SortByStarred:
		sort.Slice(s.Stats, func(i, j int) bool {
			return s.Stats[i].Starred > s.Stats[j].Starred
		})
	}
}

// NextSortColumn cycles to the next sort column.
func (s *StatsViewState) NextSortColumn() {
	s.SortColumn = (s.SortColumn + 1) % 7
	s.SortStats()
}

// MoveUp moves the selection up in the list.
func (s *StatsViewState) MoveUp() {
	if s.SelectedIndex > 0 {
		s.SelectedIndex--
	}
}

// MoveDown moves the selection down in the list.
func (s *StatsViewState) MoveDown() {
	if s.SelectedIndex < len(s.Stats)-1 {
		s.SelectedIndex++
	}
}

// ToggleFilter toggles the filter state for a player by name or initials.
// Returns true if a player was matched and toggled.
func (s *StatsViewState) ToggleFilter(input string) bool {
	if s.FilteredPlayers == nil {
		s.FilteredPlayers = make(map[string]bool)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return false
	}

	// Find players matching the input (name contains or initials match)
	var matched []string
	for _, stat := range s.Stats {
		playerLower := strings.ToLower(stat.Player)
		// Check if input is contained in player name
		if strings.Contains(playerLower, input) {
			matched = append(matched, stat.Player)
		} else if matchesInitials(stat.Player, input) {
			// Check if input matches initials
			matched = append(matched, stat.Player)
		}
	}

	// If exactly one match, toggle that player
	if len(matched) == 1 {
		player := matched[0]
		if s.FilteredPlayers[player] {
			delete(s.FilteredPlayers, player)
		} else {
			s.FilteredPlayers[player] = true
		}
		return true
	}

	// If multiple matches, toggle all of them
	if len(matched) > 1 {
		// Check if all are currently filtered
		allFiltered := true
		for _, player := range matched {
			if !s.FilteredPlayers[player] {
				allFiltered = false
				break
			}
		}
		// If all filtered, remove all; otherwise add all
		for _, player := range matched {
			if allFiltered {
				delete(s.FilteredPlayers, player)
			} else {
				s.FilteredPlayers[player] = true
			}
		}
		return len(matched) > 0
	}

	return false
}

// matchesInitials checks if the input matches the initials of a player name.
// For example, "jd" matches "John Doe", "js" matches "John Smith".
func matchesInitials(playerName, input string) bool {
	parts := strings.Fields(playerName)
	if len(parts) == 0 {
		return false
	}

	// Build initials from name parts
	var initials strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			initials.WriteRune(rune(strings.ToLower(part)[0]))
		}
	}

	return initials.String() == input
}

// ClearFilters clears all player filters.
func (s *StatsViewState) ClearFilters() {
	s.FilteredPlayers = make(map[string]bool)
	s.FilterMode = false
	s.FilterInput = ""
}

// HasFilters returns true if any players are currently filtered.
func (s *StatsViewState) HasFilters() bool {
	return len(s.FilteredPlayers) > 0
}

// IsFiltered returns true if the given player is in the filtered set.
func (s *StatsViewState) IsFiltered(player string) bool {
	return s.FilteredPlayers != nil && s.FilteredPlayers[player]
}

// GetSortedStats returns stats sorted with filtered players at the top.
func (s *StatsViewState) GetSortedStats() []PlayerStats {
	if !s.HasFilters() {
		return s.Stats
	}

	// Separate filtered and non-filtered players
	var filtered, nonFiltered []PlayerStats
	for _, stat := range s.Stats {
		if s.IsFiltered(stat.Player) {
			filtered = append(filtered, stat)
		} else {
			nonFiltered = append(nonFiltered, stat)
		}
	}

	// Return filtered first, then non-filtered
	return append(filtered, nonFiltered...)
}

// StatsView renders the stats view component.
// It displays a table of player tackle statistics.
func StatsView(state StatsViewState, width, height int) string {
	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true).
		Padding(0, 1)

	// Subtitle style
	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true).
		Padding(0, 1)

	// Build content
	var lines []string

	// Title
	title := "Tackle Statistics"
	if state.AllVideos {
		title += " (All Videos)"
	} else {
		title += " (Current Video)"
	}
	lines = append(lines, titleStyle.Render(title))

	// Subtitle with sort indicator
	sortNames := []string{"Player", "Total", "Completed", "Missed", "Possible", "%", "Starred"}
	subtitle := fmt.Sprintf("Sorted by: %s | Tab to change | V to toggle videos | / to filter | Backspace to exit", sortNames[state.SortColumn])
	lines = append(lines, subtitleStyle.Render(subtitle))

	// Filter mode indicator
	if state.FilterMode {
		filterStyle := lipgloss.NewStyle().
			Foreground(styles.Cyan).
			Bold(true).
			Padding(0, 1)
		filterPrompt := fmt.Sprintf("Filter: %s_", state.FilterInput)
		lines = append(lines, filterStyle.Render(filterPrompt))
	} else if state.HasFilters() {
		// Show filter count when not in filter mode
		filterCountStyle := lipgloss.NewStyle().
			Foreground(styles.Pink).
			Italic(true).
			Padding(0, 1)
		lines = append(lines, filterCountStyle.Render(fmt.Sprintf("Filtered: %d player(s) | Esc to clear", len(state.FilteredPlayers))))
	}
	lines = append(lines, "")

	// Empty state
	if len(state.Stats) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Lavender).
			Italic(true).
			Padding(1, 2)
		lines = append(lines, emptyStyle.Render("No tackle data available"))
		return centerContent(strings.Join(lines, "\n"), width, height)
	}

	// Column widths
	colPlayer := 15
	colNum := 6
	colPct := 6
	colTotal := colPlayer + colNum*5 + colPct + colNum + 8 // 8 for separators

	// Header row style
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)

	// Highlight current sort column in header
	headerParts := []string{"Player", "Total", "Comp", "Miss", "Poss", "%", "Star"}
	highlightedHeader := ""
	for i, part := range headerParts {
		var partWidth int
		if i == 0 {
			partWidth = colPlayer
		} else if i == 5 {
			partWidth = colPct
		} else {
			partWidth = colNum
		}

		formatted := fmt.Sprintf("%*s", partWidth, part)
		if i == 0 {
			formatted = fmt.Sprintf("%-*s", partWidth, part)
		}

		if SortColumn(i) == state.SortColumn {
			sortIndicator := lipgloss.NewStyle().
				Foreground(styles.Cyan).
				Bold(true).
				Underline(true)
			highlightedHeader += sortIndicator.Render(formatted)
		} else {
			highlightedHeader += headerStyle.Render(formatted)
		}
		if i < len(headerParts)-1 {
			highlightedHeader += " "
		}
	}
	lines = append(lines, " "+highlightedHeader)

	// Separator
	separator := strings.Repeat("-", colTotal)
	sepStyle := lipgloss.NewStyle().Foreground(styles.Purple)
	lines = append(lines, " "+sepStyle.Render(separator))

	// Calculate visible area for data rows
	visibleHeight := height - len(lines) - 2 // room for padding
	if visibleHeight < 3 {
		visibleHeight = 3
	}

	// Adjust scroll offset to keep selected item visible
	if state.SelectedIndex < state.ScrollOffset {
		state.ScrollOffset = state.SelectedIndex
	} else if state.SelectedIndex >= state.ScrollOffset+visibleHeight {
		state.ScrollOffset = state.SelectedIndex - visibleHeight + 1
	}

	// Get sorted stats (filtered players at top if filters are active)
	displayStats := state.GetSortedStats()

	// Data rows
	for i := state.ScrollOffset; i < len(displayStats) && i < state.ScrollOffset+visibleHeight; i++ {
		stat := displayStats[i]
		isSelected := i == state.SelectedIndex
		isFiltered := state.IsFiltered(stat.Player)
		hasActiveFilters := state.HasFilters()

		// Format percentage
		pctStr := "-"
		if stat.Completed+stat.Missed > 0 {
			pctStr = fmt.Sprintf("%.0f", stat.Percentage)
		}

		row := fmt.Sprintf("%-*s %*d %*d %*d %*d %*s %*d",
			colPlayer, truncateString(stat.Player, colPlayer),
			colNum, stat.Total,
			colNum, stat.Completed,
			colNum, stat.Missed,
			colNum, stat.Possible,
			colPct, pctStr,
			colNum, stat.Starred)

		var rowStyle lipgloss.Style
		if isSelected {
			rowStyle = lipgloss.NewStyle().
				Background(styles.BrightPurple).
				Foreground(styles.LightLavender).
				Bold(true)
		} else if hasActiveFilters && isFiltered {
			// Filtered player - highlighted
			rowStyle = lipgloss.NewStyle().
				Foreground(styles.Cyan).
				Bold(true)
		} else if hasActiveFilters && !isFiltered {
			// Non-filtered player when filters are active - greyed out
			rowStyle = lipgloss.NewStyle().
				Foreground(styles.Purple)
		} else {
			// Normal display (no filters)
			rowStyle = lipgloss.NewStyle().
				Foreground(styles.LightLavender)
		}
		lines = append(lines, " "+rowStyle.Render(row))
	}

	content := strings.Join(lines, "\n")
	return centerContent(content, width, height)
}

// truncateString truncates a string to maxLen characters, adding "..." if needed.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// centerContent centers content within the given dimensions.
func centerContent(content string, width, height int) string {
	contentLines := strings.Split(content, "\n")
	contentHeight := len(contentLines)
	contentWidth := 0
	for _, line := range contentLines {
		w := lipgloss.Width(line)
		if w > contentWidth {
			contentWidth = w
		}
	}

	// Add padding
	paddedWidth := contentWidth + 4
	paddedHeight := contentHeight + 2

	// Center the view
	marginLeft := (width - paddedWidth) / 2
	if marginLeft < 0 {
		marginLeft = 0
	}
	marginTop := (height - paddedHeight) / 2
	if marginTop < 0 {
		marginTop = 0
	}

	// Panel style with border
	panelStyle := lipgloss.NewStyle().
		Background(styles.DarkPurple).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BrightPurple).
		Padding(1, 2)

	// Apply panel style
	panel := panelStyle.Render(content)

	// Create positioning by adding margin
	positionedStyle := lipgloss.NewStyle().
		MarginLeft(marginLeft).
		MarginTop(marginTop)

	return positionedStyle.Render(panel)
}
