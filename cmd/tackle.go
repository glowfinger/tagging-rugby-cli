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

// Valid outcome values for tackles
var validOutcomes = []string{"missed", "completed", "possible", "other"}

var tackleCmd = &cobra.Command{
	Use:   "tackle",
	Short: "Manage tackle events",
	Long:  `Record and list tackle events with detailed tracking information.`,
}

var tackleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Record a tackle event at the current timestamp",
	Long:  `Record a tackle event at the current video position with player, attempt number, and outcome.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get required flags
		player, _ := cmd.Flags().GetString("player")
		attempt, _ := cmd.Flags().GetInt("attempt")
		outcome, _ := cmd.Flags().GetString("outcome")

		// Validate required flags
		if player == "" {
			return fmt.Errorf("--player is required")
		}
		if attempt == 0 {
			return fmt.Errorf("--attempt is required")
		}
		if outcome == "" {
			return fmt.Errorf("--outcome is required")
		}

		// Validate outcome value
		if !isValidOutcome(outcome) {
			return fmt.Errorf("invalid outcome '%s': must be one of: missed, completed, possible, other", outcome)
		}

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

		// Insert note with tackle and timing child rows
		children := db.NoteChildren{
			Tackles: []db.NoteTackle{
				{Player: player, Attempt: attempt, Outcome: outcome},
			},
			Timings: []db.NoteTiming{
				{Start: timestamp, End: timestamp},
			},
			Videos: []db.NoteVideo{
				{Path: videoPath, StoppedAt: timestamp},
			},
		}

		noteID, err := db.InsertNoteWithChildren(database, "tackle", children)
		if err != nil {
			return fmt.Errorf("failed to insert tackle: %w", err)
		}

		// Format timestamp as MM:SS
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		fmt.Printf("Tackle recorded: Note ID %d at %d:%02d\n", noteID, minutes, seconds)
		fmt.Printf("  Player: %s, Attempt: %d, Outcome: %s\n", player, attempt, outcome)
		return nil
	},
}

var tackleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tackles for the current video",
	Long:  `Display all tackles for the current video as a table, sorted by timestamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get filter flags
		playerFilter, _ := cmd.Flags().GetString("player")
		outcomeFilter, _ := cmd.Flags().GetString("outcome")

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

		// Build dynamic query with filters - join notes with note_tackles, note_timing, and note_videos
		query := `SELECT n.id, COALESCE(nt_time.start, 0), ntk.player, ntk.attempt, ntk.outcome
			 FROM notes n
			 INNER JOIN note_tackles ntk ON ntk.note_id = n.id
			 INNER JOIN note_videos nv ON nv.note_id = n.id
			 LEFT JOIN note_timing nt_time ON nt_time.note_id = n.id
			 WHERE nv.path = ?`
		queryArgs := []interface{}{videoPath}

		if playerFilter != "" {
			query += " AND ntk.player = ?"
			queryArgs = append(queryArgs, playerFilter)
		}
		if outcomeFilter != "" {
			query += " AND ntk.outcome = ?"
			queryArgs = append(queryArgs, outcomeFilter)
		}

		query += " ORDER BY nt_time.start ASC"

		// Query tackles
		rows, err := database.Query(query, queryArgs...)
		if err != nil {
			return fmt.Errorf("failed to query tackles: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NoteID\tTime\tPlayer\tAttempt\tOutcome")
		fmt.Fprintln(w, "------\t----\t------\t-------\t-------")

		count := 0
		for rows.Next() {
			var noteID int64
			var timestamp float64
			var attemptVal int
			var player, outcome sql.NullString

			if err := rows.Scan(&noteID, &timestamp, &player, &attemptVal, &outcome); err != nil {
				return fmt.Errorf("failed to scan tackle: %w", err)
			}

			// Format timestamp as MM:SS
			minutes := int(timestamp) / 60
			seconds := int(timestamp) % 60
			timeStr := fmt.Sprintf("%d:%02d", minutes, seconds)

			playerStr := nullStringValue(player)
			outcomeStr := nullStringValue(outcome)

			fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
				noteID, timeStr, playerStr, attemptVal, outcomeStr)
			count++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating tackles: %w", err)
		}

		w.Flush()

		if count == 0 {
			fmt.Println("\nNo tackles found for this video.")
		} else {
			fmt.Printf("\n%d tackle(s) found.\n", count)
		}

		return nil
	},
}

