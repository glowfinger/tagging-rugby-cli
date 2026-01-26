package mpv

import (
	"os/exec"

	"github.com/user/tagging-rugby-cli/deps"
)

// LaunchMpv starts mpv with the specified video file and IPC socket enabled.
// It checks that mpv is installed first and returns an error with install link if not.
// Returns the *exec.Cmd for the running process which can be used for cleanup.
func LaunchMpv(videoPath string) (*exec.Cmd, error) {
	// Check that mpv is installed
	if err := deps.CheckMpv(); err != nil {
		return nil, err
	}

	// Launch mpv with IPC socket flag
	cmd := exec.Command("mpv",
		"--input-ipc-server="+DefaultSocketPath,
		videoPath,
	)

	// Start the process (non-blocking)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}
