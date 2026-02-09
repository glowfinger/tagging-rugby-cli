SELECT
    nt.player,
    COUNT(*) AS total,
    SUM(CASE WHEN nt.outcome = 'success' THEN 1 ELSE 0 END) AS success_count,
    SUM(CASE WHEN nt.outcome = 'missed' THEN 1 ELSE 0 END) AS missed_count,
    SUM(CASE WHEN nt.outcome = 'dominant' THEN 1 ELSE 0 END) AS dominant_count,
    SUM(CASE WHEN nt.outcome = 'passive' THEN 1 ELSE 0 END) AS passive_count
FROM note_tackles nt
GROUP BY nt.player
ORDER BY total DESC;
