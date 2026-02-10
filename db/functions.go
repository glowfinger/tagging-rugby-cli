package db

import (
	"database/sql"
	"fmt"
)

// InsertNote inserts a new note and returns its ID.
func InsertNote(db *sql.DB, category string) (int64, error) {
	result, err := db.Exec(InsertNoteSQL, category)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}
	return result.LastInsertId()
}

// InsertNoteVideo inserts a note_videos row.
func InsertNoteVideo(db *sql.DB, noteID int64, path string, size int64, duration float64, format string, stoppedAt float64) error {
	_, err := db.Exec(InsertNoteVideoSQL, noteID, path, size, duration, format, stoppedAt)
	if err != nil {
		return fmt.Errorf("insert note video: %w", err)
	}
	return nil
}

// InsertNoteClip inserts a note_clips row.
func InsertNoteClip(db *sql.DB, noteID int64, name string, duration float64, startedAt, finishedAt, errorAt interface{}, clipError string) error {
	_, err := db.Exec(InsertNoteClipSQL, noteID, name, duration, startedAt, finishedAt, errorAt, clipError)
	if err != nil {
		return fmt.Errorf("insert note clip: %w", err)
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

	// Insert parent note
	result, err := tx.Exec(InsertNoteSQL, category)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}
	noteID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get note id: %w", err)
	}

	// Insert child records
	for _, v := range children.Videos {
		if _, err := tx.Exec(InsertNoteVideoSQL, noteID, v.Path, v.Size, v.Duration, v.Format, v.StoppedAt); err != nil {
			return 0, fmt.Errorf("insert note video: %w", err)
		}
	}
	for _, c := range children.Clips {
		if _, err := tx.Exec(InsertNoteClipSQL, noteID, c.Name, c.Duration, c.StartedAt, c.FinishedAt, c.ErrorAt, c.Error); err != nil {
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
		if err := rows.Scan(&v.ID, &v.NoteID, &v.Path, &v.Size, &v.Duration, &v.Format, &v.StoppedAt); err != nil {
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
		if err := rows.Scan(&c.ID, &c.NoteID, &c.Name, &c.Duration, &c.StartedAt, &c.FinishedAt, &c.ErrorAt, &c.Error); err != nil {
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
