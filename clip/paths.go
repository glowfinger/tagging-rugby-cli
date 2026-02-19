package clip

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ClipPaths computes the output folder and filename for a clip from note data.
// Folder is derived from the video directory: <videoDir>/clips/<category>/<player>
// Filename format: {HHMMSS}-{player}-{category}-{outcome}-{attempt}.mp4
func ClipPaths(videoPath, category, player string, attempt int, outcome string, startSeconds float64) (folder, filename string) {
	categorySlug := strings.ToLower(strings.ReplaceAll(category, " ", "_"))
	playerSlug := strings.ToLower(strings.ReplaceAll(player, " ", "_"))
	outcomeSlug := strings.ToLower(strings.ReplaceAll(outcome, " ", "_"))

	folder = filepath.Join(filepath.Dir(videoPath), "clips", categorySlug, playerSlug)

	totalSecs := int(startSeconds)
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60
	timestamp := fmt.Sprintf("%02d%02d%02d", hours, minutes, seconds)

	filename = fmt.Sprintf("%s-%s-%s-%s-%d.mp4", timestamp, playerSlug, categorySlug, outcomeSlug, attempt)
	return folder, filename
}
