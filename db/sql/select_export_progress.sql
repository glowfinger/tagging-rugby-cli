SELECT COUNT(DISTINCT n.id) AS total_tackles,
       COUNT(CASE WHEN nc.status = 'completed' THEN 1 END) AS completed_clips,
       COUNT(CASE WHEN nc.status = 'pending' THEN 1 END) AS pending_clips,
       COUNT(CASE WHEN nc.status = 'error' THEN 1 END) AS error_clips
FROM notes n
INNER JOIN videos v ON v.id = n.video_id
LEFT JOIN note_clips nc ON nc.note_id = n.id
WHERE v.path = ? AND n.category = 'tackle'
