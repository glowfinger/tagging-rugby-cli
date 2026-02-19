SELECT v.id, n.id, v.path, v.filesize, 0.0, v.format
FROM notes n
INNER JOIN videos v ON v.id = n.video_id
WHERE n.id = ?;
