package timeutil

import (
	"fmt"
	"strings"
)

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

// ParseTimeToSeconds parses a time string in HH:MM:SS, MM:SS, or raw seconds format.
// Uses colon count: 2 colons = H:M:S, 1 colon = M:S, 0 colons = raw seconds.
func ParseTimeToSeconds(timeStr string) (float64, error) {
	colons := strings.Count(timeStr, ":")

	switch colons {
	case 2:
		// HH:MM:SS format
		var hours, minutes, seconds int
		if n, err := fmt.Sscanf(timeStr, "%d:%d:%d", &hours, &minutes, &seconds); n == 3 && err == nil {
			return float64(hours*3600 + minutes*60 + seconds), nil
		}
	case 1:
		// MM:SS format
		var minutes, seconds int
		if n, err := fmt.Sscanf(timeStr, "%d:%d", &minutes, &seconds); n == 2 && err == nil {
			return float64(minutes*60 + seconds), nil
		}
	case 0:
		// Raw seconds (float)
		var secs float64
		if n, err := fmt.Sscanf(timeStr, "%f", &secs); n == 1 && err == nil {
			return secs, nil
		}
	}

	return 0, fmt.Errorf("expected HH:MM:SS, MM:SS, or seconds, got '%s'", timeStr)
}
