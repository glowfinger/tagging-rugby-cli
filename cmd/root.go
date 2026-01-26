package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/deps"
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

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies",
	Long:  `Check that all required system dependencies (mpv, ffmpeg) are installed and available.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking dependencies...")
		fmt.Println()

		allGood := true

		// Check mpv
		if err := deps.CheckMpv(); err != nil {
			fmt.Println("✗ mpv: NOT FOUND")
			fmt.Printf("  Install from: %s\n", deps.MpvInstallURL)
			allGood = false
		} else {
			fmt.Println("✓ mpv: OK")
		}

		// Check ffmpeg
		if err := deps.CheckFfmpeg(); err != nil {
			fmt.Println("✗ ffmpeg: NOT FOUND")
			fmt.Printf("  Install from: %s\n", deps.FfmpegInstallURL)
			allGood = false
		} else {
			fmt.Println("✓ ffmpeg: OK")
		}

		fmt.Println()
		if allGood {
			fmt.Println("All dependencies are installed!")
		} else {
			fmt.Println("Some dependencies are missing. Please install them to use all features.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
