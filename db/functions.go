package db

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/tagging-rugby-cli/clip"
)

// EnsureVideoTiming selects the video_timing row for the given videoID; inserts one (with stopped=NULL) if not found.
// If length > 0, it is always written to the row (whether new or existing) so the duration stays current.
func EnsureVideoTiming(db *sql.DB, videoID int64, length float64) (*VideoTiming, error) {
	var vt VideoTiming
	err := db.QueryRow(SelectVideoTimingByVideoSQL, videoID).Scan(&vt.ID, &vt.VideoID, &vt.Stopped, &vt.Length)
	if err == nil {
		if length > 0 && vt.Length != length {
			db.Exec(UpdateVideoTimingLengthSQL, length, videoID)
			vt.Length = length
		}
		return &vt, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("select video timing: %w", err)
	}
	result, err := db.Exec(InsertVideoTimingSQL, videoID, nil, length)
	if err != nil {
		return nil, fmt.Errorf("insert video timing: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get video timing id: %w", err)
	}
	return &VideoTiming{ID: id, VideoID: videoID, Stopped: nil, Length: length}, nil
}

// UpdateVideoTimingStopped upserts a video_timings row setting stopped to the given value.
func UpdateVideoTimingStopped(db *sql.DB, videoID int64, stopped float64) error {
	_, err := db.Exec(UpsertVideoTimingStoppedSQL, videoID, stopped)
	if err != nil {
		return fmt.Errorf("upsert video timing stopped: %w", err)
	}
	return nil
}

// EnsureVideo returns the existing video ID for the given path, or inserts a new row and returns its ID.
func EnsureVideo(db *sql.DB, path string, filesize int64, format string) (int64, error) {
	var videoID int64
	err := db.QueryRow(SelectVideoByPathSQL, path).Scan(&videoID)
	if err == nil {
		return videoID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("select video by path: %w", err)
	}
	base := filepath.Base(path)
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	result, err := db.Exec(InsertVideoSQL, path, base, ext, format, filesize)
	if err != nil {
		return 0, fmt.Errorf("insert video: %w", err)
	}
	return result.LastInsertId()
}

// InsertNote inserts a new note with the given video_id and returns its ID.
func InsertNote(db *sql.DB, category string, videoID int64) (int64, error) {
	result, err := db.Exec(InsertNoteSQL, category, videoID)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}
	return result.LastInsertId()
}

// getOrCreateVideo looks up a video by path within a transaction; inserts it if not found.
// Returns the video ID.
func getOrCreateVideo(tx *sql.Tx, v NoteVideo) (int64, error) {
	var videoID int64
	err := tx.QueryRow(SelectVideoByPathSQL, v.Path).Scan(&videoID)
	if err == nil {
		return videoID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("query video by path: %w", err)
	}
	base := filepath.Base(v.Path)
	ext := strings.TrimPrefix(filepath.Ext(v.Path), ".")
	result, err := tx.Exec(InsertVideoSQL, v.Path, base, ext, v.Format, v.Size)
	if err != nil {
		return 0, fmt.Errorf("insert video: %w", err)
	}
	return result.LastInsertId()
}

// MarkClipProcessing updates a note_clips row to processing status with the given start time.
func MarkClipProcessing(db *sql.DB, clipID int64, startedAt time.Time) error {
	_, err := db.Exec(MarkClipProcessingSQL, startedAt, clipID)
	if err != nil {
		return fmt.Errorf("mark clip processing: %w", err)
	}
	return nil
}

// MarkClipComplete updates a note_clips row to complete status with the given finish time and filesize.
func MarkClipComplete(db *sql.DB, clipID int64, finishedAt time.Time, filesize int64) error {
	_, err := db.Exec(MarkClipCompleteSQL, finishedAt, filesize, clipID)
	if err != nil {
		return fmt.Errorf("mark clip complete: %w", err)
	}
	return nil
}

