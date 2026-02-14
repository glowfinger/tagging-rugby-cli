SELECT
    n.id,
    n.category,
    nt.start,
    nt.end,
    nv.path,
    ntk.player,
    ntk.outcome,
    nc.finished_at
FROM notes n
JOIN note_timing nt ON nt.note_id = n.id
JOIN note_videos nv ON nv.note_id = n.id
JOIN note_tackles ntk ON ntk.note_id = n.id
LEFT JOIN note_clips nc ON nc.note_id = n.id
WHERE n.category = 'tackle'
ORDER BY nv.path ASC, nt.start ASC
