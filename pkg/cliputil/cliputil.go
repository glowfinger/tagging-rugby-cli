package cliputil

// CalculateClipBounds returns the start and end times for a clip.
// If clipStart and clipEnd are both non-zero and different, those values are used directly.
// Otherwise, start defaults to max(0, timestamp-4) and end defaults to min(videoDuration, timestamp+10).
// The returned values are clamped to [0, videoDuration].
func CalculateClipBounds(timestamp, clipStart, clipEnd, videoDuration float64) (start, end float64) {
	if clipStart != 0 && clipEnd != 0 && clipStart != clipEnd {
		start = clipStart
		end = clipEnd
	} else {
		start = timestamp - 4.0
		if start < 0 {
			start = 0
		}
		end = timestamp + 10.0
		if end > videoDuration {
			end = videoDuration
		}
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if end > videoDuration {
		end = videoDuration
	}

	return start, end
}
