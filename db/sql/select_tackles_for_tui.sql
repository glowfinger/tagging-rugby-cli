SELECT id, timestamp_seconds, player, team, outcome, notes, star
FROM tackles
WHERE video_path = ?
ORDER BY timestamp_seconds ASC;
