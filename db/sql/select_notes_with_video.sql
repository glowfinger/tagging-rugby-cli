SELECT
    n.id,
    n.category,
    n.created_at,
    v.path,
    v.filesize,
    v.format,
    v.stop_time
FROM notes n
INNER JOIN videos v ON v.id = n.video_id
WHERE v.path = ?
ORDER BY n.created_at DESC;
