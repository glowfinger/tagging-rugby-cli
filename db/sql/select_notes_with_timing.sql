SELECT
    n.id,
    n.category,
    n.created_at,
    nt.start,
    nt.end
FROM notes n
LEFT JOIN note_timing nt ON nt.note_id = n.id
ORDER BY nt.start ASC;
