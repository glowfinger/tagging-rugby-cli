SELECT id, start_seconds, end_seconds, category, description FROM clips WHERE video_path = ? ORDER BY start_seconds ASC;
