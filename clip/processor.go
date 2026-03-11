package clip

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/tagging-rugby-cli/db"
)

// Processor manages the background clip generation worker.
type Processor struct {
	DB *sql.DB
}

// Start launches a goroutine that continuously polls for pending clips and processes them.
// The goroutine exits when ctx is cancelled.
func (p *Processor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			clip, err := db.SelectNextPendingClip(p.DB)
			if err != nil {
				// On DB error, wait and retry
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
				continue
			}
			if clip == nil {
				// No pending clips; sleep and retry
				select {
				case <-ctx.Done():
					return
				case <-time.After(2 * time.Second):
				}
				continue
			}

			p.processClip(ctx, clip)
		}
	}()
}

// processClip handles the full lifecycle of generating a single clip.
func (p *Processor) processClip(ctx context.Context, c *db.PendingClip) {
	// Check ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		_ = db.MarkClipError(p.DB, c.ClipID, time.Now(), "ffmpeg not found in PATH")
		return
	}

	if err := db.MarkClipProcessing(p.DB, c.ClipID, time.Now()); err != nil {
		return
	}

	// Create output directory
	outDir := c.Folder
	if err := os.MkdirAll(outDir, 0755); err != nil {
		_ = db.MarkClipError(p.DB, c.ClipID, time.Now(), fmt.Sprintf("mkdir: %v", err))
		return
	}

	outPath := filepath.Join(outDir, c.Filename)

	// Compute clip duration
	duration := c.End - c.Start
	if duration < 4.0 {
		duration = 4.0
	}

	// ffmpeg filtergraph has two escaping levels:
	//   Level 2 (filtergraph): sees '\\:' and converts '\\' → '\', leaving ':' bare.
	//   Level 1 (option):      sees '\:' and treats the ':' as a literal (not a separator).
	// So colons in option values need '\\:' in the runtime string, which requires
	// '\\\\:' (4 backslashes + colon) in the Go source.
	totalSecs := int(c.Start)
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60
	startTimestamp := fmt.Sprintf("%02d\\\\:%02d\\\\:%02d", hours, minutes, seconds)

	endTotalSecs := int(c.End)
	endHours := endTotalSecs / 3600
	endMinutes := (endTotalSecs % 3600) / 60
	endSeconds := endTotalSecs % 60
	endTimestamp := fmt.Sprintf("%02d\\\\:%02d\\\\:%02d", endHours, endMinutes, endSeconds)

	// Escape colons in free-text fields using the same double-escape.
	outcome := strings.ReplaceAll(c.Outcome, ":", "\\\\:")
	player := strings.ReplaceAll(c.Player, ":", "\\\\:")
	filename := strings.ReplaceAll(c.Filename, ":", "\\\\:")

	// Build drawtext filter chain. No single quotes used — only backslash escaping.
	// The comma in lt(t,2) is escaped as '\,' so it is not treated as a filter
	// separator at the filtergraph level.
	drawtext := fmt.Sprintf(
		"drawtext=text=Note ID\\\\: %d:x=10:y=h-th:fontsize=28:fontcolor=white:bordercolor=#131211:borderw=2:enable=lt(t\\,2),"+
			"drawtext=text=Filename\\\\: %s:x=10:y=h-th-36:fontsize=28:fontcolor=white:bordercolor=#131211:borderw=2:enable=lt(t\\,2),"+
			"drawtext=text=Outcome\\\\: %s %d:x=10:y=h-th-72:fontsize=28:fontcolor=white:bordercolor=#131211:borderw=2:enable=lt(t\\,2),"+
			"drawtext=text=Player\\\\: %s:x=10:y=h-th-108:fontsize=28:fontcolor=white:bordercolor=#131211:borderw=2:enable=lt(t\\,2),"+
			"drawtext=text=Time\\\\: %s/%s:x=10:y=h-th-144:fontsize=28:fontcolor=white:bordercolor=#131211:borderw=2:enable=lt(t\\,2)",
		c.NoteID,
		filename,
		outcome,
		c.Attempt,
		player,
		startTimestamp,
		endTimestamp,
	)

	args := []string{
		"-y",
		"-i", c.VideoPath,
		"-ss", fmt.Sprintf("%f", c.Start),
		"-t", fmt.Sprintf("%f", duration),
		"-vf", drawtext,
		outPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	runErr := cmd.Run()
	if runErr != nil {
		_ = db.MarkClipError(p.DB, c.ClipID, time.Now(), out.String())
		return
	}

	// Stat the output file for filesize
	info, err := os.Stat(outPath)
	if err != nil {
		_ = db.MarkClipError(p.DB, c.ClipID, time.Now(), fmt.Sprintf("stat output: %v", err))
		return
	}

	_ = db.MarkClipComplete(p.DB, c.ClipID, time.Now(), info.Size())
}
