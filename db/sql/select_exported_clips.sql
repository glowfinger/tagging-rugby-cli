SELECT
    nc.note_id,
    nc.name,
    ntk.player,
    n.category,
    ntk.outcome,
    nc.duration,
    CASE
        WHEN nc.finished_at IS NOT NULL THEN 'completed'
        WHEN nc.error_at IS NOT NULL THEN 'error'
        ELSE 'pending'
    END AS status,
    COALESCE(nc.error, '') AS error
FROM note_clips nc
JOIN notes n ON n.id = nc.note_id
JOIN note_tackles ntk ON ntk.note_id = nc.note_id
LEFT JOIN note_videos nv ON nv.note_id = nc.note_id
ORDER BY nv.path ASC, nc.name ASC
