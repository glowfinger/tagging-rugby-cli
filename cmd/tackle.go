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
	Long:  `Record a tackle event at the current video position with player, team, attempt number, and outcome.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get required flags
		player, _ := cmd.Flags().GetString("player")
		team, _ := cmd.Flags().GetString("team")
		attempt, _ := cmd.Flags().GetInt("attempt")
		outcome, _ := cmd.Flags().GetString("outcome")

		// Validate required flags
		if player == "" {
			return fmt.Errorf("--player is required")
		}
		if team == "" {
			return fmt.Errorf("--team is required")
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

		// Get optional flags
		followed, _ := cmd.Flags().GetString("followed")
		star, _ := cmd.Flags().GetBool("star")
		notes, _ := cmd.Flags().GetString("notes")
		zone, _ := cmd.Flags().GetString("zone")

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

		// Convert star bool to int
		starInt := 0
		if star {
			starInt = 1
		}

		// Insert tackle
		result, err := database.Exec(
			`INSERT INTO tackles (video_path, timestamp_seconds, player, team, attempt, outcome, followed, star, notes, zone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			videoPath, timestamp, player, team, attempt, outcome, followed, starInt, notes, zone,
		)
		if err != nil {
			return fmt.Errorf("failed to insert tackle: %w", err)
		}

		tackleID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get tackle ID: %w", err)
		}

		// Format timestamp as MM:SS
		minutes := int(timestamp) / 60
		seconds := int(timestamp) % 60

		fmt.Printf("Tackle recorded: ID %d at %d:%02d\n", tackleID, minutes, seconds)
		fmt.Printf("  Player: %s, Team: %s, Attempt: %d, Outcome: %s\n", player, team, attempt, outcome)
		if star {
			fmt.Println("  ★ Starred")
		}
		return nil
	},
}

var tackleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tackles for the current video",
	Long:  `Display all tackles for the current video as a table, sorted by timestamp. Supports filtering by player, outcome, and starred status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get filter flags
		playerFilter, _ := cmd.Flags().GetString("player")
		outcomeFilter, _ := cmd.Flags().GetString("outcome")
		starFilter, _ := cmd.Flags().GetBool("star")
		starFilterSet := cmd.Flags().Changed("star")

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
		query := `SELECT id, timestamp_seconds, player, team, attempt, outcome, followed, star, notes, zone FROM tackles WHERE video_path = ?`
		queryArgs := []interface{}{videoPath}

		if playerFilter != "" {
			query += " AND player = ?"
			queryArgs = append(queryArgs, playerFilter)
		}
		if outcomeFilter != "" {
			query += " AND outcome = ?"
			queryArgs = append(queryArgs, outcomeFilter)
		}
		if starFilterSet && starFilter {
			query += " AND star = 1"
		}

		query += " ORDER BY timestamp_seconds ASC"

		// Query tackles
		rows, err := database.Query(query, queryArgs...)
		if err != nil {
			return fmt.Errorf("failed to query tackles: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTime\tPlayer\tTeam\tAttempt\tOutcome\tFollowed\tStar\tZone\tNotes")
		fmt.Fprintln(w, "--\t----\t------\t----\t-------\t-------\t--------\t----\t----\t-----")

		count := 0
		for rows.Next() {
			var id int64
			var timestamp float64
			var attemptVal int
			var starVal int
			var player, team, outcome sql.NullString
			var followed, notes, zone sql.NullString

			if err := rows.Scan(&id, &timestamp, &player, &team, &attemptVal, &outcome, &followed, &starVal, &notes, &zone); err != nil {
				return fmt.Errorf("failed to scan tackle: %w", err)
			}

			// Format timestamp as MM:SS
			minutes := int(timestamp) / 60
			seconds := int(timestamp) % 60
			timeStr := fmt.Sprintf("%d:%02d", minutes, seconds)

			// Handle NULL values
			playerStr := nullStringValue(player)
			teamStr := nullStringValue(team)
			outcomeStr := nullStringValue(outcome)
			followedStr := nullStringValue(followed)
			notesStr := nullStringValue(notes)
			zoneStr := nullStringValue(zone)

			// Star indicator
			starStr := ""
			if starVal == 1 {
				starStr = "★"
			}

			// Truncate notes if too long
			if len(notesStr) > 20 {
				notesStr = notesStr[:17] + "..."
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
				id, timeStr, playerStr, teamStr, attemptVal, outcomeStr, followedStr, starStr, zoneStr, notesStr)
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
	Long:  `Export detailed tackle statistics for a specific player to a text file. Includes total counts, completion percentage, and a list of all tackle events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get required --player flag
		player, _ := cmd.Flags().GetString("player")
		if player == "" {
			return fmt.Errorf("--player is required")
		}

		// Get optional flags
		outputPath, _ := cmd.Flags().GetString("output")
		videoFilter, _ := cmd.Flags().GetString("video")

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

		// Build query for tackle counts
		countQuery := `SELECT
			COUNT(*) as total,
			SUM(CASE WHEN outcome = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN outcome = 'missed' THEN 1 ELSE 0 END) as missed,
			SUM(CASE WHEN outcome = 'possible' THEN 1 ELSE 0 END) as possible,
			SUM(CASE WHEN outcome = 'other' THEN 1 ELSE 0 END) as other
			FROM tackles WHERE player = ?`
		countArgs := []interface{}{player}

		if videoFilter != "" {
			countQuery += " AND video_path = ?"
			countArgs = append(countArgs, videoFilter)
		}

		// Get counts
		var total, completed, missed, possible, other int
		err = database.QueryRow(countQuery, countArgs...).Scan(&total, &completed, &missed, &possible, &other)
		if err != nil {
			return fmt.Errorf("failed to query tackle counts: %w", err)
		}

		if total == 0 {
			return fmt.Errorf("no tackles found for player '%s'", player)
		}

		// Calculate completion percentage
		completionPct := 0.0
		if total > 0 {
			completionPct = float64(completed) / float64(total) * 100
		}

		// Build query for tackle details
		detailQuery := `SELECT timestamp_seconds, video_path, team, attempt, outcome, followed, star, notes, zone, created_at
			FROM tackles WHERE player = ?`
		detailArgs := []interface{}{player}

		if videoFilter != "" {
			detailQuery += " AND video_path = ?"
			detailArgs = append(detailArgs, videoFilter)
		}
		detailQuery += " ORDER BY created_at ASC, timestamp_seconds ASC"

		// Query tackle details
		rows, err := database.Query(detailQuery, detailArgs...)
		if err != nil {
			return fmt.Errorf("failed to query tackle details: %w", err)
		}
		defer rows.Close()

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
		fmt.Fprintf(file, "Total:      %d\n", total)
		fmt.Fprintf(file, "Completed:  %d\n", completed)
		fmt.Fprintf(file, "Missed:     %d\n", missed)
		fmt.Fprintf(file, "Possible:   %d\n", possible)
		fmt.Fprintf(file, "Other:      %d\n", other)
		fmt.Fprintf(file, "Completion: %.1f%%\n\n", completionPct)

		// Write tackle events header
		fmt.Fprintf(file, "Tackle Events\n")
		fmt.Fprintf(file, "-------------\n")

		// Write each tackle event
		eventNum := 1
		for rows.Next() {
			var timestamp float64
			var attemptVal int
			var starVal int
			var videoPath, team, outcome sql.NullString
			var followed, notes, zone sql.NullString
			var createdAt sql.NullString

			if err := rows.Scan(&timestamp, &videoPath, &team, &attemptVal, &outcome, &followed, &starVal, &notes, &zone, &createdAt); err != nil {
				return fmt.Errorf("failed to scan tackle: %w", err)
			}

			// Format timestamp as MM:SS
			minutes := int(timestamp) / 60
			seconds := int(timestamp) % 60
			timeStr := fmt.Sprintf("%d:%02d", minutes, seconds)

			// Star symbol for starred tackles
			starSymbol := ""
			if starVal == 1 {
				starSymbol = " ★"
			}

			// Write event line
			fmt.Fprintf(file, "\n%d. [%s] %s - Attempt #%d%s\n",
				eventNum, timeStr, nullStringValue(outcome), attemptVal, starSymbol)

			// Write optional details
			if team.Valid && team.String != "" {
				fmt.Fprintf(file, "   Team: %s\n", team.String)
			}
			if followed.Valid && followed.String != "" {
				fmt.Fprintf(file, "   Followed by: %s\n", followed.String)
			}
			if zone.Valid && zone.String != "" {
				fmt.Fprintf(file, "   Zone: %s\n", zone.String)
			}
			if notes.Valid && notes.String != "" {
				fmt.Fprintf(file, "   Notes: %s\n", notes.String)
			}
			if videoPath.Valid && videoPath.String != "" {
				fmt.Fprintf(file, "   Video: %s\n", videoPath.String)
			}

			eventNum++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating tackles: %w", err)
		}

		fmt.Printf("Exported %d tackles for %s to %s\n", total, player, outputPath)
		return nil
	},
}

