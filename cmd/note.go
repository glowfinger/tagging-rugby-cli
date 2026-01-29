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
			db.InsertNoteSQL,
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
	Long:  `Display all notes for the current video as a table, sorted by timestamp. Supports filtering by category, player, team, and time range.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get filter flags
		categoryFilter, _ := cmd.Flags().GetString("category")
		playerFilter, _ := cmd.Flags().GetString("player")
		teamFilter, _ := cmd.Flags().GetString("team")
		fromFilter, _ := cmd.Flags().GetString("from")
		toFilter, _ := cmd.Flags().GetString("to")

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

		// Build dynamic query with filters
		query := db.SelectNotesByVideoSQL
		queryArgs := []interface{}{videoPath}

		if categoryFilter != "" {
			query += " AND category = ?"
			queryArgs = append(queryArgs, categoryFilter)
		}
		if playerFilter != "" {
			query += " AND player = ?"
			queryArgs = append(queryArgs, playerFilter)
		}
		if teamFilter != "" {
			query += " AND team = ?"
			queryArgs = append(queryArgs, teamFilter)
		}

		// Parse and apply time range filters
		if fromFilter != "" {
			fromSeconds, err := parseTimeToSeconds(fromFilter)
			if err != nil {
				return fmt.Errorf("invalid --from time format: %w", err)
			}
			query += " AND timestamp_seconds >= ?"
			queryArgs = append(queryArgs, fromSeconds)
		}
		if toFilter != "" {
			toSeconds, err := parseTimeToSeconds(toFilter)
			if err != nil {
				return fmt.Errorf("invalid --to time format: %w", err)
			}
			query += " AND timestamp_seconds <= ?"
			queryArgs = append(queryArgs, toSeconds)
		}

		query += " ORDER BY timestamp_seconds ASC"

		// Query notes
		rows, err := database.Query(query, queryArgs...)
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
			fmt.Println("\nNo matching notes found.")
		} else {
			fmt.Printf("\n%d note(s) found.\n", count)
		}

		return nil
	},
}

var noteEditCmd = &cobra.Command{
	Use:   "edit <id> <text>",
	Short: "Edit a note's text and/or metadata",
	Long:  `Edit an existing note by ID. Update the text, category, player, team, or timestamp.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var noteID int64
		if _, err := fmt.Sscanf(args[0], "%d", &noteID); err != nil {
			return fmt.Errorf("invalid note ID: %s", args[0])
		}
		newText := args[1]

		// Get flags
		category, _ := cmd.Flags().GetString("category")
		player, _ := cmd.Flags().GetString("player")
		team, _ := cmd.Flags().GetString("team")
		updateTimestamp, _ := cmd.Flags().GetBool("timestamp")

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Check if note exists
		var existingID int64
		err = database.QueryRow(db.SelectNoteIDSQL, noteID).Scan(&existingID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("note with ID %d not found", noteID)
		} else if err != nil {
			return fmt.Errorf("failed to check note: %w", err)
		}

		// Build the update query dynamically based on which flags were set
		updateFields := []string{"text = ?"}
		updateArgs := []interface{}{newText}

		if cmd.Flags().Changed("category") {
			updateFields = append(updateFields, "category = ?")
			updateArgs = append(updateArgs, category)
		}
		if cmd.Flags().Changed("player") {
			updateFields = append(updateFields, "player = ?")
			updateArgs = append(updateArgs, player)
		}
		if cmd.Flags().Changed("team") {
			updateFields = append(updateFields, "team = ?")
			updateArgs = append(updateArgs, team)
		}

		// If --timestamp flag is set, get current position from mpv
		if updateTimestamp {
			client := mpv.NewClient("")
			if err := client.Connect(); err != nil {
				return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
			}
			defer client.Close()

			timestamp, err := client.GetTimePos()
			if err != nil {
				return fmt.Errorf("failed to get current timestamp: %w", err)
			}
			updateFields = append(updateFields, "timestamp_seconds = ?")
			updateArgs = append(updateArgs, timestamp)
		}

		// Build and execute the update query
		query := fmt.Sprintf("UPDATE notes SET %s WHERE id = ?", joinStrings(updateFields, ", "))
		updateArgs = append(updateArgs, noteID)

		_, err = database.Exec(query, updateArgs...)
		if err != nil {
			return fmt.Errorf("failed to update note: %w", err)
		}

		// Fetch and display the updated note
		var timestamp float64
		var categoryVal, playerVal, teamVal, textVal sql.NullString
		err = database.QueryRow(
			db.SelectNoteDetailsSQL,
			noteID,
		).Scan(&timestamp, &categoryVal, &playerVal, &teamVal, &textVal)
		if err != nil {
			return fmt.Errorf("failed to fetch updated note: %w", err)
		}

		// Format and display
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		fmt.Printf("Note %d updated:\n", noteID)
		fmt.Printf("  Time:     %d:%02d\n", minutes, seconds)
		fmt.Printf("  Text:     %s\n", nullStringValue(textVal))
		if categoryVal.Valid && categoryVal.String != "" {
			fmt.Printf("  Category: %s\n", categoryVal.String)
		}
		if playerVal.Valid && playerVal.String != "" {
			fmt.Printf("  Player:   %s\n", playerVal.String)
		}
		if teamVal.Valid && teamVal.String != "" {
			fmt.Printf("  Team:     %s\n", teamVal.String)
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
		var timestamp float64
		var categoryVal, playerVal, teamVal, textVal sql.NullString
		err = database.QueryRow(
			db.SelectNoteDetailsSQL,
			noteID,
		).Scan(&timestamp, &categoryVal, &playerVal, &teamVal, &textVal)
		if err == sql.ErrNoRows {
			return fmt.Errorf("note with ID %d not found", noteID)
		} else if err != nil {
			return fmt.Errorf("failed to fetch note: %w", err)
		}

		// Connect to mpv
		client := mpv.NewClient("")
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to mpv: %w\n(Is mpv running with a video open?)", err)
		}
		defer client.Close()

		// Seek to the note's timestamp
		if err := client.Seek(timestamp); err != nil {
			return fmt.Errorf("failed to seek to timestamp: %w", err)
		}

		// Format timestamp
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		// Display note details
		fmt.Printf("Jumped to note %d at %d:%02d\n", noteID, minutes, seconds)
		fmt.Printf("  Text:     %s\n", nullStringValue(textVal))
		if categoryVal.Valid && categoryVal.String != "" {
			fmt.Printf("  Category: %s\n", categoryVal.String)
		}
		if playerVal.Valid && playerVal.String != "" {
			fmt.Printf("  Player:   %s\n", playerVal.String)
		}
		if teamVal.Valid && teamVal.String != "" {
			fmt.Printf("  Team:     %s\n", teamVal.String)
		}

		return nil
	},
}

var noteDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a note",
	Long:  `Delete an existing note by ID. Prompts for confirmation unless --force is used.`,
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
		var timestamp float64
		var textVal sql.NullString
		err = database.QueryRow(
			db.SelectNoteBriefSQL,
			noteID,
		).Scan(&timestamp, &textVal)
		if err == sql.ErrNoRows {
			return fmt.Errorf("note with ID %d not found", noteID)
		} else if err != nil {
			return fmt.Errorf("failed to fetch note: %w", err)
		}

		// Format timestamp
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		// Display note info
		fmt.Printf("Note %d at %d:%02d: %s\n", noteID, minutes, seconds, nullStringValue(textVal))

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

		// Delete the note
		result, err := database.Exec(db.DeleteNoteSQL, noteID)
		if err != nil {
			return fmt.Errorf("failed to delete note: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("note with ID %d not found", noteID)
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

// parseTimeToSeconds parses a time string in MM:SS or seconds format.
func parseTimeToSeconds(timeStr string) (float64, error) {
	// Try MM:SS format first
	var minutes, seconds int
	if n, err := fmt.Sscanf(timeStr, "%d:%d", &minutes, &seconds); n == 2 && err == nil {
		return float64(minutes*60 + seconds), nil
	}

	// Try seconds format (float)
	var secs float64
	if n, err := fmt.Sscanf(timeStr, "%f", &secs); n == 1 && err == nil {
		return secs, nil
	}

	return 0, fmt.Errorf("expected MM:SS or seconds, got '%s'", timeStr)
}

func init() {
	// Add flags to note add command
	noteAddCmd.Flags().StringP("category", "c", "", "Note category (e.g., try, tackle, turnover)")
	noteAddCmd.Flags().StringP("player", "p", "", "Player name or number")
	noteAddCmd.Flags().StringP("team", "t", "", "Team name")

	// Add flags to note list command for filtering
	noteListCmd.Flags().StringP("category", "c", "", "Filter by category")
	noteListCmd.Flags().StringP("player", "p", "", "Filter by player")
	noteListCmd.Flags().StringP("team", "t", "", "Filter by team")
	noteListCmd.Flags().String("from", "", "Filter notes from this time (MM:SS or seconds)")
	noteListCmd.Flags().String("to", "", "Filter notes up to this time (MM:SS or seconds)")

	// Add flags to note edit command
	noteEditCmd.Flags().StringP("category", "c", "", "Update note category")
	noteEditCmd.Flags().StringP("player", "p", "", "Update player name or number")
	noteEditCmd.Flags().StringP("team", "t", "", "Update team name")
	noteEditCmd.Flags().Bool("timestamp", false, "Update timestamp to current mpv position")

	// Add flags to note delete command
	noteDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Build command tree
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
	noteCmd.AddCommand(noteEditCmd)
	noteCmd.AddCommand(noteDeleteCmd)
	noteCmd.AddCommand(noteGotoCmd)
	rootCmd.AddCommand(noteCmd)
}
