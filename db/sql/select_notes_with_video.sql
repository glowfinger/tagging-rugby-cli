SELECT
    n.id,
    n.category,
    n.created_at,
    nv.path,
    nv.size,
    nv.duration,
    nv.format,
    nv.stopped_at
FROM notes n
INNER JOIN note_videos nv ON nv.note_id = n.id
WHERE nv.path = ?
ORDER BY n.created_at DESC;
