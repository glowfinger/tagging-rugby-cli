package db

import (
	_ "embed"
)

// Schema and migrations

//go:embed sql/create_tables.sql
var CreateTablesSQL string

//go:embed sql/seed_categories.sql
var SeedCategoriesSQL string

// Notes queries

//go:embed sql/insert_note.sql
var InsertNoteSQL string

//go:embed sql/select_notes_by_video.sql
var SelectNotesByVideoSQL string

//go:embed sql/select_note_id.sql
var SelectNoteIDSQL string

//go:embed sql/select_note_details.sql
var SelectNoteDetailsSQL string

//go:embed sql/select_note_brief.sql
var SelectNoteBriefSQL string

//go:embed sql/delete_note.sql
var DeleteNoteSQL string

//go:embed sql/select_notes_for_tui.sql
var SelectNotesForTUISQL string

//go:embed sql/count_notes_by_video.sql
var CountNotesByVideoSQL string

// Clips queries

//go:embed sql/insert_clip.sql
var InsertClipSQL string

//go:embed sql/insert_clip_basic.sql
var InsertClipBasicSQL string

//go:embed sql/select_clip_play.sql
var SelectClipPlaySQL string

//go:embed sql/select_clip_export.sql
var SelectClipExportSQL string

//go:embed sql/select_clips_by_video_for_export.sql
var SelectClipsByVideoForExportSQL string

//go:embed sql/select_clips_by_video.sql
var SelectClipsByVideoSQL string

//go:embed sql/count_clips_by_video.sql
var CountClipsByVideoSQL string

// Categories queries

//go:embed sql/select_categories.sql
var SelectCategoriesSQL string

//go:embed sql/insert_category.sql
var InsertCategorySQL string

//go:embed sql/delete_category.sql
var DeleteCategorySQL string

// Tackles queries

//go:embed sql/insert_tackle.sql
var InsertTackleSQL string

//go:embed sql/insert_tackle_basic.sql
var InsertTackleBasicSQL string

//go:embed sql/insert_tackle_with_extras.sql
var InsertTackleWithExtrasSQL string

//go:embed sql/select_tackles_by_video.sql
var SelectTacklesByVideoSQL string

//go:embed sql/select_tackle_counts.sql
var SelectTackleCountsSQL string

//go:embed sql/select_tackle_details.sql
var SelectTackleDetailsSQL string

//go:embed sql/select_tackles_for_tui.sql
var SelectTacklesForTUISQL string

//go:embed sql/count_tackles_by_video.sql
var CountTacklesByVideoSQL string

// Tackle stats queries

//go:embed sql/select_tackle_stats_all.sql
var SelectTackleStatsAllSQL string

//go:embed sql/select_tackle_stats_by_video.sql
var SelectTackleStatsByVideoSQL string