// MarkClipError updates a note_clips row to error status with the given error time and log message.
func MarkClipError(db *sql.DB, clipID int64, errorAt time.Time, logMsg string) error {
	_, err := db.Exec(MarkClipErrorSQL, errorAt, logMsg, clipID)
	if err != nil {
		return fmt.Errorf("mark clip error: %w", err)
	}
	return nil
}

// UpsertNoteClipPending inserts or resets a note_clips row to pending status so the background worker can pick it up.
func UpsertNoteClipPending(db *sql.DB, noteID int64, folder, filename string) error {
	_, err := db.Exec(UpsertNoteClipPendingSQL, noteID, folder, filename)
	if err != nil {
		return fmt.Errorf("upsert note clip pending: %w", err)
	}
	return nil
}

// InsertNoteClip inserts a note_clips row.
func InsertNoteClip(db *sql.DB, noteID int64, folder, filename, extension, format string, filesize int64, status string, startedAt, finishedAt, errorAt interface{}, log string) error {
	_, err := db.Exec(InsertNoteClipSQL, noteID, folder, filename, extension, format, filesize, status, startedAt, finishedAt, errorAt, log)
	if err != nil {
		return fmt.Errorf("insert note clip: %w", err)
	}
	return nil
}

// SelectNoteClipByID returns a single note_clips row by ID.
func SelectNoteClipByID(database *sql.DB, id int64) (*NoteClip, error) {
	var c NoteClip
	err := database.QueryRow(SelectNoteClipByIDSQL, id).Scan(&c.ID, &c.NoteID, &c.Folder, &c.Filename, &c.Extension, &c.Format, &c.Filesize, &c.Status, &c.StartedAt, &c.FinishedAt, &c.ErrorAt, &c.Log)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateNoteClip updates the lifecycle fields of a note_clips row.
func UpdateNoteClip(database *sql.DB, id int64, status string, startedAt, finishedAt, errorAt interface{}, log string) error {
	_, err := database.Exec(UpdateNoteClipSQL, status, startedAt, finishedAt, errorAt, log, id)
	if err != nil {
		return fmt.Errorf("update note clip: %w", err)
	}
	return nil
}

// InsertNoteTiming inserts a note_timing row.
func InsertNoteTiming(db *sql.DB, noteID int64, start, end float64) error {
	_, err := db.Exec(InsertNoteTimingSQL, noteID, start, end)
	if err != nil {
		return fmt.Errorf("insert note timing: %w", err)
	}
	return nil
}

// InsertNoteTackle inserts a note_tackles row.
func InsertNoteTackle(db *sql.DB, noteID int64, player string, attempt int, outcome string) error {
	_, err := db.Exec(InsertNoteTackleSQL, noteID, player, attempt, outcome)
	if err != nil {
		return fmt.Errorf("insert note tackle: %w", err)
	}
	return nil
}

// InsertNoteZone inserts a note_zones row.
func InsertNoteZone(db *sql.DB, noteID int64, horizontal, vertical string) error {
	_, err := db.Exec(InsertNoteZoneSQL, noteID, horizontal, vertical)
	if err != nil {
		return fmt.Errorf("insert note zone: %w", err)
	}
	return nil
}

// InsertNoteDetail inserts a note_details row.
func InsertNoteDetail(db *sql.DB, noteID int64, detailType, note string) error {
	_, err := db.Exec(InsertNoteDetailSQL, noteID, detailType, note)
	if err != nil {
		return fmt.Errorf("insert note detail: %w", err)
	}
	return nil
}

// InsertNoteHighlight inserts a note_highlights row.
func InsertNoteHighlight(db *sql.DB, noteID int64, highlightType string) error {
	_, err := db.Exec(InsertNoteHighlightSQL, noteID, highlightType)
	if err != nil {
		return fmt.Errorf("insert note highlight: %w", err)
	}
	return nil
}

// QueueClipIfNeeded checks if the note has all required data (category, timing, tackle) and queues a clip
// generation job by upserting a pending note_clips row. Silently returns nil if any data is missing.
func QueueClipIfNeeded(database *sql.DB, noteID int64, videoPath string) error {
	note, err := SelectNoteByID(database, noteID)
	if err != nil {
		return nil
	}

	timings, err := SelectNoteTimingByNote(database, noteID)
	if err != nil || len(timings) == 0 {
		return nil
	}

	tackles, err := SelectNoteTacklesByNote(database, noteID)
	if err != nil || len(tackles) == 0 {
		return nil
	}

	folder, filename := clip.ClipPaths(videoPath, note.Category, tackles[0].Player, tackles[0].Attempt, tackles[0].Outcome, timings[0].Start)
	return UpsertNoteClipPending(database, noteID, folder, filename)
}

// InsertNoteWithChildren inserts a note and its related child records in a transaction.
// It accepts the note category plus optional child records to insert.
type NoteChildren struct {
	Videos     []NoteVideo
	Clips      []NoteClip
	Timings    []NoteTiming
	Tackles    []NoteTackle
	Zones      []NoteZone
	Details    []NoteDetail
	Highlights []NoteHighlight
}

func InsertNoteWithChildren(database *sql.DB, category string, children NoteChildren) (int64, error) {
	tx, err := database.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get or create the video record and resolve its ID.
	var videoID int64
	if len(children.Videos) > 0 {
		id, err := getOrCreateVideo(tx, children.Videos[0])
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		videoID = id
	}

	// Insert parent note with video_id.
	result, err := tx.Exec(InsertNoteSQL, category, videoID)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}
	noteID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get note id: %w", err)
	}
	for _, c := range children.Clips {
		if _, err := tx.Exec(InsertNoteClipSQL, noteID, c.Folder, c.Filename, c.Extension, c.Format, c.Filesize, c.Status, c.StartedAt, c.FinishedAt, c.ErrorAt, c.Log); err != nil {
			return 0, fmt.Errorf("insert note clip: %w", err)
		}
	}
	for _, t := range children.Timings {
		if _, err := tx.Exec(InsertNoteTimingSQL, noteID, t.Start, t.End); err != nil {
			return 0, fmt.Errorf("insert note timing: %w", err)
		}
	}
	for _, t := range children.Tackles {
		if _, err := tx.Exec(InsertNoteTackleSQL, noteID, t.Player, t.Attempt, t.Outcome); err != nil {
			return 0, fmt.Errorf("insert note tackle: %w", err)
		}
	}
	for _, z := range children.Zones {
		if _, err := tx.Exec(InsertNoteZoneSQL, noteID, z.Horizontal, z.Vertical); err != nil {
			return 0, fmt.Errorf("insert note zone: %w", err)
		}
	}
	for _, d := range children.Details {
		if _, err := tx.Exec(InsertNoteDetailSQL, noteID, d.Type, d.Note); err != nil {
			return 0, fmt.Errorf("insert note detail: %w", err)
		}
	}
	for _, h := range children.Highlights {
		if _, err := tx.Exec(InsertNoteHighlightSQL, noteID, h.Type); err != nil {
			return 0, fmt.Errorf("insert note highlight: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	if len(children.Videos) > 0 {
		if err := QueueClipIfNeeded(database, noteID, children.Videos[0].Path); err != nil {
			log.Printf("queue clip after insert: %v", err)
		}
	}

	return noteID, nil
}

// UpdateNoteWithChildren deletes existing child rows and re-inserts from the provided NoteChildren struct in a transaction.
func UpdateNoteWithChildren(database *sql.DB, noteID int64, children NoteChildren) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing child rows
	if _, err := tx.Exec(DeleteNoteDetailsSQL, noteID); err != nil {
		return fmt.Errorf("delete note details: %w", err)
	}
	if _, err := tx.Exec(DeleteNoteZonesSQL, noteID); err != nil {
		return fmt.Errorf("delete note zones: %w", err)
	}
	if _, err := tx.Exec(DeleteNoteHighlightsSQL, noteID); err != nil {
		return fmt.Errorf("delete note highlights: %w", err)
	}
	if _, err := tx.Exec(DeleteNoteTacklesSQL, noteID); err != nil {
		return fmt.Errorf("delete note tackles: %w", err)
	}

	// Re-insert child records
	for _, t := range children.Tackles {
		if _, err := tx.Exec(InsertNoteTackleSQL, noteID, t.Player, t.Attempt, t.Outcome); err != nil {
			return fmt.Errorf("insert note tackle: %w", err)
		}
	}
	for _, z := range children.Zones {
		if _, err := tx.Exec(InsertNoteZoneSQL, noteID, z.Horizontal, z.Vertical); err != nil {
			return fmt.Errorf("insert note zone: %w", err)
		}
	}
	for _, d := range children.Details {
		if _, err := tx.Exec(InsertNoteDetailSQL, noteID, d.Type, d.Note); err != nil {
			return fmt.Errorf("insert note detail: %w", err)
		}
	}
	for _, h := range children.Highlights {
		if _, err := tx.Exec(InsertNoteHighlightSQL, noteID, h.Type); err != nil {
			return fmt.Errorf("insert note highlight: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	videos, err := SelectNoteVideosByNote(database, noteID)
	if err == nil && len(videos) > 0 {
		if err := QueueClipIfNeeded(database, noteID, videos[0].Path); err != nil {
			log.Printf("queue clip after update: %v", err)
		}
	}

	return nil
}

// UpdateNoteTiming updates the timing record for a given note.
func UpdateNoteTiming(database *sql.DB, noteID int64, start, end float64) error {
	result, err := database.Exec(UpdateNoteTimingSQL, start, end, noteID)
	if err != nil {
		return fmt.Errorf("update note timing: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SelectNoteByID returns a single note by ID.
func SelectNoteByID(database *sql.DB, id int64) (*Note, error) {
	var n Note
	err := database.QueryRow(SelectNoteByIDSQL, id).Scan(&n.ID, &n.Category, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// SelectNotes returns all notes ordered by created_at DESC.
func SelectNotes(database *sql.DB) ([]Note, error) {
	rows, err := database.Query(SelectNotesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Category, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// SelectNoteVideosByNote returns all videos for a given note.
func SelectNoteVideosByNote(database *sql.DB, noteID int64) ([]NoteVideo, error) {
	rows, err := database.Query(SelectNoteVideosByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []NoteVideo
	for rows.Next() {
		var v NoteVideo
		if err := rows.Scan(&v.ID, &v.NoteID, &v.Path, &v.Size, &v.Duration, &v.Format); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

// SelectNoteClipsByNote returns all clips for a given note.
func SelectNoteClipsByNote(database *sql.DB, noteID int64) ([]NoteClip, error) {
	rows, err := database.Query(SelectNoteClipsByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clips []NoteClip
	for rows.Next() {
		var c NoteClip
		if err := rows.Scan(&c.ID, &c.NoteID, &c.Folder, &c.Filename, &c.Extension, &c.Format, &c.Filesize, &c.Status, &c.StartedAt, &c.FinishedAt, &c.ErrorAt, &c.Log); err != nil {
			return nil, err
		}
		clips = append(clips, c)
	}
	return clips, rows.Err()
}

// SelectNoteTimingByNote returns all timing records for a given note.
func SelectNoteTimingByNote(database *sql.DB, noteID int64) ([]NoteTiming, error) {
	rows, err := database.Query(SelectNoteTimingByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timings []NoteTiming
	for rows.Next() {
		var t NoteTiming
		if err := rows.Scan(&t.ID, &t.NoteID, &t.Start, &t.End); err != nil {
			return nil, err
		}
		timings = append(timings, t)
	}
	return timings, rows.Err()
}

// SelectNoteTacklesByNote returns all tackles for a given note.
func SelectNoteTacklesByNote(database *sql.DB, noteID int64) ([]NoteTackle, error) {
	rows, err := database.Query(SelectNoteTacklesByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tackles []NoteTackle
	for rows.Next() {
		var t NoteTackle
		if err := rows.Scan(&t.ID, &t.NoteID, &t.Player, &t.Attempt, &t.Outcome); err != nil {
			return nil, err
		}
		tackles = append(tackles, t)
	}
	return tackles, rows.Err()
}

// SelectNoteZonesByNote returns all zones for a given note.
func SelectNoteZonesByNote(database *sql.DB, noteID int64) ([]NoteZone, error) {
	rows, err := database.Query(SelectNoteZonesByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []NoteZone
	for rows.Next() {
		var z NoteZone
		if err := rows.Scan(&z.ID, &z.NoteID, &z.Horizontal, &z.Vertical); err != nil {
			return nil, err
		}
		zones = append(zones, z)
	}
	return zones, rows.Err()
}

// SelectNoteDetailsByNote returns all details for a given note.
func SelectNoteDetailsByNote(database *sql.DB, noteID int64) ([]NoteDetail, error) {
	rows, err := database.Query(SelectNoteDetailsByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []NoteDetail
	for rows.Next() {
		var d NoteDetail
		if err := rows.Scan(&d.ID, &d.NoteID, &d.Type, &d.Note); err != nil {
			return nil, err
		}
		details = append(details, d)
	}
	return details, rows.Err()
}

// SelectNoteHighlightsByNote returns all highlights for a given note.
func SelectNoteHighlightsByNote(database *sql.DB, noteID int64) ([]NoteHighlight, error) {
	rows, err := database.Query(SelectNoteHighlightsByNoteSQL, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var highlights []NoteHighlight
	for rows.Next() {
		var h NoteHighlight
		if err := rows.Scan(&h.ID, &h.NoteID, &h.Type); err != nil {
			return nil, err
		}
		highlights = append(highlights, h)
	}
	return highlights, rows.Err()
}

// EditTackleData holds all the data needed to populate an edit tackle form.
type EditTackleData struct {
	Player     string
	Attempt    int
	Outcome    string
	Followed   string
	Notes      string
	Zone       string
	Star       bool
	Timestamp  float64
	EndSeconds float64
}

// LoadNoteForEdit loads all tackle-related data for a note to populate an edit form.
// Returns the tackle fields, timing (as timestamp + endSeconds), details, zone, and star highlight.
func LoadNoteForEdit(database *sql.DB, noteID int64) (*EditTackleData, error) {
	data := &EditTackleData{}

	// Load tackle data
	tackles, err := SelectNoteTacklesByNote(database, noteID)
	if err != nil {
		return nil, fmt.Errorf("load tackles: %w", err)
	}
	if len(tackles) > 0 {
		data.Player = tackles[0].Player
		data.Attempt = tackles[0].Attempt
		data.Outcome = tackles[0].Outcome
	}

	// Load timing data
	timings, err := SelectNoteTimingByNote(database, noteID)
	if err != nil {
		return nil, fmt.Errorf("load timing: %w", err)
	}
	if len(timings) > 0 {
		data.Timestamp = timings[0].Start
		endSecs := timings[0].End - timings[0].Start
		if endSecs <= 0 {
			endSecs = 2.0
		}
		data.EndSeconds = endSecs
	} else {
		data.EndSeconds = 2.0
	}

	// Load details (followed, notes)
	details, err := SelectNoteDetailsByNote(database, noteID)
	if err != nil {
		return nil, fmt.Errorf("load details: %w", err)
	}
	for _, d := range details {
		switch d.Type {
		case "followed":
			data.Followed = d.Note
		case "notes":
			data.Notes = d.Note
		}
	}

	// Load zone
	zones, err := SelectNoteZonesByNote(database, noteID)
	if err != nil {
		return nil, fmt.Errorf("load zones: %w", err)
	}
	if len(zones) > 0 {
		data.Zone = zones[0].Horizontal
	}

	// Load highlights (star)
	highlights, err := SelectNoteHighlightsByNote(database, noteID)
	if err != nil {
		return nil, fmt.Errorf("load highlights: %w", err)
	}
	for _, h := range highlights {
		if h.Type == "star" {
			data.Star = true
		}
	}

	return data, nil
}

// DeleteNote deletes a note by ID. Cascade handles child records.
func DeleteNote(database *sql.DB, id int64) error {
	result, err := database.Exec(DeleteNoteSQL, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
