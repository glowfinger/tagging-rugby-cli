package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/mpv"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage timestamped notes",
	Long:  `Add, list, edit, and delete timestamped notes for video analysis.`,
}

var noteAddCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Add a note at the current timestamp",
	Long:  `Add a timestamped note at the current video position. The note is associated with the current video file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]

		// Get flags
		category, _ := cmd.Flags().GetString("category")
		player, _ := cmd.Flags().GetString("player")
		team, _ := cmd.Flags().GetString("team")

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

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Insert note
		result, err := database.Exec(
			`INSERT INTO notes (video_path, timestamp_seconds, text, category, player, team) VALUES (?, ?, ?, ?, ?, ?)`,
			videoPath, timestamp, text, category, player, team,
		)
		if err != nil {
			return fmt.Errorf("failed to insert note: %w", err)
		}

		noteID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get note ID: %w", err)
		}

		// Format timestamp as MM:SS
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		fmt.Printf("Note added: ID %d at %d:%02d\n", noteID, minutes, seconds)
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

		// Query notes
		rows, err := database.Query(
			`SELECT id, timestamp_seconds, category, player, team, text FROM notes WHERE video_path = ? ORDER BY timestamp_seconds ASC`,
			videoPath,
		)
		if err != nil {
			return fmt.Errorf("failed to query notes: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTime\tCategory\tPlayer\tTeam\tText")
		fmt.Fprintln(w, "--\t----\t--------\t------\t----\t----")

		count := 0
		for rows.Next() {
			var id int64
			var timestamp float64
			var category, player, team, text sql.NullString

			if err := rows.Scan(&id, &timestamp, &category, &player, &team, &text); err != nil {
				return fmt.Errorf("failed to scan note: %w", err)
			}

			// Format timestamp as MM:SS
			minutes := int(timestamp) / 60
			seconds := int(timestamp) % 60
			timeStr := fmt.Sprintf("%d:%02d", minutes, seconds)

			// Handle NULL values
			catStr := nullStringValue(category)
			playerStr := nullStringValue(player)
			teamStr := nullStringValue(team)
			textStr := nullStringValue(text)

			// Truncate text if too long
			if len(textStr) > 40 {
				textStr = textStr[:37] + "..."
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", id, timeStr, catStr, playerStr, teamStr, textStr)
			count++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating notes: %w", err)
		}

		w.Flush()

		if count == 0 {
			fmt.Println("\nNo notes found for this video.")
		} else {
			fmt.Printf("\n%d note(s) found.\n", count)
		}

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

func init() {
	// Add flags to note add command
	noteAddCmd.Flags().StringP("category", "c", "", "Note category (e.g., try, tackle, turnover)")
	noteAddCmd.Flags().StringP("player", "p", "", "Player name or number")
	noteAddCmd.Flags().StringP("team", "t", "", "Team name")

	// Build command tree
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	rootCmd.AddCommand(noteCmd)
}
