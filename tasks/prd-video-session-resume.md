# PRD: Video Session Resume

## Introduction

Improve video interaction by tracking per-video playback state in a dedicated `video_timings` table. When a user reopens a previously viewed video, the app automatically resumes from where they left off — paused at the last known position. The stopped position is updated whenever playback pauses or a note is created, and saved on exit. This replaces the unused `videos.stop_time` column with a proper 1-to-1 relationship table.

## Goals

- Remove `videos.stop_time` from the schema and replace with `video_timings` (1-to-1 with `videos`)
- Ensure a `videos` row is created on first open of any video
- Resume playback at the last stopped position when a video is reopened
- Keep `video_timings.stopped` up-to-date: on pause, on note creation, and on app exit
- Store the video duration in `video_timings.length` when first known

## User Stories

### US-001: Update migration schema
**Description:** As a developer, I need the database schema updated so that video timing state is stored in a dedicated table rather than a column on `videos`.

**Acceptance Criteria:**
- [ ] `videos.stop_time` column is removed from `001_create_videos_table.sql`
- [ ] `video_timings` table is added to `001_create_videos_table.sql` with columns: `id INTEGER PRIMARY KEY`, `video_id INTEGER NOT NULL REFERENCES videos(id)`, `stopped REAL` (nullable), `length REAL`
- [ ] A `UNIQUE` constraint exists on `video_timings.video_id` (enforces 1-to-1 relationship)
- [ ] `insert_video.sql` is updated to remove `stop_time` from the INSERT statement
- [ ] `db/models.go` `NoteVideo.StoppedAt` field is removed
- [ ] `db/functions.go` `getOrCreateVideo` no longer passes `StoppedAt` to the insert SQL
- [ ] `tui/tui.go` `newNoteVideo` helper no longer accepts or passes `stoppedAt`
- [ ] All callers of `newNoteVideo` are updated to remove the `stoppedAt` argument
- [ ] `CGO_ENABLED=0 go build ./...` passes with no errors
- [ ] `CGO_ENABLED=0 go vet ./...` passes with no errors

### US-002: Add VideoTiming DB model and SQL queries
**Description:** As a developer, I need Go structs and SQL queries to read and write `video_timings` rows.

**Acceptance Criteria:**
- [ ] `db/models.go` has a `VideoTiming` struct with fields: `ID int64`, `VideoID int64`, `Stopped *float64` (pointer, nullable), `Length float64`
- [ ] `db/sql/insert_video_timing.sql` inserts a new `video_timings` row: `INSERT INTO video_timings (video_id, stopped, length) VALUES (?, ?, ?)`
- [ ] `db/sql/select_video_timing_by_video.sql` selects a timing row by video_id: `SELECT id, video_id, stopped, length FROM video_timings WHERE video_id = ? LIMIT 1`
- [ ] `db/sql/upsert_video_timing_stopped.sql` inserts or updates `stopped` for a given `video_id`: uses `INSERT INTO video_timings (video_id, stopped, length) VALUES (?, ?, 0) ON CONFLICT(video_id) DO UPDATE SET stopped = excluded.stopped`
- [ ] `db/sql/update_video_timing_length.sql` updates `length` for a given `video_id`: `UPDATE video_timings SET length = ? WHERE video_id = ?`
- [ ] All new SQL files are embedded in `db/sql.go` (or equivalent embed file) following the existing pattern
- [ ] `db/functions.go` has `EnsureVideoTiming(db, videoID int64, length float64) (*VideoTiming, error)` — inserts a row if none exists, returns the existing or new row
- [ ] `db/functions.go` has `UpdateVideoTimingStopped(db, videoID int64, stopped float64) error` — upserts the `stopped` value for the given video
- [ ] `CGO_ENABLED=0 go build ./...` passes with no errors

### US-003: Ensure video row exists on open
**Description:** As the app, I need to register a video in the database the first time it is opened so that timing data can be associated with it.

**Acceptance Criteria:**
- [ ] `db/functions.go` has a public `EnsureVideo(db *sql.DB, path string, filesize int64, format string) (videoID int64, err error)` function that returns the existing video ID if a row for `path` already exists, or inserts a new row and returns the new ID
- [ ] In `cmd/root.go` `openCmd`, after mpv connects and the duration is fetched, `EnsureVideo` is called with `absPath`, `info.Size()`, and format (empty string if unknown)
- [ ] After `EnsureVideo`, `EnsureVideoTiming` is called with the returned `videoID` and the video duration (0 if duration fetch failed)
- [ ] The returned `*VideoTiming` is passed into `tui.Run` (or stored so the TUI can access it)
- [ ] `CGO_ENABLED=0 go build ./...` passes with no errors

### US-004: Resume playback from stopped position
**Description:** As a user, I want the video to resume at the position I last stopped at when I reopen it, so I don't have to manually seek back to where I was.

