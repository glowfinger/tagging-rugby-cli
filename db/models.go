package db

import "time"

// Note represents a row in the notes table.
type Note struct {
	ID        int64
	Category  string
	CreatedAt time.Time
}

// NoteVideo represents a row in the note_videos table.
type NoteVideo struct {
	ID        int64
	NoteID    int64
	Path      string
	Size      int64
	Duration  float64
	Format    string
	StoppedAt float64
}

// NoteClip represents a row in the note_clips table.
type NoteClip struct {
	ID         int64
	NoteID     int64
	Name       string
	Duration   float64
	StartedAt  *time.Time
	FinishedAt *time.Time
	ErrorAt    *time.Time
	Error      string
}

// NoteTiming represents a row in the note_timing table.
type NoteTiming struct {
	ID     int64
	NoteID int64
	Start  float64
	End    float64
}

// NoteTackle represents a row in the note_tackles table.
type NoteTackle struct {
	ID      int64
	NoteID  int64
	Player  string
	Attempt int
	Outcome string
}

// NoteZone represents a row in the note_zones table.
type NoteZone struct {
	ID         int64
	NoteID     int64
	Horizontal string
	Vertical   string
}

// NoteDetail represents a row in the note_details table.
type NoteDetail struct {
	ID     int64
	NoteID int64
	Type   string
	Note   string
}

// NoteHighlight represents a row in the note_highlights table.
type NoteHighlight struct {
	ID     int64
	NoteID int64
	Type   string
}

// ExportedClipRow represents a joined row for viewing exported clips.
type ExportedClipRow struct {
	NoteID   int64
	FileName string
	Player   string
	Category string
	Outcome  string
	Duration float64
	Status   string
	Error    string
}

// TackleClipRow represents a joined row for tackle clip export.
type TackleClipRow struct {
	NoteID       int64
	Category     string
	Start        float64
	End          float64
	VideoPath    string
	Player       string
	Outcome      string
	ClipFinishedAt *time.Time
}