func init() {
	// Add required flags to tackle add command
	tackleAddCmd.Flags().StringP("player", "p", "", "Player name or number (required)")
	tackleAddCmd.Flags().StringP("team", "t", "", "Team name (required)")
	tackleAddCmd.Flags().IntP("attempt", "a", 0, "Tackle attempt number (required)")
	tackleAddCmd.Flags().StringP("outcome", "o", "", "Tackle outcome: missed, completed, possible, other (required)")

	// Add optional flags to tackle add command
	tackleAddCmd.Flags().StringP("followed", "f", "", "Who followed up on the tackle")
	tackleAddCmd.Flags().BoolP("star", "s", false, "Mark this tackle as starred/important")
	tackleAddCmd.Flags().StringP("notes", "n", "", "Additional notes about the tackle")
	tackleAddCmd.Flags().StringP("zone", "z", "", "Field zone where the tackle occurred")

	// Add filter flags to tackle list command
	tackleListCmd.Flags().StringP("player", "p", "", "Filter by player name or number")
	tackleListCmd.Flags().StringP("outcome", "o", "", "Filter by outcome: missed, completed, possible, other")
	tackleListCmd.Flags().BoolP("star", "s", false, "Filter to show only starred tackles")

	// Add flags to tackle export command
	tackleExportCmd.Flags().StringP("player", "p", "", "Player name or number to export (required)")
	tackleExportCmd.Flags().StringP("output", "o", "", "Output file path (default: <player>-tackles.txt)")
	tackleExportCmd.Flags().StringP("video", "v", "", "Filter to specific video path")

	// Build command tree
	tackleCmd.AddCommand(tackleAddCmd)
	tackleCmd.AddCommand(tackleListCmd)
	tackleCmd.AddCommand(tackleExportCmd)
	rootCmd.AddCommand(tackleCmd)
}
