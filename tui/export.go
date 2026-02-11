package tui

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/pkg/cliputil"
)

// exportProgressMsg carries progress updates from the export goroutine.
type exportProgressMsg struct {
	current int
	total   int
}

// exportCompleteMsg is sent when export finishes successfully.
type exportCompleteMsg struct {
	count     int
	outputDir string
}

// exportErrorMsg is sent when export encounters an error.
type exportErrorMsg struct {
	err error
}

// waitForExportMsg returns a tea.Cmd that waits for the next message on the channel.
func waitForExportMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// startExportGoroutine validates prerequisites and starts the ffmpeg export in
// a background goroutine. Progress messages are sent to the returned channel.
func startExportGoroutine(database *sql.DB, videoPath string, videoDuration float64) (<-chan tea.Msg, error) {
	// Check ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found. Install: brew install ffmpeg (macOS) or apt install ffmpeg (Linux)")
	}

	// Check video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Video file not found: %s", videoPath)
	}

	// Query tackles
	tackles, err := db.SelectTacklesForExport(database, videoPath)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	if len(tackles) == 0 {
		return nil, fmt.Errorf("No tackles with player data found for this video")
	}

	outputDir := cliputil.GetOutputDir(videoPath)
	total := len(tackles)

	// Use a reasonable default if video duration is unknown
	if videoDuration <= 0 {
		videoDuration = 86400.0
	}

	ch := make(chan tea.Msg)
	go func() {
		defer close(ch)

		exported := 0
		for i, t := range tackles {
			ch <- exportProgressMsg{current: i + 1, total: total}

			player := cliputil.SanitizePlayerName(t.Player)
			timestamp := cliputil.FormatTimestamp(t.Timestamp)

			start, end := cliputil.CalculateClipBounds(t.Timestamp, t.ClipStart, t.ClipEnd, videoDuration)

			playerDir := filepath.Join(outputDir, player)
			if err := os.MkdirAll(playerDir, 0755); err != nil {
				ch <- exportErrorMsg{fmt.Errorf("Failed to create directory: %s - %v", playerDir, err)}
				return
			}

			outputPath := cliputil.GetPlayerClipPath(outputDir, player, timestamp)

			if err := cliputil.ExtractClip(videoPath, start, end, outputPath); err != nil {
				ch <- exportErrorMsg{err}
				return
			}

			exported++
		}

		ch <- exportCompleteMsg{count: exported, outputDir: outputDir}
	}()

	return ch, nil
}
