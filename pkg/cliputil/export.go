package cliputil

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ExtractClip runs ffmpeg to extract a video clip from inputPath.
// The clip starts at 'start' seconds and has a duration of (end - start) seconds.
// Output is encoded as H.264 video with AAC audio.
func ExtractClip(inputPath string, start, end float64, outputPath string) error {
	duration := end - start

	cmd := exec.Command("ffmpeg",
		"-y",
		"-ss", fmt.Sprintf("%.3f", start),
		"-i", inputPath,
		"-t", fmt.Sprintf("%.3f", duration),
		"-c:v", "libx264",
		"-c:a", "aac",
		"-preset", "fast",
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg error: %s", stderr.String())
	}

	return nil
}
