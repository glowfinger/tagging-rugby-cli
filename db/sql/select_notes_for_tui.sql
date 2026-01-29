SELECT id, timestamp_seconds, text, category, player, team
FROM notes
WHERE video_path = ?
ORDER BY timestamp_seconds ASC;
