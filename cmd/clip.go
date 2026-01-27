package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/tabwriter"

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
		fmt.Println("Use 'clip end <description>' to complete the clip.")
		return nil
	},
}

var clipEndCmd = &cobra.Command{
	Use:   "end <description>",
	Short: "Mark the end of a clip and save it",
	Long:  `Mark the end point of a clip at the current video position and save it to the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := args[0]

		// Get flags
		category, _ := cmd.Flags().GetString("category")
		player, _ := cmd.Flags().GetString("player")
		team, _ := cmd.Flags().GetString("team")

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

		// Insert clip
		result, err := database.Exec(
			`INSERT INTO clips (video_path, start_seconds, end_seconds, description, category, player, team) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			videoPath, startTimestamp, endTimestamp, description, category, player, team,
		)
		if err != nil {
			return fmt.Errorf("failed to insert clip: %w", err)
		}

		clipID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get clip ID: %w", err)
		}

		// Format timestamps
		startMin := int(startTimestamp) / 60
		startSec := int(startTimestamp) % 60
		endMin := int(endTimestamp) / 60
		endSec := int(endTimestamp) % 60
		duration := endTimestamp - startTimestamp

		fmt.Printf("Clip saved: ID %d (%d:%02d - %d:%02d, %.1fs)\n", clipID, startMin, startSec, endMin, endSec, duration)
		return nil
	},
}

var clipAddCmd = &cobra.Command{
	Use:   "add <start> <end> <description>",
	Short: "Add a clip with specified start and end times",
	Long:  `Add a video clip with manually specified start and end times. Times can be in MM:SS or seconds format.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse start time
		startTimestamp, err := parseTimeToSeconds(args[0])
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}

		// Parse end time
		endTimestamp, err := parseTimeToSeconds(args[1])
		if err != nil {
			return fmt.Errorf("invalid end time: %w", err)
		}

		description := args[2]

		// Get flags
		category, _ := cmd.Flags().GetString("category")
		player, _ := cmd.Flags().GetString("player")
		team, _ := cmd.Flags().GetString("team")

		// Validate start < end
		if startTimestamp >= endTimestamp {
			return fmt.Errorf("clip end time must be after start time")
		}

		// Connect to mpv to get video path
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

		// Insert clip
		result, err := database.Exec(
			`INSERT INTO clips (video_path, start_seconds, end_seconds, description, category, player, team) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			videoPath, startTimestamp, endTimestamp, description, category, player, team,
		)
		if err != nil {
			return fmt.Errorf("failed to insert clip: %w", err)
		}

		clipID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get clip ID: %w", err)
		}

		// Format timestamps
		startMin := int(startTimestamp) / 60
		startSec := int(startTimestamp) % 60
		endMin := int(endTimestamp) / 60
		endSec := int(endTimestamp) % 60
		duration := endTimestamp - startTimestamp

		fmt.Printf("Clip added: ID %d (%d:%02d - %d:%02d, %.1fs)\n", clipID, startMin, startSec, endMin, endSec, duration)
		return nil
	},
}

var clipPlayCmd = &cobra.Command{
	Use:   "play <id>",
	Short: "Play a clip with A-B loop",
	Long:  `Seek to a clip's start timestamp and set mpv A-B loop to loop the clip continuously.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var clipID int64
		if _, err := fmt.Sscanf(args[0], "%d", &clipID); err != nil {
			return fmt.Errorf("invalid clip ID: %s", args[0])
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Query clip by ID
		var startSec, endSec float64
		var description sql.NullString
		err = database.QueryRow(
			`SELECT start_seconds, end_seconds, description FROM clips WHERE id = ?`,
			clipID,
		).Scan(&startSec, &endSec, &description)
		if err == sql.ErrNoRows {
			return fmt.Errorf("clip not found: ID %d", clipID)
		}
		if err != nil {
			return fmt.Errorf("failed to query clip: %w", err)
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

		descStr := nullStringValue(description)
		if descStr == "" {
			descStr = "(no description)"
		}

		fmt.Printf("Playing clip %d: %s\n", clipID, descStr)
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
	Use:   "export [clip-id]",
	Short: "Export clips as video files using ffmpeg",
	Long:  `Export a clip or all clips as video files using ffmpeg. By default uses stream copy (-c copy) for fast export.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check ffmpeg is installed
		if err := deps.CheckFfmpeg(); err != nil {
			return err
		}

		// Get flags
		outputPath, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		reencode, _ := cmd.Flags().GetBool("reencode")
		exportAll, _ := cmd.Flags().GetBool("all")

		// Validate format
		validFormats := map[string]bool{"mp4": true, "webm": true, "mkv": true}
		if !validFormats[format] {
			return fmt.Errorf("invalid format: %s (supported: mp4, webm, mkv)", format)
		}

		// If --all flag is set, export all clips for current video
		if exportAll {
			return exportAllClips(format, reencode)
		}

		// Otherwise, require clip ID
		if len(args) == 0 {
			return fmt.Errorf("clip ID required (or use --all to export all clips)")
		}

		var clipID int64
		if _, err := fmt.Sscanf(args[0], "%d", &clipID); err != nil {
			return fmt.Errorf("invalid clip ID: %s", args[0])
		}

		return exportClip(clipID, outputPath, format, reencode)
	},
}

