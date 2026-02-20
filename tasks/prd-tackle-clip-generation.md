# PRD: Tackle Clip Generation

## Introduction

Automatically generate video clips using ffmpeg whenever a tackle note is created or updated. Clips are stored on disk relative to the source video, with a text overlay showing the timestamp, outcome, and attempt number for the first 3 seconds. A background worker polls the database for pending clips and processes them one at a time. Users can also trigger re-generation manually from the TUI.

## Goals

- Automatically queue a clip for generation whenever a note is saved or updated
- Process clips sequentially in a background goroutine without blocking the TUI
- Store lifecycle timestamps (started_at, finished_at, error_at) and error logs in `note_clips`
- Overlay timestamp, outcome, and attempt text (bottom-left, stacked) for the first 3 seconds
- Enforce a minimum clip duration of 4 seconds
- Organise output files in a predictable folder/filename structure relative to the source video
- Allow the user to manually trigger clip re-generation from the TUI

## User Stories

### US-001: Upsert a pending clip record on note save
**Description:** As a developer, I need a DB function that inserts or resets a `note_clips` row to `pending` status so the background worker can pick it up.

**Acceptance Criteria:**
- [ ] New SQL file `db/sql/upsert_note_clip_pending.sql` uses `INSERT OR REPLACE` (or equivalent upsert) to set `status='pending'`, `folder`, `filename`, `extension='mp4'`, `format='mp4'`, `filesize=0`, and clears `started_at`, `finished_at`, `error_at`, `log` for the given `note_id`
- [ ] `db/functions.go` has `UpsertNoteClipPending(db, noteID, folder, filename string) error`
- [ ] Folder and filename are computed from note data (category, player, timing start) before calling this function — see FR-3 and FR-4 for the exact format
- [ ] `InsertNoteWithChildren` calls `UpsertNoteClipPending` after committing, if the note has tackle and timing data
- [ ] `UpdateNoteWithChildren` calls `UpsertNoteClipPending` after committing, if the note has tackle and timing data
- [ ] Typecheck passes (`CGO_ENABLED=0`)

### US-002: Add SQL query to select the next pending clip with full context
**Description:** As a developer, I need a query that returns the next pending clip along with all the data the background worker needs to run ffmpeg.

**Acceptance Criteria:**
- [ ] New SQL file `db/sql/select_next_pending_clip.sql` selects one row from `note_clips` with `status='pending'`, joining `notes`, `note_timing`, `note_tackles`, and `videos` (via `notes.video_id`) to return: clip id, note_id, folder, filename, video path, note category, tackle player/attempt/outcome, timing start/end
- [ ] New Go struct `PendingClip` in `db/models.go` holds all joined fields
- [ ] New function `SelectNextPendingClip(db) (*PendingClip, error)` in `db/functions.go` — returns `nil, nil` when no pending clips exist
- [ ] Typecheck passes

### US-003: Background clip processor goroutine
**Description:** As a user, I want clips to be generated automatically in the background after I save a note, without the TUI freezing.

**Acceptance Criteria:**
- [ ] New file `clip/processor.go` with a `Processor` struct that holds a `*sql.DB` reference
- [ ] `Processor.Start(ctx context.Context)` launches a goroutine that loops: poll for a pending clip → process it → repeat; sleeps 2 seconds between polls when queue is empty
- [ ] When a pending clip is found:
  - [ ] Updates `note_clips` status to `processing`, sets `started_at = now()`
  - [ ] Creates the output directory (FR-5) if it does not exist
  - [ ] Builds and runs the ffmpeg command (FR-6, FR-7)
  - [ ] On ffmpeg exit code 0: sets `status='complete'`, `finished_at = now()`, updates `filesize` from the output file on disk
  - [ ] On ffmpeg non-zero exit or error: sets `status='error'`, `error_at = now()`, writes combined stdout+stderr to `log`
- [ ] The processor is started from `main.go` (or the TUI init) passing the app's `context.Context` so it shuts down cleanly when the app exits
- [ ] Typecheck passes

### US-004: New DB function to update clip status to processing/complete/error
**Description:** As a developer, I need fine-grained DB update functions so the processor can set individual lifecycle fields atomically.

**Acceptance Criteria:**
- [ ] `db/functions.go` has `MarkClipProcessing(db, clipID int64, startedAt time.Time) error` — sets `status='processing'`, `started_at`
- [ ] `db/functions.go` has `MarkClipComplete(db, clipID int64, finishedAt time.Time, filesize int64) error` — sets `status='complete'`, `finished_at`, `filesize`
- [ ] `db/functions.go` has `MarkClipError(db, clipID int64, errorAt time.Time, log string) error` — sets `status='error'`, `error_at`, `log`
- [ ] Corresponding SQL files added in `db/sql/`
- [ ] Typecheck passes

### US-005: Manual re-generate keybinding in TUI
**Description:** As a user, I want to press a key on a selected note to force clip re-generation, even if the clip already exists.

