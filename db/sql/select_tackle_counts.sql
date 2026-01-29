SELECT
    COUNT(*) as total,
    SUM(CASE WHEN outcome = 'completed' THEN 1 ELSE 0 END) as completed,
    SUM(CASE WHEN outcome = 'missed' THEN 1 ELSE 0 END) as missed,
    SUM(CASE WHEN outcome = 'possible' THEN 1 ELSE 0 END) as possible,
    SUM(CASE WHEN outcome = 'other' THEN 1 ELSE 0 END) as other
FROM tackles WHERE player = ?
