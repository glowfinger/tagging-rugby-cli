package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/tagging-rugby-cli/deps"
	"github.com/user/tagging-rugby-cli/mpv"
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

var openCmd = &cobra.Command{
	Use:   "open <video-file>",
	Short: "Open a video file for analysis",
	Long:  `Open a video file in mpv for analysis. The video player will launch and the CLI can be used to add notes and annotations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		videoPath := args[0]

		// Resolve to absolute path
		absPath, err := filepath.Abs(videoPath)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Check video file exists
		info, err := os.Stat(absPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("video file not found: %s", absPath)
		}
		if err != nil {
			return fmt.Errorf("failed to access video file: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a video file: %s", absPath)
		}

		// Launch mpv with video file
		fmt.Printf("Opening video: %s\n", filepath.Base(absPath))
		process, err := mpv.LaunchMpv(absPath)
		if err != nil {
			return fmt.Errorf("failed to launch mpv: %w", err)
		}

		// Wait briefly for socket to be ready
		client := mpv.NewClient("")
		var connectErr error
		for i := 0; i < 50; i++ { // Wait up to 5 seconds
			time.Sleep(100 * time.Millisecond)
			connectErr = client.Connect()
			if connectErr == nil {
				break
			}
		}

		if connectErr != nil {
			// Kill mpv if we couldn't connect
			if process.Process != nil {
				process.Process.Kill()
			}
			return fmt.Errorf("failed to connect to mpv: %w", connectErr)
		}
		defer client.Close()

		// Get duration and print confirmation
		duration, err := client.GetDuration()
		if err != nil {
			fmt.Printf("Video session started: %s\n", filepath.Base(absPath))
		} else {
			minutes := int(duration) / 60
			seconds := int(duration) % 60
			fmt.Printf("Video session started: %s (duration: %d:%02d)\n", filepath.Base(absPath), minutes, seconds)
		}

		// Wait for mpv to exit
		return process.Wait()
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
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(doctorCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