**Acceptance Criteria:**
- [ ] Keybinding `Ctrl+R` in `FocusNotes` mode calls a new `startRegenerateClip(noteID)` method in `tui/tui.go`
- [ ] `startRegenerateClip` loads the note's tackle and timing data, computes folder/filename, calls `UpsertNoteClipPending`, then deletes the existing clip file from disk if it exists at `{folder}/{filename}`
- [ ] A brief status message is shown in the TUI confirming the clip was queued (e.g. "Clip queued for regeneration")
- [ ] Keybinding is guarded by existing form/command mode checks (no action when a form is open)
- [ ] Typecheck passes

## Functional Requirements

- **FR-1:** On every note save (insert or update), if the note has at least one `note_tackles` row and one `note_timing` row, upsert a `note_clips` row with `status='pending'`.
- **FR-2:** Clip duration = `max(4.0, note_timing.end - note_timing.start)` seconds, measured from `note_timing.start`.
- **FR-3:** Output folder path = `{video_dir}/clips/{note.category}/{note_tackle.player}/` where `{video_dir}` is the directory portion of the source video's path (e.g. `filepath.Dir(video.path)`).
- **FR-4:** Output filename = `{HHMMSS}-{player}-{category}-{outcome}-{attempt}.mp4` where `{HHMMSS}` is `note_timing.start` formatted as zero-padded hours, minutes, seconds (e.g. `003045` for 30m 45s). All spaces in player/category/outcome are replaced with underscores; all characters are lowercased.
- **FR-5:** The output directory must be created (including all parents) before ffmpeg is invoked. Use `os.MkdirAll`.
- **FR-6:** The ffmpeg command must be:
  ```
  ffmpeg -y -i {video_path} -ss {start_seconds} -t {duration_seconds}
    -vf "drawtext=text='{HH\:MM\:SS}':x=10:y=h-th-60:fontsize=28:fontcolor=white:enable='lt(t,3)',
         drawtext=text='{outcome}':x=10:y=h-th-30:fontsize=28:fontcolor=white:enable='lt(t,3)',
         drawtext=text='Attempt {attempt}':x=10:y=h-th:fontsize=28:fontcolor=white:enable='lt(t,3)'"
    {output_path}
  ```
  where `{HH\:MM\:SS}` is the human-readable timestamp from `note_timing.start` (e.g. `0\:30\:45`), colons escaped for ffmpeg's drawtext filter.
- **FR-7:** The text overlay is stacked vertically at the bottom-left of the frame (10px from the left, working up from the bottom edge) and is only visible for the first 3 seconds of the clip.
- **FR-8:** If a clip file already exists at the output path when re-generating, it must be deleted before ffmpeg runs (ffmpeg `-y` flag also handles overwrite, but explicit deletion ensures a clean state).
- **FR-9:** The background processor must not run more than one ffmpeg process at a time (sequential, not parallel).
- **FR-10:** When the app's context is cancelled (user quits), the processor goroutine must exit cleanly; any in-progress ffmpeg process should be allowed to complete (not killed).

## Non-Goals

- No clip preview inside the TUI (clips are only viewable externally)
- No parallel clip processing (one at a time only)
- No progress bar or percentage display for ffmpeg encoding
- No automatic cleanup of old clip files when a note is deleted
- No support for notes without a tackle record or timing record (skip silently)
- No transcoding format options — output is always `.mp4`

## Technical Considerations

- `note_clips` table already exists in the schema with all required columns (`status`, `started_at`, `finished_at`, `error_at`, `log`, `folder`, `filename`, `extension`, `format`, `filesize`)
- `NoteClip`, `InsertNoteClip`, `UpdateNoteClip`, `SelectNoteClipsByNote` already exist — new functions should follow the same patterns in `db/functions.go`
- The `clip/` package is a new top-level package. It imports `db` but must not import `tui` (avoid circular imports)
- Use `exec.CommandContext` with the app context so the goroutine respects cancellation
- `note_timing.start` and `note_timing.end` are stored as `REAL` (float64 seconds); use `timeutil.FormatTime` for HH:MM:SS display and hand-roll the `HHMMSS` filename format (e.g. `fmt.Sprintf("%02d%02d%02d", h, m, s)`)
- ffmpeg must be present on the system PATH; if not found, log an error to the clip's `log` column and set `status='error'`
- Build with `CGO_ENABLED=0` (modernc.org/sqlite requirement)

## Success Metrics

- A clip file is created on disk within a reasonable time after each note save
- The `note_clips` row reflects accurate `started_at`, `finished_at` or `error_at` timestamps
- Clip playback starts at the correct video timestamp and runs for the correct duration
- Text overlay is visible at the bottom-left for exactly the first 3 seconds
- No TUI freezing or slowdown during clip generation

## Open Questions

- Should the processor log to a file or only to the `note_clips.log` column? (Assume column only for now)
- Is there a maximum number of retries if ffmpeg fails, or does the clip stay in `error` state until manually re-queued? (Assume stays in `error` — user re-queues with Ctrl+R)