// exportClip exports a single clip to a video file
func exportClip(clipID int64, outputPath, format string, reencode bool) error {
	// Open database
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Query clip by ID
	var videoPath string
	var startSec, endSec float64
	err = database.QueryRow(
		`SELECT video_path, start_seconds, end_seconds FROM clips WHERE id = ?`,
		clipID,
	).Scan(&videoPath, &startSec, &endSec)
	if err == sql.ErrNoRows {
		return fmt.Errorf("clip not found: ID %d", clipID)
	}
	if err != nil {
		return fmt.Errorf("failed to query clip: %w", err)
	}

	// Determine output path
	if outputPath == "" {
		outputPath = fmt.Sprintf("clip-%d.%s", clipID, format)
	}

	// Build ffmpeg command
	ffmpegArgs := buildFfmpegArgs(videoPath, startSec, endSec, outputPath, format, reencode)

	fmt.Printf("Exporting clip %d to %s...\n", clipID, outputPath)

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
		fmt.Printf("Exported clip %d to %s (%.2f MB)\n", clipID, outputPath, float64(fileInfo.Size())/(1024*1024))
	} else {
		fmt.Printf("Exported clip %d to %s\n", clipID, outputPath)
	}

	return nil
}

// exportAllClips exports all clips for the current video
func exportAllClips(format string, reencode bool) error {
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

	// Query all clips for this video
	rows, err := database.Query(
		`SELECT id, start_seconds, end_seconds FROM clips WHERE video_path = ? ORDER BY id ASC`,
		videoPath,
	)
	if err != nil {
		return fmt.Errorf("failed to query clips: %w", err)
	}
	defer rows.Close()

	var clips []struct {
		id       int64
		startSec float64
		endSec   float64
	}

	for rows.Next() {
		var clip struct {
			id       int64
			startSec float64
			endSec   float64
		}
		if err := rows.Scan(&clip.id, &clip.startSec, &clip.endSec); err != nil {
			return fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating clips: %w", err)
	}

	if len(clips) == 0 {
		fmt.Println("No clips found for current video.")
		return nil
	}

	fmt.Printf("Exporting %d clip(s)...\n\n", len(clips))

	successCount := 0
	for _, clip := range clips {
		outputPath := fmt.Sprintf("clip-%d.%s", clip.id, format)

		// Build ffmpeg command
		ffmpegArgs := buildFfmpegArgs(videoPath, clip.startSec, clip.endSec, outputPath, format, reencode)

		fmt.Printf("Exporting clip %d to %s...\n", clip.id, outputPath)

		// Run ffmpeg
		ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)
		ffmpegCmd.Stdout = os.Stdout
		ffmpegCmd.Stderr = os.Stderr

		if err := ffmpegCmd.Run(); err != nil {
			fmt.Printf("Failed to export clip %d: %v\n", clip.id, err)
			continue
		}

		successCount++
	}

	fmt.Printf("\nExported %d/%d clips successfully.\n", successCount, len(clips))
	return nil
}

// buildFfmpegArgs builds the ffmpeg command arguments
func buildFfmpegArgs(videoPath string, startSec, endSec float64, outputPath, format string, reencode bool) []string {
	args := []string{
		"-y",                               // Overwrite output
		"-ss", fmt.Sprintf("%.3f", startSec), // Start time (input seeking for faster seek)
		"-i", videoPath,                    // Input file
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

		// Query clips
		rows, err := database.Query(
			`SELECT id, start_seconds, end_seconds, category, description FROM clips WHERE video_path = ? ORDER BY start_seconds ASC`,
			videoPath,
		)
		if err != nil {
			return fmt.Errorf("failed to query clips: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tStart\tEnd\tDuration\tCategory\tDescription")
		fmt.Fprintln(w, "--\t-----\t---\t--------\t--------\t-----------")

		count := 0
		for rows.Next() {
			var id int64
			var startSec, endSec float64
			var category, description sql.NullString

			if err := rows.Scan(&id, &startSec, &endSec, &category, &description); err != nil {
				return fmt.Errorf("failed to scan clip: %w", err)
			}

			// Format timestamps
			startMin := int(startSec) / 60
			startSecInt := int(startSec) % 60
			endMin := int(endSec) / 60
			endSecInt := int(endSec) % 60
			duration := endSec - startSec

			startStr := fmt.Sprintf("%d:%02d", startMin, startSecInt)
			endStr := fmt.Sprintf("%d:%02d", endMin, endSecInt)
			durationStr := fmt.Sprintf("%.1fs", duration)

			// Handle NULL values
			catStr := nullStringValue(category)
			descStr := nullStringValue(description)

			// Truncate description if too long
			if len(descStr) > 40 {
				descStr = descStr[:37] + "..."
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", id, startStr, endStr, durationStr, catStr, descStr)
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

func init() {
	// Add flags to clip end command
	clipEndCmd.Flags().StringP("category", "c", "", "Clip category (e.g., try, tackle, turnover)")
	clipEndCmd.Flags().StringP("player", "p", "", "Player name or number")
	clipEndCmd.Flags().StringP("team", "t", "", "Team name")

	// Add flags to clip add command
	clipAddCmd.Flags().StringP("category", "c", "", "Clip category (e.g., try, tackle, turnover)")
	clipAddCmd.Flags().StringP("player", "p", "", "Player name or number")
	clipAddCmd.Flags().StringP("team", "t", "", "Team name")

	// Add flags to clip export command
	clipExportCmd.Flags().StringP("output", "o", "", "Custom output file path")
	clipExportCmd.Flags().StringP("format", "f", "mp4", "Output format (mp4, webm, mkv)")
	clipExportCmd.Flags().Bool("reencode", false, "Re-encode video instead of stream copy")
	clipExportCmd.Flags().Bool("all", false, "Export all clips for current video")

	// Build command tree
	clipCmd.AddCommand(clipStartCmd)
	clipCmd.AddCommand(clipEndCmd)
	clipCmd.AddCommand(clipAddCmd)
	clipCmd.AddCommand(clipListCmd)
	clipCmd.AddCommand(clipPlayCmd)
	clipCmd.AddCommand(clipStopCmd)
	clipCmd.AddCommand(clipExportCmd)
	rootCmd.AddCommand(clipCmd)
}
