SELECT
    nt.player,
    ntm.start AS timestamp,
    ntm.start AS clip_start,
    ntm.end AS clip_end,
    nt.note_id
FROM note_tackles nt
INNER JOIN note_timing ntm ON ntm.note_id = nt.note_id
INNER JOIN note_videos nv ON nv.note_id = nt.note_id
WHERE nt.player IS NOT NULL AND nt.player != ''
    AND nv.path = ?
ORDER BY nt.player ASC, ntm.start ASC;
