package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/db"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage annotation categories",
	Long:  `Add, list, and delete annotation categories used for tagging notes and clips.`,
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	Long:  `Display all available annotation categories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Query categories
		rows, err := database.Query(`SELECT id, name FROM categories ORDER BY name ASC`)
		if err != nil {
			return fmt.Errorf("failed to query categories: %w", err)
		}
		defer rows.Close()

		// Create table writer
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tName")
		fmt.Fprintln(w, "--\t----")

		count := 0
		for rows.Next() {
			var id int64
			var name string

			if err := rows.Scan(&id, &name); err != nil {
				return fmt.Errorf("failed to scan category: %w", err)
			}

			fmt.Fprintf(w, "%d\t%s\n", id, name)
			count++
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating categories: %w", err)
		}

		w.Flush()

		if count == 0 {
			fmt.Println("\nNo categories found.")
		} else {
			fmt.Printf("\n%d category(ies) found.\n", count)
		}

		return nil
	},
}

var categoryAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new category",
	Long:  `Add a new annotation category. Category names must be unique.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Insert category
		result, err := database.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
		if err != nil {
			// Check if it's a unique constraint violation
			if isUniqueConstraintError(err) {
				return fmt.Errorf("category '%s' already exists", name)
			}
			return fmt.Errorf("failed to add category: %w", err)
		}

		categoryID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get category ID: %w", err)
		}

		fmt.Printf("Category added: ID %d, name '%s'\n", categoryID, name)
		return nil
	},
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a category",
	Long:  `Delete an annotation category by name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Open database
		database, err := db.Open()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer database.Close()

		// Delete category
		result, err := database.Exec(`DELETE FROM categories WHERE name = ?`, name)
		if err != nil {
			return fmt.Errorf("failed to delete category: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check deletion result: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("category '%s' not found", name)
		}

		fmt.Printf("Category '%s' deleted.\n", name)
		return nil
	},
}

// isUniqueConstraintError checks if an error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	// SQLite unique constraint error contains "UNIQUE constraint failed"
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "UNIQUE constraint"))
}

// contains checks if a string contains a substring (simple helper to avoid importing strings package).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	// Build command tree
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryAddCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	rootCmd.AddCommand(categoryCmd)
}
