package export

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/user/tagging-rugby-cli/deps"
)

// unsafeChars matches characters not safe for filenames: / \ : * ? < > | and spaces
var unsafeChars = regexp.MustCompile(`[/\\:*?<>|\s]`)

// sanitize replaces unsafe filename characters with underscores.
func sanitize(s string) string {
	return unsafeChars.ReplaceAllString(s, "_")
}

// BuildClipPath returns the full output path for an exported clip.
// Format: {videoDir}/clips/{videoFilenameNoExt}/{category}/{player}/{hhmmss}-{player}-{category}-{outcome}.mp4
func BuildClipPath(videoPath, category, player, outcome string, startSeconds float64) string {
	videoDir := filepath.Dir(videoPath)
	videoBase := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	safeCat := sanitize(category)
	safePlayer := sanitize(player)
	safeOutcome := sanitize(outcome)

	total := int(math.Floor(startSeconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	hhmmss := fmt.Sprintf("%02d%02d%02d", h, m, s)

	filename := fmt.Sprintf("%s-%s-%s-%s.mp4", hhmmss, safePlayer, safeCat, safeOutcome)

	return filepath.Join(videoDir, "clips", videoBase, safeCat, safePlayer, filename)
}

// EffectiveEnd returns the effective end time, enforcing a 4-second minimum duration.
// If end <= 0 or end < start+4, returns start+4.
func EffectiveEnd(start, end float64) float64 {
	if end <= 0 {
		return start + 4.0
	}
	return math.Max(end, start+4.0)
}

// RunFfmpeg creates the output directory, checks for ffmpeg, and runs the ffmpeg
// command to extract a clip using stream copy.
func RunFfmpeg(videoPath string, start, end float64, outputPath string) error {
	if err := deps.CheckFfmpeg(); err != nil {
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	duration := end - start
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", start),
		"-i", videoPath,
		"-to", fmt.Sprintf("%.3f", duration),
		"-c", "copy",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\n%s", err, string(output))
	}

	return nil
}
