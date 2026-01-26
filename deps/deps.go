package deps

import (
	"fmt"
	"os/exec"
)

const (
	MpvInstallURL    = "https://mpv.io/installation/"
	FfmpegInstallURL = "https://ffmpeg.org/download.html"
)

// DependencyError contains information about a missing dependency
type DependencyError struct {
	Name       string
	InstallURL string
}

func (e *DependencyError) Error() string {
	return fmt.Sprintf("%s not found. Install from: %s", e.Name, e.InstallURL)
}

// CheckMpv checks if mpv is installed and available in PATH
func CheckMpv() error {
	_, err := exec.LookPath("mpv")
	if err != nil {
		return &DependencyError{
			Name:       "mpv",
			InstallURL: MpvInstallURL,
		}
	}
	return nil
}

// CheckFfmpeg checks if ffmpeg is installed and available in PATH
func CheckFfmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return &DependencyError{
			Name:       "ffmpeg",
			InstallURL: FfmpegInstallURL,
		}
	}
	return nil
}

// CheckAll checks all dependencies and returns a slice of errors for missing ones
func CheckAll() []error {
	var errors []error

	if err := CheckMpv(); err != nil {
		errors = append(errors, err)
	}

	if err := CheckFfmpeg(); err != nil {
		errors = append(errors, err)
	}

	return errors
}
