package timeutil

import "fmt"

// FormatTime formats seconds as H:MM:SS (e.g. 0:01:30, 1:11:22).
func FormatTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	mins := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, mins, secs)
}
