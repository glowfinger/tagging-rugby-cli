INSERT INTO note_clips (note_id, folder, filename, extension, format, filesize, status, started_at, finished_at, error_at, log)
VALUES (?, ?, ?, 'mp4', 'mp4', 0, 'pending', NULL, NULL, NULL, '')
ON CONFLICT(note_id) DO UPDATE SET
    folder=excluded.folder,
    filename=excluded.filename,
    extension='mp4',
    format='mp4',
    filesize=0,
    status='pending',
    started_at=NULL,
    finished_at=NULL,
    error_at=NULL,
    log=''
