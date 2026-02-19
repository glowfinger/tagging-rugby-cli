package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/mpv"
	"github.com/user/tagging-rugby-cli/pkg/timeutil"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage timestamped notes",
	Long:  `Add, list, and delete timestamped notes for video analysis.`,
}

var noteAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a note at the current timestamp",
	Long:  `Add a timestamped note at the current video position. Creates a note with timing and video child records.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		category, _ := cmd.Flags().GetString("category")
		text, _ := cmd.Flags().GetString("text")

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

		// Get video duration
		duration, _ := client.GetDuration()

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Get video file metadata
		var videoSize int64
		if info, err := os.Stat(videoPath); err == nil {
			videoSize = info.Size()
		}
		videoFormat := strings.TrimPrefix(filepath.Ext(videoPath), ".")

		// Build children
		children := db.NoteChildren{
			Timings: []db.NoteTiming{
				{Start: timestamp, End: timestamp},
			},
			Videos: []db.NoteVideo{
				{Path: videoPath, Duration: duration, StoppedAt: timestamp, Size: videoSize, Format: videoFormat},
			},
		}

		// Add detail if text was provided
		if text != "" {
			children.Details = []db.NoteDetail{
				{Type: "text", Note: text},
			}
		}

		// Insert note with children
		noteID, err := db.InsertNoteWithChildren(database, category, children)
		if err != nil {
			return fmt.Errorf("failed to insert note: %w", err)
		}

		fmt.Printf("Note added: ID %d at %s\n", noteID, timeutil.FormatTime(timestamp))
		return nil
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes for the current video",
	Long:  `Display all notes for the current video as a table, sorted by timestamp.`,
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

		// Query notes with video join to filter by current video, plus timing
		rows, err := database.Query(
			`SELECT n.id, n.category, COALESCE(nt.start, 0) as start_time
			 FROM notes n
			 INNER JOIN videos v ON v.id = n.video_id
			 LEFT JOIN note_timing nt ON nt.note_id = n.id
			 WHERE v.path = ?
			 ORDER BY start_time ASC`, videoPath)
		if err != nil {
			return fmt.Errorf("failed to query notes: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTime\tCategory")
		fmt.Fprintln(w, "--\t----\t--------")

		count := 0
		for rows.Next() {
			var id int64
			var category sql.NullString
			var startTime float64

			if err := rows.Scan(&id, &category, &startTime); err != nil {
				return fmt.Errorf("failed to scan note: %w", err)
			}

			timeStr := timeutil.FormatTime(startTime)

			catStr := nullStringValue(category)

			fmt.Fprintf(w, "%d\t%s\t%s\n", id, timeStr, catStr)
			count++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating notes: %w", err)
		}

		w.Flush()

		if count == 0 {
			fmt.Println("\nNo matching notes found.")
		} else {
			fmt.Printf("\n%d note(s) found.\n", count)
		}

		return nil
	},
}

var noteGotoCmd = &cobra.Command{
	Use:   "goto <id>",
	Short: "Jump to a note's timestamp",
	Long:  `Seek mpv to the timestamp of an existing note by ID.`,
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

		// Fetch the note
		note, err := db.SelectNoteByID(database, noteID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("note with ID %d not found", noteID)
		} else if err != nil {
			return fmt.Errorf("failed to fetch note: %w", err)
		}

		// Get timing for seek position
		timings, err := db.SelectNoteTimingByNote(database, noteID)
		if err != nil {
			return fmt.Errorf("failed to fetch note timing: %w", err)
		}

		var seekPos float64
		if len(timings) > 0 {
			seekPos = timings[0].Start
		}

		// Connect to mpv
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Seek to the note's timestamp
		if err := client.Seek(seekPos); err != nil {
			return fmt.Errorf("failed to seek to timestamp: %w", err)
		}

		fmt.Printf("Jumped to note %d at %s\n", noteID, timeutil.FormatTime(seekPos))
		if note.Category != "" {
			fmt.Printf("  Category: %s\n", note.Category)
		}

		return nil
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a note",
	Long:  `Delete an existing note by ID. Cascade deletes all child records. Prompts for confirmation unless --force is used.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var noteID int64
		if _, err := fmt.Sscanf(args[0], "%d", &noteID); err != nil {
			return fmt.Errorf("invalid note ID: %s", args[0])
		}

		force, _ := cmd.Flags().GetBool("force")

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Fetch the note to display before deletion
		note, err := db.SelectNoteByID(database, noteID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("note with ID %d not found", noteID)
		} else if err != nil {
			return fmt.Errorf("failed to fetch note: %w", err)
		}

		// Display note info
		fmt.Printf("Note %d (category: %s)\n", note.ID, note.Category)

		// Prompt for confirmation unless --force
		if !force {
			fmt.Print("Are you sure you want to delete this note? [y/N] ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Delete the note (cascade handles children)
		if err := db.DeleteNote(database, noteID); err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}

		fmt.Printf("Note %d deleted.\n", noteID)
		return nil
	},
}

// nullStringValue returns the string value or empty string if NULL.
func nullStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// joinStrings joins strings with a separator (simple helper to avoid importing strings package).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func init() {
	// Add flags to note add command
	noteAddCmd.Flags().StringP("category", "c", "", "Note category")
	noteAddCmd.Flags().StringP("text", "x", "", "Note text")

	// Add flags to note delete command
	noteDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Build command tree
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteGotoCmd)
	rootCmd.AddCommand(noteCmd)
}
