package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "tagging-rugby-cli",
	Short: "A CLI tool for rugby match analysis",
	Long: `tagging-rugby-cli is a CLI tool for rugby coaches and analysts
to review game footage via mpv with timestamped annotations stored in SQLite.

Features:
  - Open video files in mpv for analysis
  - Add timestamped notes, clips, and tackle events
  - Filter and search annotations
  - Export clips and statistics`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tagging-rugby-cli version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
