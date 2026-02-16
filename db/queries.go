package db

import (
	_ "embed"
)

// Schema and migrations

//go:embed sql/create_tables.sql
var CreateTablesSQL string

// Notes queries

//go:embed sql/insert_note.sql
var InsertNoteSQL string

//go:embed sql/select_notes.sql
var SelectNotesSQL string

//go:embed sql/select_note_by_id.sql
var SelectNoteByIDSQL string

//go:embed sql/delete_note.sql
var DeleteNoteSQL string

// Video queries

//go:embed sql/insert_video.sql
var InsertVideoSQL string

//go:embed sql/select_video_by_id.sql
var SelectVideoByIDSQL string

//go:embed sql/select_video_by_path.sql
var SelectVideoByPathSQL string

//go:embed sql/update_video_stop_time.sql
var UpdateVideoStopTimeSQL string

// Note child table insert queries

//go:embed sql/insert_note_clip.sql
var InsertNoteClipSQL string

//go:embed sql/insert_note_timing.sql
var InsertNoteTimingSQL string

//go:embed sql/insert_note_tackle.sql
var InsertNoteTackleSQL string

//go:embed sql/insert_note_zone.sql
var InsertNoteZoneSQL string

//go:embed sql/insert_note_detail.sql
var InsertNoteDetailSQL string

//go:embed sql/insert_note_highlight.sql
var InsertNoteHighlightSQL string

// Note child table select queries

//go:embed sql/select_note_clips_by_note.sql
var SelectNoteClipsByNoteSQL string

//go:embed sql/select_note_timing_by_note.sql
var SelectNoteTimingByNoteSQL string

//go:embed sql/select_note_tackles_by_note.sql
var SelectNoteTacklesByNoteSQL string

//go:embed sql/select_note_zones_by_note.sql
var SelectNoteZonesByNoteSQL string

//go:embed sql/select_note_details_by_note.sql
var SelectNoteDetailsByNoteSQL string

//go:embed sql/select_note_highlights_by_note.sql
var SelectNoteHighlightsByNoteSQL string

// Note child table delete queries

//go:embed sql/delete_note_details.sql
var DeleteNoteDetailsSQL string

//go:embed sql/delete_note_zones.sql
var DeleteNoteZonesSQL string

//go:embed sql/delete_note_highlights.sql
var DeleteNoteHighlightsSQL string

//go:embed sql/delete_note_tackles.sql
var DeleteNoteTacklesSQL string

// Note child table update queries

//go:embed sql/update_note_timing.sql
var UpdateNoteTimingSQL string

// Joined queries for TUI views

//go:embed sql/select_notes_with_timing.sql
var SelectNotesWithTimingSQL string

//go:embed sql/select_notes_with_video.sql
var SelectNotesWithVideoSQL string

//go:embed sql/select_tackle_stats.sql
var SelectTackleStatsSQL string

