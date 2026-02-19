INSERT INTO video_timings (video_id, stopped, length) VALUES (?, ?, 0) ON CONFLICT(video_id) DO UPDATE SET stopped = excluded.stopped;
