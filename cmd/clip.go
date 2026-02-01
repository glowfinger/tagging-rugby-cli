package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/deps"
	"github.com/user/tagging-rugby-cli/mpv"
)

// clipStartState holds the temporary clip start timestamp.
// This is stored in memory and used when 'clip end' is called.
var clipStartState struct {
	mu        sync.Mutex
	timestamp float64
	videoPath string
	isSet     bool
}

var clipCmd = &cobra.Command{
	Use:   "clip",
	Short: "Manage video clips",
	Long:  `Create, list, and manage video clips for analysis.`,
}

var clipStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Mark the start of a clip at current timestamp",
	Long:  `Mark the start point of a new clip at the current video position. Use 'clip end' to complete the clip.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to mpv to get current timestamp and video path
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Get current timestamp
		timestamp, err := client.GetTimePos()
		if err != nil {
			return fmt.Errorf("failed to get current timestamp: %w", err)
		}

		// Get video path from mpv
		videoPathRaw, err := client.GetProperty("path")
		if err != nil {
			return fmt.Errorf("failed to get video path: %w", err)
		}
		videoPath, ok := videoPathRaw.(string)
		if !ok {
			return fmt.Errorf("unexpected video path type: %T", videoPathRaw)
		}

		// Store the start timestamp
		clipStartState.mu.Lock()
		clipStartState.timestamp = timestamp
		clipStartState.videoPath = videoPath
		clipStartState.isSet = true
		clipStartState.mu.Unlock()

		// Format timestamp as MM:SS
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		fmt.Printf("Clip start marked at %d:%02d\n", minutes, seconds)
		fmt.Println("Use 'clip end <name>' to complete the clip.")
		return nil
	},
}

var clipEndCmd = &cobra.Command{
	Use:   "end <name>",
	Short: "Mark the end of a clip and save it",
	Long:  `Mark the end point of a clip at the current video position and save it to the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clipName := args[0]

		// Check if start was marked
		clipStartState.mu.Lock()
		if !clipStartState.isSet {
			clipStartState.mu.Unlock()
			return fmt.Errorf("no clip start marked. Use 'clip start' first")
		}
		startTimestamp := clipStartState.timestamp
		startVideoPath := clipStartState.videoPath
		clipStartState.isSet = false // Clear the state
		clipStartState.mu.Unlock()

		// Connect to mpv to get current timestamp and video path
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Get current timestamp (end point)
		endTimestamp, err := client.GetTimePos()
		if err != nil {
			return fmt.Errorf("failed to get current timestamp: %w", err)
		}

		// Get video path from mpv to verify it's the same video
		videoPathRaw, err := client.GetProperty("path")
		if err != nil {
			return fmt.Errorf("failed to get video path: %w", err)
		}
		videoPath, ok := videoPathRaw.(string)
		if !ok {
			return fmt.Errorf("unexpected video path type: %T", videoPathRaw)
		}

		// Verify same video
		if videoPath != startVideoPath {
			return fmt.Errorf("video changed since clip start was marked")
		}

		// Validate start < end
		if startTimestamp >= endTimestamp {
			return fmt.Errorf("clip end time (%d:%02d) must be after start time (%d:%02d)",
				int(endTimestamp)/60, int(endTimestamp)%60,
				int(startTimestamp)/60, int(startTimestamp)%60)
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		duration := endTimestamp - startTimestamp
		now := time.Now()

		// Insert note with clip and timing child rows
		children := db.NoteChildren{
			Clips: []db.NoteClip{
				{Name: clipName, Duration: duration, StartedAt: &now, FinishedAt: &now},
			},
			Timings: []db.NoteTiming{
				{Start: startTimestamp, End: endTimestamp},
			},
			Videos: []db.NoteVideo{
				{Path: videoPath, StoppedAt: startTimestamp},
			},
		}

		noteID, err := db.InsertNoteWithChildren(database, "clip", children)
		if err != nil {
			return fmt.Errorf("failed to insert clip: %w", err)
		}

		// Format timestamps
		startMin := int(startTimestamp) / 60
		startSec := int(startTimestamp) % 60
		endMin := int(endTimestamp) / 60
		endSec := int(endTimestamp) % 60

		fmt.Printf("Clip saved: Note ID %d (%d:%02d - %d:%02d, %.1fs)\n", noteID, startMin, startSec, endMin, endSec, duration)
		return nil
	},
}

var clipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clips for the current video",
	Long:  `Display all clips for the current video as a table, sorted by start time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to mpv to get current video path
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Get video path from mpv
		videoPathRaw, err := client.GetProperty("path")
		if err != nil {
			return fmt.Errorf("failed to get video path: %w", err)
		}
		videoPath, ok := videoPathRaw.(string)
		if !ok {
			return fmt.Errorf("unexpected video path type: %T", videoPathRaw)
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Query clips joined with timing and video tables
		rows, err := database.Query(
			`SELECT n.id, nc.name, nc.duration, COALESCE(nt.start, 0), COALESCE(nt.end, 0)
			 FROM notes n
			 INNER JOIN note_clips nc ON nc.note_id = n.id
			 INNER JOIN note_videos nv ON nv.note_id = n.id
			 LEFT JOIN note_timing nt ON nt.note_id = n.id
			 WHERE nv.path = ?
			 ORDER BY nt.start ASC`, videoPath)
		if err != nil {
			return fmt.Errorf("failed to query clips: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NoteID\tName\tStart\tEnd\tDuration")
		fmt.Fprintln(w, "------\t----\t-----\t---\t--------")

		count := 0
		for rows.Next() {
			var noteID int64
			var name string
			var duration, startSec, endSec float64

			if err := rows.Scan(&noteID, &name, &duration, &startSec, &endSec); err != nil {
				return fmt.Errorf("failed to scan clip: %w", err)
			}

			// Format timestamps
			startMin := int(startSec) / 60
			startSecInt := int(startSec) % 60
			endMin := int(endSec) / 60
			endSecInt := int(endSec) % 60

			startStr := fmt.Sprintf("%d:%02d", startMin, startSecInt)
			endStr := fmt.Sprintf("%d:%02d", endMin, endSecInt)
			durationStr := fmt.Sprintf("%.1fs", duration)

			// Truncate name if too long
			if len(name) > 40 {
				name = name[:37] + "..."
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", noteID, name, startStr, endStr, durationStr)
			count++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating clips: %w", err)
		}

		w.Flush()

		if count == 0 {
			fmt.Println("\nNo clips found for this video.")
		} else {
			fmt.Printf("\n%d clip(s) found.\n", count)
		}

		return nil
	},
}

var clipPlayCmd = &cobra.Command{
	Use:   "play <note-id>",
	Short: "Play a clip with A-B loop",
	Long:  `Seek to a clip's start timestamp and set mpv A-B loop to loop the clip continuously.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var noteID int64
		if _, err := fmt.Sscanf(args[0], "%d", &noteID); err != nil {
			return fmt.Errorf("invalid note ID: %s", args[0])
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get timing for this note
		timings, err := db.SelectNoteTimingByNote(database, noteID)
		if err != nil {
			return fmt.Errorf("failed to query timing: %w", err)
		}
		if len(timings) == 0 {
			return fmt.Errorf("no timing found for note ID %d", noteID)
		}

		startSec := timings[0].Start
		endSec := timings[0].End

		// Get clip name
		clips, err := db.SelectNoteClipsByNote(database, noteID)
		if err != nil {
			return fmt.Errorf("failed to query clip: %w", err)
		}
		clipName := "(no name)"
		if len(clips) > 0 && clips[0].Name != "" {
			clipName = clips[0].Name
		}

		// Connect to mpv
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Seek to clip start
		if err := client.Seek(startSec); err != nil {
			return fmt.Errorf("failed to seek to clip start: %w", err)
		}

		// Set A-B loop
		if err := client.SetABLoop(startSec, endSec); err != nil {
			return fmt.Errorf("failed to set A-B loop: %w", err)
		}

		// Format timestamps
		startMin := int(startSec) / 60
		startSecInt := int(startSec) % 60
		endMin := int(endSec) / 60
		endSecInt := int(endSec) % 60
		duration := endSec - startSec

		fmt.Printf("Playing clip (note %d): %s\n", noteID, clipName)
		fmt.Printf("Looping %d:%02d - %d:%02d (%.1fs)\n", startMin, startSecInt, endMin, endSecInt, duration)
		fmt.Println("Use 'clip stop' to clear the loop.")
		return nil
	},
}

var clipStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop clip loop playback",
	Long:  `Clear the A-B loop to stop looping the current clip.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Connect to mpv
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Clear A-B loop
		if err := client.ClearABLoop(); err != nil {
			return fmt.Errorf("failed to clear A-B loop: %w", err)
		}

		fmt.Println("A-B loop cleared.")
		return nil
	},
}

var clipExportCmd = &cobra.Command{
	Use:   "export <note-id>",
	Short: "Export a clip as a video file using ffmpeg",
	Long:  `Export a clip as a video file using ffmpeg. By default uses stream copy (-c copy) for fast export.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check ffmpeg is installed
		if err := deps.CheckFfmpeg(); err != nil {
			return err
		}

		var noteID int64
		if _, err := fmt.Sscanf(args[0], "%d", &noteID); err != nil {
			return fmt.Errorf("invalid note ID: %s", args[0])
		}

		// Get flags
		outputPath, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		reencode, _ := cmd.Flags().GetBool("reencode")

		// Validate format
		validFormats := map[string]bool{"mp4": true, "webm": true, "mkv": true}
		if !validFormats[format] {
			return fmt.Errorf("invalid format: %s (supported: mp4, webm, mkv)", format)
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get video path
		videos, err := db.SelectNoteVideosByNote(database, noteID)
		if err != nil || len(videos) == 0 {
			return fmt.Errorf("no video found for note ID %d", noteID)
		}
		videoPath := videos[0].Path

		// Get timing
		timings, err := db.SelectNoteTimingByNote(database, noteID)
		if err != nil || len(timings) == 0 {
			return fmt.Errorf("no timing found for note ID %d", noteID)
		}
		startSec := timings[0].Start
		endSec := timings[0].End

		// Determine output path
		if outputPath == "" {
			outputPath = fmt.Sprintf("clip-%d.%s", noteID, format)
		}

		// Build ffmpeg command
		ffmpegArgs := buildFfmpegArgs(videoPath, startSec, endSec, outputPath, format, reencode)

		fmt.Printf("Exporting clip (note %d) to %s...\n", noteID, outputPath)

		// Run ffmpeg
		ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)
		ffmpegCmd.Stdout = os.Stdout
		ffmpegCmd.Stderr = os.Stderr

		if err := ffmpegCmd.Run(); err != nil {
			return fmt.Errorf("ffmpeg export failed: %w", err)
		}

		// Get file size
		fileInfo, err := os.Stat(outputPath)
		if err == nil {
			fmt.Printf("Exported clip (note %d) to %s (%.2f MB)\n", noteID, outputPath, float64(fileInfo.Size())/(1024*1024))
		} else {
			fmt.Printf("Exported clip (note %d) to %s\n", noteID, outputPath)
		}

		return nil
	},
}

// buildFfmpegArgs builds the ffmpeg command arguments
func buildFfmpegArgs(videoPath string, startSec, endSec float64, outputPath, format string, reencode bool) []string {
	args := []string{
		"-y",                                 // Overwrite output
		"-ss", fmt.Sprintf("%.3f", startSec), // Start time (input seeking for faster seek)
		"-i", videoPath, // Input file
		"-to", fmt.Sprintf("%.3f", endSec-startSec), // Duration (relative to start)
	}

	if reencode {
		// Re-encode with appropriate codec for format
		switch format {
		case "webm":
			args = append(args, "-c:v", "libvpx-vp9", "-c:a", "libopus")
		case "mkv", "mp4":
			args = append(args, "-c:v", "libx264", "-c:a", "aac")
		}
	} else {
		// Stream copy (fast, no re-encoding)
		args = append(args, "-c", "copy")
	}

	// Add output file with proper extension
	if filepath.Ext(outputPath) == "" {
		outputPath = outputPath + "." + format
	}
	args = append(args, outputPath)

	return args
}

func init() {
	// Add flags to clip export command
	clipExportCmd.Flags().StringP("output", "o", "", "Custom output file path")
	clipExportCmd.Flags().StringP("format", "f", "mp4", "Output format (mp4, webm, mkv)")
	clipExportCmd.Flags().Bool("reencode", false, "Re-encode video instead of stream copy")

	// Build command tree
	clipCmd.AddCommand(clipStartCmd)
	clipCmd.AddCommand(clipEndCmd)
	clipCmd.AddCommand(clipListCmd)
	clipCmd.AddCommand(clipPlayCmd)
	clipCmd.AddCommand(clipStopCmd)
	clipCmd.AddCommand(clipExportCmd)
	rootCmd.AddCommand(clipCmd)
}