// isValidOutcome checks if the outcome value is valid.
func isValidOutcome(outcome string) bool {
	for _, v := range validOutcomes {
		if v == outcome {
			return true
		}
	}
	return false
}

var tackleExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export player tackle statistics to a text file",
	Long:  `Export detailed tackle statistics for a specific player to a text file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get required --player flag
		player, _ := cmd.Flags().GetString("player")
		if player == "" {
			return fmt.Errorf("--player is required")
		}

		// Get optional flags
		outputPath, _ := cmd.Flags().GetString("output")

		// Set default output path if not specified
		if outputPath == "" {
			outputPath = player + "-tackles.txt"
		}

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Query tackle stats using the embedded SQL
		rows, err := database.Query(db.SelectTackleStatsSQL)
		if err != nil {
			return fmt.Errorf("failed to query tackle stats: %w", err)
		}
		defer rows.Close()

		// Find this player's stats
		var total, successCount, missedCount, dominantCount, passiveCount int
		found := false
		for rows.Next() {
			var p string
			var t, s, m, d, pa int
			if err := rows.Scan(&p, &t, &s, &m, &d, &pa); err != nil {
				return fmt.Errorf("failed to scan stats: %w", err)
			}
			if p == player {
				total = t
				successCount = s
				missedCount = m
				dominantCount = d
				passiveCount = pa
				found = true
				break
			}
		}
		rows.Close()

		if !found || total == 0 {
			return fmt.Errorf("no tackles found for player '%s'", player)
		}

		// Create output file
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		// Write header
		fmt.Fprintf(file, "Tackle Statistics for %s\n", player)
		fmt.Fprintf(file, "================================\n\n")

		// Write summary statistics
		fmt.Fprintf(file, "Summary\n")
		fmt.Fprintf(file, "-------\n")
		fmt.Fprintf(file, "Total:     %d\n", total)
		fmt.Fprintf(file, "Success:   %d\n", successCount)
		fmt.Fprintf(file, "Missed:    %d\n", missedCount)
		fmt.Fprintf(file, "Dominant:  %d\n", dominantCount)
		fmt.Fprintf(file, "Passive:   %d\n", passiveCount)

		fmt.Printf("Exported tackle stats for %s to %s\n", player, outputPath)
		return nil
	},
}

func init() {
	// Add required flags to tackle add command
	tackleAddCmd.Flags().StringP("player", "p", "", "Player name or number (required)")
	tackleAddCmd.Flags().IntP("attempt", "a", 0, "Tackle attempt number (required)")
	tackleAddCmd.Flags().StringP("outcome", "o", "", "Tackle outcome: missed, completed, possible, other (required)")

	// Add filter flags to tackle list command
	tackleListCmd.Flags().StringP("player", "p", "", "Filter by player name or number")
	tackleListCmd.Flags().StringP("outcome", "o", "", "Filter by outcome: missed, completed, possible, other")

	// Add flags to tackle export command
	tackleExportCmd.Flags().StringP("player", "p", "", "Player name or number to export (required)")
	tackleExportCmd.Flags().StringP("output", "o", "", "Output file path (default: <player>-tackles.txt)")

	// Build command tree
	tackleCmd.AddCommand(tackleAddCmd)
	tackleCmd.AddCommand(tackleListCmd)
	tackleCmd.AddCommand(tackleExportCmd)
	rootCmd.AddCommand(tackleCmd)
}
