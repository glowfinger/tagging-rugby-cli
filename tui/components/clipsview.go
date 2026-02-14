// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/tui/styles"
)

// ExportedClip represents a single exported clip for display.
type ExportedClip struct {
	NoteID   int64
	FileName string
	Player   string
	Category string
	Outcome  string
	Duration float64
	Status   string // "completed" or "error"
	Error    string
}

// ClipsViewState holds the state for the clips list view overlay.
type ClipsViewState struct {
	// Active indicates if the clips view is currently displayed
	Active bool
	// Clips is the list of exported clips
	Clips []ExportedClip
	// ScrollOffset is the scroll position
	ScrollOffset int
}

// MoveUp moves the scroll position up in the clips view.
func (s *ClipsViewState) MoveUp() {
	if s.ScrollOffset > 0 {
		s.ScrollOffset--
	}
}

// MoveDown moves the scroll position down.
func (s *ClipsViewState) MoveDown(visibleHeight int, totalLines int) {
	if s.ScrollOffset < totalLines-visibleHeight {
		s.ScrollOffset++
	}
}

// ClipsView renders the clips list view overlay.
func ClipsView(state ClipsViewState, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true).
		Padding(0, 1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)

	completedStyle := lipgloss.NewStyle().
		Foreground(styles.Green)

	errorStyle := lipgloss.NewStyle().
		Foreground(styles.Red)

	fileStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	detailStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	var lines []string

	// Title
	lines = append(lines, titleStyle.Render("Exported Clips"))

	// Subtitle with instructions
	subtitle := "J/K to scroll | Backspace/Esc to exit"
	lines = append(lines, subtitleStyle.Render(subtitle))

	// Summary
	completed := 0
	errors := 0
	for _, clip := range state.Clips {
		if clip.Status == "completed" {
			completed++
		} else if clip.Status == "error" {
			errors++
		}
	}
	summary := fmt.Sprintf("%d clips exported", completed)
	if errors > 0 {
		summary += fmt.Sprintf(", %d errors", errors)
	}
	lines = append(lines, subtitleStyle.Render(summary))
	lines = append(lines, "")

	if len(state.Clips) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Lavender).
			Italic(true).
			Padding(1, 2)
		lines = append(lines, emptyStyle.Render("No exported clips yet. Press Ctrl+E to export."))
		return centerContent(strings.Join(lines, "\n"), width, height)
	}

	// Group clips by video path
	type videoGroup struct {
		videoName string
		clips     []ExportedClip
	}
	var groups []videoGroup
	groupMap := make(map[string]int)

	for _, clip := range state.Clips {
		videoName := filepath.Base(clip.FileName)
		// Extract video name from clip path: clips/{videoName}/...
		parts := strings.Split(clip.FileName, string(filepath.Separator))
		for i, part := range parts {
			if part == "clips" && i+1 < len(parts) {
				videoName = parts[i+1]
				break
			}
		}

		if idx, ok := groupMap[videoName]; ok {
			groups[idx].clips = append(groups[idx].clips, clip)
		} else {
			groupMap[videoName] = len(groups)
			groups = append(groups, videoGroup{videoName: videoName, clips: []ExportedClip{clip}})
		}
	}

	// Build content lines for each group
	var contentLines []string
	for _, group := range groups {
		// Video header
		contentLines = append(contentLines, " "+headerStyle.Render("── "+group.videoName+" ──"))

		for _, clip := range group.clips {
			// Status indicator
			var statusStr string
			if clip.Status == "completed" {
				statusStr = completedStyle.Render("✓")
			} else {
				statusStr = errorStyle.Render("✗")
			}

			// Clip info line
			playerOutcome := clip.Player
			if clip.Outcome != "" {
				playerOutcome += " / " + clip.Outcome
			}
			durationStr := ""
			if clip.Duration > 0 {
				durationStr = fmt.Sprintf(" (%.1fs)", clip.Duration)
			}
			line := fmt.Sprintf("  %s %s%s",
				statusStr,
				fileStyle.Render(playerOutcome),
				detailStyle.Render(durationStr))
			contentLines = append(contentLines, line)

			// Show file name on next line
			contentLines = append(contentLines, "    "+detailStyle.Render(filepath.Base(clip.FileName)))

			// Show error on next line if present
			if clip.Error != "" {
				contentLines = append(contentLines, "    "+errorStyle.Render("Error: "+clip.Error))
			}
		}
		contentLines = append(contentLines, "")
	}

	// Apply scrolling
	visibleHeight := height - len(lines) - 6 // room for panel padding/border
	if visibleHeight < 3 {
		visibleHeight = 3
	}

	// Clamp scroll offset
	if state.ScrollOffset > len(contentLines)-visibleHeight {
		state.ScrollOffset = len(contentLines) - visibleHeight
	}
	if state.ScrollOffset < 0 {
		state.ScrollOffset = 0
	}

	endIdx := state.ScrollOffset + visibleHeight
	if endIdx > len(contentLines) {
		endIdx = len(contentLines)
	}

	lines = append(lines, contentLines[state.ScrollOffset:endIdx]...)

	content := strings.Join(lines, "\n")
	return centerContent(content, width, height)
}
