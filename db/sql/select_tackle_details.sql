SELECT timestamp_seconds, video_path, team, attempt, outcome, followed, star, notes, zone, created_at
FROM tackles WHERE player = ?