**Acceptance Criteria:**
- [ ] In `cmd/root.go` `openCmd`, after `EnsureVideoTiming` is called, if `VideoTiming.Stopped` is not nil and its value is greater than 0, `client.Seek(*timing.Stopped)` is called followed by `client.Pause()` before the TUI launches
- [ ] After seeking and pausing, the terminal prints `Resuming from <H:MM:SS>` using `timeutil.FormatTime` so the user knows the video was resumed
- [ ] If `VideoTiming.Stopped` is nil or 0, the video plays from the beginning as normal (no seek)
- [ ] `CGO_ENABLED=0 go build ./...` passes

### US-005: Update stopped position when video is paused
**Description:** As a user, I want the app to record my current position whenever I pause, so the correct resume point is always saved.

**Acceptance Criteria:**
- [ ] The `Model` in `tui/tui.go` stores `videoID int64` (set at model creation, passed in from the `VideoTiming`)
- [ ] When the pause toggle key (`p` / Space in `handleVideoKeys`) is invoked and the video transitions to paused, `db.UpdateVideoTimingStopped` is called with the current `time-pos` fetched from mpv
- [ ] When the `pause` command is executed in `handleCommand`, `db.UpdateVideoTimingStopped` is called with the current `time-pos`
- [ ] If fetching `time-pos` from mpv fails (e.g. not connected), the update is silently skipped
- [ ] `CGO_ENABLED=0 go build ./...` passes

### US-006: Update stopped position when a note is started
**Description:** As a user, I want the resume point to be saved whenever I start creating a note, clip, or tackle, since those actions mark a meaningful pause point in the analysis.

**Acceptance Criteria:**
- [ ] When `startNoteForm`, `startClipForm` (clip start), or `startTackleForm` is invoked in `tui/tui.go`, the current `time-pos` is fetched from mpv (it is already fetched as `timestamp` in the existing code)
- [ ] Immediately after fetching `timestamp`, `db.UpdateVideoTimingStopped(m.db, m.videoID, timestamp)` is called
- [ ] If the video is not connected to mpv at that point, the update is silently skipped
- [ ] `CGO_ENABLED=0 go build ./...` passes

### US-007: Update stopped position on app exit
**Description:** As a user, I want my exact position saved when I quit the app so I can resume accurately next time.

**Acceptance Criteria:**
- [ ] In the TUI quit handler (where `tea.Quit` is returned), `time-pos` is fetched from mpv via `client.GetTimePos()`
- [ ] `db.UpdateVideoTimingStopped(m.db, m.videoID, timePos)` is called before returning `tea.Quit`
- [ ] If mpv is not connected or the position cannot be fetched, the update is skipped and quit proceeds normally
- [ ] `CGO_ENABLED=0 go build ./...` passes

## Functional Requirements

- FR-1: The `videos` table must NOT have a `stop_time` column
- FR-2: A `video_timings` table must exist with a UNIQUE constraint on `video_id` (enforcing 1-to-1 with `videos`)
- FR-3: `video_timings.stopped` is nullable — NULL means "never stopped / play from beginning"
- FR-4: `video_timings.length` stores the total duration of the video in seconds
- FR-5: A `videos` row must be created the first time any video is opened, before the TUI starts
- FR-6: If `video_timings.stopped` is set (non-null, > 0) when a video is opened, mpv must be seeked to that position and paused before the TUI launches
- FR-7: `video_timings.stopped` must be updated on any of: video pause, note/clip/tackle start, app exit
- FR-8: All existing `newNoteVideo` / `getOrCreateVideo` calls that previously passed `stop_time` must be updated to not reference `StoppedAt`

## Non-Goals

- No UI in the TUI showing the resume position (beyond the terminal print in US-004)
- No manual "bookmark" feature — stopped position is automatic only
- No per-session history — only the single most recent stopped position is tracked
- No migration script — the `001` migration is modified directly since all data is test data

## Technical Considerations

- `video_timings.stopped` uses a `*float64` (pointer) in Go to represent SQL NULL
- The UNIQUE constraint on `video_timings.video_id` enables `ON CONFLICT` upsert syntax in SQLite
- `EnsureVideo` reuses the existing `select_video_by_path.sql` query and `insert_video.sql`; it does NOT go through `getOrCreateVideo` (which runs inside a transaction for note inserts)
- `EnsureVideoTiming` uses `SELECT` first, inserts only if no row found — returns the existing row so `stopped` can be checked
- `tui.Run` signature will need updating to accept `videoID int64` (or the full `*VideoTiming`) alongside the existing parameters
- The `select_note_videos_by_note.sql` scan of `StoppedAt` must be updated to reflect the schema change (or removed from the scan)

## Success Metrics

- A video opened a second time automatically seeks to the last pause point
- `video_timings.stopped` reflects the last paused position after any pause, note creation, or quit
- No regressions: note/clip/tackle creation continues to work as before

## Open Questions

- Should `video_timings.length` be updated after every open (in case the file changes)? Currently spec'd as write-once on first open.
- Should the `select_note_videos_by_note.sql` query continue to return a `StoppedAt`-equivalent from `video_timings`, or should callers stop expecting it?
