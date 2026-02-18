-- Migration 002: Migrate note_videos data to videos table, update notes.video_id, drop note_videos
--
-- On fresh databases, note_videos has no rows (created by create_tables.sql but empty),
-- so the INSERT and UPDATE are no-ops. The table is then dropped safely.

-- Step 1: Insert unique videos from note_videos into videos, deduplicated by path.
-- filename = everything after the last '/' in path
-- extension = everything after the last '.' in filename
-- Maps: size -> filesize, stopped_at -> stop_time. duration is NOT migrated.
INSERT INTO videos (path, filename, extension, format, filesize, stop_time)
SELECT
    nv.path,
    CASE
        WHEN INSTR(nv.path, '/') > 0
        THEN SUBSTR(nv.path, LENGTH(nv.path) - LENGTH(REPLACE(nv.path, '/', '')) + 1)
        ELSE nv.path
    END AS filename,
    CASE
        WHEN INSTR(nv.path, '.') > 0
        THEN SUBSTR(nv.path, LENGTH(nv.path) - LENGTH(REPLACE(nv.path, '.', '')) + 1)
        ELSE ''
    END AS extension,
    nv.format,
    nv.size,
    nv.stopped_at
FROM note_videos nv
INNER JOIN (
    SELECT path, MIN(id) AS min_id
    FROM note_videos
    GROUP BY path
) dedup ON nv.id = dedup.min_id
WHERE EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='note_videos');

-- Step 2: Update notes.video_id to the matching videos.id based on the path from note_videos.
UPDATE notes
SET video_id = (
    SELECT v.id
    FROM videos v
    INNER JOIN note_videos nv ON nv.path = v.path
    WHERE nv.note_id = notes.id
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1 FROM note_videos nv WHERE nv.note_id = notes.id
)
AND EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='note_videos');

-- Step 3: Drop note_videos table.
DROP TABLE IF EXISTS note_videos;
