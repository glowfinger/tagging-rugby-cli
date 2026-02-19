SELECT
    nc.id,
    nc.note_id,
    nc.folder,
    nc.filename,
    v.path,
    n.category,
    nt.player,
    nt.attempt,
    nt.outcome,
    ntim.start,
    ntim.end
FROM note_clips nc
INNER JOIN notes n ON n.id = nc.note_id
INNER JOIN videos v ON v.id = n.video_id
INNER JOIN note_timing ntim ON ntim.note_id = nc.note_id
INNER JOIN note_tackles nt ON nt.note_id = nc.note_id
WHERE nc.status = 'pending'
ORDER BY nc.id ASC
LIMIT 1
