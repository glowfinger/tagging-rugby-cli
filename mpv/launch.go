package mpv

import (
	"os"
	"os/exec"

	"github.com/user/tagging-rugby-cli/deps"
)

// LaunchMpv starts mpv with the specified video file and IPC socket enabled.
// It checks that mpv is installed first and returns an error with install link if not.
// Default keybindings are disabled; only space/q/f are active via embedded input.conf.
// Returns the *exec.Cmd and a cleanup function that removes the temporary input config.
func LaunchMpv(videoPath string) (*exec.Cmd, func(), error) {
	// Check that mpv is installed
	if err := deps.CheckMpv(); err != nil {
		return nil, nil, err
	}

	// Write embedded input.conf to a temporary file
	tmpFile, err := os.CreateTemp("", "tagging-rugby-input-*.conf")
	if err != nil {
		return nil, nil, err
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(InputConf); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, nil, err
	}
	tmpFile.Close()

	cleanup := func() { os.Remove(tmpPath) }

	// Launch mpv with IPC socket, disabled default bindings, and custom input config
	cmd := exec.Command("mpv",
		"--input-ipc-server="+DefaultSocketPath,
		"--no-input-default-bindings",
		"--input-conf="+tmpPath,
		videoPath,
	)

	// Start the process (non-blocking)
	if err := cmd.Start(); err != nil {
		cleanup()
		return nil, nil, err
	}

	return cmd, cleanup, nil
}
