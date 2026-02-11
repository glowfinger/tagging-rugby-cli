package cliputil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizePlayerName replaces spaces with underscores and removes filesystem-unsafe characters.
func SanitizePlayerName(player string) string {
	if player == "" {
		return "Unknown"
	}
	name := strings.ReplaceAll(player, " ", "_")
	for _, c := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		name = strings.ReplaceAll(name, c, "")
	}
	return name
}

// FormatTimestamp converts seconds to H-MM-SS format (hyphens instead of colons for filenames).
func FormatTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	total := int(seconds)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	return fmt.Sprintf("%d-%02d-%02d", h, m, s)
}

// GetOutputDir returns the clips output directory path based on the video file path.
// For example, "/path/to/rugby-game.mp4" returns "/path/to/rugby-game-clips".
func GetOutputDir(videoPath string) string {
	dir := filepath.Dir(videoPath)
	base := filepath.Base(videoPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, name+"-clips")
}

// GetPlayerClipPath returns the full path for a player's tackle clip.
// Format: {outputDir}/{player}/{player}_{timestamp}_tackle.mp4
func GetPlayerClipPath(outputDir, player, timestamp string) string {
	return filepath.Join(outputDir, player, player+"_"+timestamp+"_tackle.mp4")
}
