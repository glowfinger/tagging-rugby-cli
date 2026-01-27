package tui

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tagging-rugby-cli/mpv"
	"github.com/user/tagging-rugby-cli/tui/components"
)

const (
	// tickInterval is the interval for polling mpv status.
	tickInterval = 100 * time.Millisecond
	// defaultStepSize is the default seek step size in seconds.
	defaultStepSize = 1.0
	// resultDisplayDuration is how long to show command results.
	resultDisplayDuration = 3 * time.Second
)

// tickMsg is a message sent on every tick interval to update playback status.
type tickMsg time.Time

// clearResultMsg is sent to clear the command result message.
type clearResultMsg struct{}

// Model is the Bubbletea model for the TUI application.
// It implements the tea.Model interface with Init, Update, and View methods.
type Model struct {
	// mpv client for controlling video playback
	client *mpv.Client
	// database connection for notes, clips, and tackles
	db *sql.DB
	// current video file path
	videoPath string
	// error message to display (if any)
	err error
	// quitting flag to signal shutdown
	quitting bool
	// terminal width
	width int
	// terminal height
	height int
	// status bar state
	statusBar components.StatusBarState
	// notes list state
	notesList components.NotesListState
	// command input state
	commandInput components.CommandInputState
	// clip start timestamp (for clip start/end workflow)
	clipStartTimestamp float64
	// clipStartSet indicates if a clip start has been marked
	clipStartSet bool
}

// NewModel creates a new TUI model with the given mpv client, database connection, and video path.
func NewModel(client *mpv.Client, db *sql.DB, videoPath string) *Model {
	return &Model{
		client:    client,
		db:        db,
		videoPath: videoPath,
		statusBar: components.StatusBarState{
			StepSize: defaultStepSize,
		},
	}
}

// Init initializes the model. It returns an optional command to run.
func (m *Model) Init() tea.Cmd {
	// Start the ticker for polling mpv status
	return tickCmd()
}

// tickCmd returns a command that sends a tickMsg after the tick interval.
func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Update status bar from mpv
		m.updateStatusFromMpv()
		// Continue ticking
		return m, tickCmd()

	case clearResultMsg:
		// Clear the command result message
		m.commandInput.ClearResult()
		return m, nil

	case tea.KeyMsg:
		// Handle command mode input
		if m.commandInput.Active {
			return m.handleCommandInput(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case ":":
			// Enter command mode
			m.commandInput.Active = true
			m.commandInput.Input = ""
			m.commandInput.CursorPos = 0
			m.commandInput.ClearResult()
			return m, nil
		}
	}

	return m, nil
}

// handleCommandInput handles key events when in command mode.
func (m *Model) handleCommandInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		// Cancel command mode
		m.commandInput.Clear()
		return m, nil

	case "enter":
		// Execute command
		cmd := m.commandInput.GetCommand()
		if cmd != "" {
			result, err := m.executeCommand(cmd)
			if err != nil {
				m.commandInput.SetResult("Error: "+err.Error(), true)
			} else {
				m.commandInput.SetResult(result, false)
			}
			// Schedule clearing the result message
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}
		return m, nil

	case "backspace":
		m.commandInput.Backspace()
		return m, nil

	case "delete":
		m.commandInput.Delete()
		return m, nil

	case "left":
		m.commandInput.MoveCursorLeft()
		return m, nil

	case "right":
		m.commandInput.MoveCursorRight()
		return m, nil

	default:
		// Insert character if it's a printable rune
		if len(msg.String()) == 1 {
			m.commandInput.InsertChar(rune(msg.String()[0]))
		} else if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.commandInput.InsertChar(r)
			}
		}
		return m, nil
	}
}

// executeCommand parses and executes a command string.
// Returns a result message or an error.
func (m *Model) executeCommand(cmdStr string) (string, error) {
	// Parse command and arguments
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "note":
		return m.executeNoteCommand(args)
	case "clip":
		return m.executeClipCommand(args)
	case "tackle":
		return m.executeTackleCommand(args)
	case "pause", "p":
		if err := m.client.Pause(); err != nil {
			return "", err
		}
		return "Paused", nil
	case "play":
		if err := m.client.Play(); err != nil {
			return "", err
		}
		return "Playing", nil
	case "mute", "m":
		muted, err := m.client.GetMute()
		if err != nil {
			return "", err
		}
		if err := m.client.SetMute(!muted); err != nil {
			return "", err
		}
		if !muted {
			return "Muted", nil
		}
		return "Unmuted", nil
	case "seek":
		if len(args) < 1 {
			return "", fmt.Errorf("seek requires a time argument (e.g., seek 1:30 or seek 90)")
		}
		seconds, err := parseTimeToSeconds(args[0])
		if err != nil {
			return "", err
		}
		if err := m.client.Seek(seconds); err != nil {
			return "", err
		}
		return fmt.Sprintf("Seeked to %s", formatTimeString(seconds)), nil
	case "speed":
		if len(args) < 1 {
			speed, err := m.client.GetSpeed()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Speed: %.1fx", speed), nil
		}
		var speed float64
		if _, err := fmt.Sscanf(args[0], "%f", &speed); err != nil {
			return "", fmt.Errorf("invalid speed: %s", args[0])
		}
		if err := m.client.SetSpeed(speed); err != nil {
			return "", err
		}
		return fmt.Sprintf("Speed set to %.1fx", speed), nil
	case "q", "quit":
		m.quitting = true
		return "", nil
	case "help", "h":
		return "Commands: note add/list/goto, clip start/end/list/play/stop, tackle add/list, pause, play, mute, seek, speed, quit", nil
	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}

// executeNoteCommand handles note subcommands.
func (m *Model) executeNoteCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("note requires a subcommand: add, list, goto")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "add":
		if len(subargs) == 0 {
			return "", fmt.Errorf("note add requires text argument")
		}
		text := strings.Join(subargs, " ")
		return m.addNote(text, "", "", "")

	case "list":
		count, err := m.countNotes()
		if err != nil {
			return "", err
		}
		// Reload notes list
		m.loadNotesAndTackles()
		return fmt.Sprintf("%d note(s) for this video", count), nil

	case "goto":
		if len(subargs) == 0 {
			return "", fmt.Errorf("note goto requires note ID")
		}
		var noteID int64
		if _, err := fmt.Sscanf(subargs[0], "%d", &noteID); err != nil {
			return "", fmt.Errorf("invalid note ID: %s", subargs[0])
		}
		return m.gotoNote(noteID)

	default:
		return "", fmt.Errorf("unknown note subcommand: %s", subcmd)
	}
}

// executeClipCommand handles clip subcommands.
func (m *Model) executeClipCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("clip requires a subcommand: start, end, list, play, stop")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "start":
		timestamp, err := m.client.GetTimePos()
		if err != nil {
			return "", err
		}
		m.clipStartTimestamp = timestamp
		m.clipStartSet = true
		return fmt.Sprintf("Clip start marked at %s", formatTimeString(timestamp)), nil

	case "end":
		if !m.clipStartSet {
			return "", fmt.Errorf("no clip start marked. Use 'clip start' first")
		}
		endTimestamp, err := m.client.GetTimePos()
		if err != nil {
			return "", err
		}
		if m.clipStartTimestamp >= endTimestamp {
			return "", fmt.Errorf("clip end must be after start")
		}
		description := ""
		if len(subargs) > 0 {
			description = strings.Join(subargs, " ")
		}
		clipID, err := m.addClip(m.clipStartTimestamp, endTimestamp, description)
		if err != nil {
			return "", err
		}
		m.clipStartSet = false
		duration := endTimestamp - m.clipStartTimestamp
		return fmt.Sprintf("Clip %d saved (%.1fs)", clipID, duration), nil

	case "list":
		count, err := m.countClips()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d clip(s) for this video", count), nil

	case "play":
		if len(subargs) == 0 {
			return "", fmt.Errorf("clip play requires clip ID")
		}
		var clipID int64
		if _, err := fmt.Sscanf(subargs[0], "%d", &clipID); err != nil {
			return "", fmt.Errorf("invalid clip ID: %s", subargs[0])
		}
		return m.playClip(clipID)

	case "stop":
		if err := m.client.ClearABLoop(); err != nil {
			return "", err
		}
		return "A-B loop cleared", nil

	default:
		return "", fmt.Errorf("unknown clip subcommand: %s", subcmd)
	}
}

// executeTackleCommand handles tackle subcommands.
func (m *Model) executeTackleCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("tackle requires a subcommand: add, list")
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "add":
		// Parse flags from subargs
		player := ""
		team := ""
		attempt := 0
		outcome := ""
		for i := 0; i < len(subargs); i++ {
			switch subargs[i] {
			case "-p", "--player":
				if i+1 < len(subargs) {
					player = subargs[i+1]
					i++
				}
			case "-t", "--team":
				if i+1 < len(subargs) {
					team = subargs[i+1]
					i++
				}
			case "-a", "--attempt":
				if i+1 < len(subargs) {
					fmt.Sscanf(subargs[i+1], "%d", &attempt)
					i++
				}
			case "-o", "--outcome":
				if i+1 < len(subargs) {
					outcome = subargs[i+1]
					i++
				}
			}
		}
		if player == "" {
			return "", fmt.Errorf("tackle add requires --player")
		}
		if team == "" {
			return "", fmt.Errorf("tackle add requires --team")
		}
		if attempt == 0 {
			return "", fmt.Errorf("tackle add requires --attempt")
		}
		if outcome == "" {
			return "", fmt.Errorf("tackle add requires --outcome")
		}
		return m.addTackle(player, team, attempt, outcome)

	case "list":
		count, err := m.countTackles()
		if err != nil {
			return "", err
		}
		// Reload notes list (includes tackles)
		m.loadNotesAndTackles()
		return fmt.Sprintf("%d tackle(s) for this video", count), nil

	default:
		return "", fmt.Errorf("unknown tackle subcommand: %s", subcmd)
	}
}

// addNote adds a note at the current timestamp.
func (m *Model) addNote(text, category, player, team string) (string, error) {
	timestamp, err := m.client.GetTimePos()
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %w", err)
	}

	result, err := m.db.Exec(
		`INSERT INTO notes (video_path, timestamp_seconds, text, category, player, team) VALUES (?, ?, ?, ?, ?, ?)`,
		m.videoPath, timestamp, text, category, player, team,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert note: %w", err)
	}

	noteID, _ := result.LastInsertId()

	// Reload notes list
	m.loadNotesAndTackles()

	return fmt.Sprintf("Note %d added at %s", noteID, formatTimeString(timestamp)), nil
}

// countNotes counts notes for the current video.
func (m *Model) countNotes() (int, error) {
	var count int
	err := m.db.QueryRow(`SELECT COUNT(*) FROM notes WHERE video_path = ?`, m.videoPath).Scan(&count)
	return count, err
}

// gotoNote seeks to a note's timestamp.
func (m *Model) gotoNote(noteID int64) (string, error) {
	var timestamp float64
	var text sql.NullString
	err := m.db.QueryRow(
		`SELECT timestamp_seconds, text FROM notes WHERE id = ?`,
		noteID,
	).Scan(&timestamp, &text)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("note %d not found", noteID)
	}
	if err != nil {
		return "", err
	}

	if err := m.client.Seek(timestamp); err != nil {
		return "", err
	}

	textStr := ""
	if text.Valid {
		textStr = text.String
		if len(textStr) > 30 {
			textStr = textStr[:27] + "..."
		}
	}

	return fmt.Sprintf("Jumped to note %d: %s", noteID, textStr), nil
}

// addClip adds a clip to the database.
func (m *Model) addClip(start, end float64, description string) (int64, error) {
	result, err := m.db.Exec(
		`INSERT INTO clips (video_path, start_seconds, end_seconds, description) VALUES (?, ?, ?, ?)`,
		m.videoPath, start, end, description,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// countClips counts clips for the current video.
func (m *Model) countClips() (int, error) {
	var count int
	err := m.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE video_path = ?`, m.videoPath).Scan(&count)
	return count, err
}

// playClip seeks to a clip and sets A-B loop.
func (m *Model) playClip(clipID int64) (string, error) {
	var startSec, endSec float64
	var description sql.NullString
	err := m.db.QueryRow(
		`SELECT start_seconds, end_seconds, description FROM clips WHERE id = ?`,
		clipID,
	).Scan(&startSec, &endSec, &description)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("clip %d not found", clipID)
	}
	if err != nil {
		return "", err
	}

	if err := m.client.Seek(startSec); err != nil {
		return "", err
	}
	if err := m.client.SetABLoop(startSec, endSec); err != nil {
		return "", err
	}

	duration := endSec - startSec
	return fmt.Sprintf("Playing clip %d (%.1fs loop)", clipID, duration), nil
}

// addTackle adds a tackle at the current timestamp.
func (m *Model) addTackle(player, team string, attempt int, outcome string) (string, error) {
	// Validate outcome
	validOutcomes := map[string]bool{"missed": true, "completed": true, "possible": true, "other": true}
	if !validOutcomes[outcome] {
		return "", fmt.Errorf("invalid outcome '%s': must be missed, completed, possible, or other", outcome)
	}

	timestamp, err := m.client.GetTimePos()
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %w", err)
	}

	result, err := m.db.Exec(
		`INSERT INTO tackles (video_path, timestamp_seconds, player, team, attempt, outcome) VALUES (?, ?, ?, ?, ?, ?)`,
		m.videoPath, timestamp, player, team, attempt, outcome,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert tackle: %w", err)
	}

	tackleID, _ := result.LastInsertId()

	// Reload notes list
	m.loadNotesAndTackles()

	return fmt.Sprintf("Tackle %d recorded: %s %s", tackleID, player, outcome), nil
}

// countTackles counts tackles for the current video.
func (m *Model) countTackles() (int, error) {
	var count int
	err := m.db.QueryRow(`SELECT COUNT(*) FROM tackles WHERE video_path = ?`, m.videoPath).Scan(&count)
	return count, err
}

// parseTimeToSeconds parses a time string in MM:SS or seconds format.
func parseTimeToSeconds(timeStr string) (float64, error) {
	// Try MM:SS format first
	var minutes, seconds int
	if n, err := fmt.Sscanf(timeStr, "%d:%d", &minutes, &seconds); n == 2 && err == nil {
		return float64(minutes*60 + seconds), nil
	}

	// Try seconds format (float)
	var secs float64
	if n, err := fmt.Sscanf(timeStr, "%f", &secs); n == 1 && err == nil {
		return secs, nil
	}

	return 0, fmt.Errorf("expected MM:SS or seconds, got '%s'", timeStr)
}

// formatTimeString formats seconds as MM:SS.
func formatTimeString(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalSeconds := int(seconds)
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

// updateStatusFromMpv polls mpv for current playback status and updates the status bar.
func (m *Model) updateStatusFromMpv() {
	if m.client == nil || !m.client.IsConnected() {
		return
	}

	// Get pause state
	paused, err := m.client.GetPaused()
	if err == nil {
		m.statusBar.Paused = paused
	}

	// Get mute state
	muted, err := m.client.GetMute()
	if err == nil {
		m.statusBar.Muted = muted
	}

	// Get current position
	timePos, err := m.client.GetTimePos()
	if err == nil {
		m.statusBar.TimePos = timePos
	}

	// Get duration
	duration, err := m.client.GetDuration()
	if err == nil {
		m.statusBar.Duration = duration
	}
}

// View renders the current state of the model as a string.
func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return components.StatusBar(m.statusBar, m.width) + "\n\nError: " + m.err.Error() + "\n\nPress q to quit.\n"
	}

	// Render status bar at top
	statusBar := components.StatusBar(m.statusBar, m.width)

	// Calculate available height for notes list (minus status bar and command input)
	listHeight := m.height - 2 // 1 for status bar, 1 for command input
	if listHeight < 5 {
		listHeight = 5
	}

	// Render notes list
	notesList := components.NotesList(m.notesList, m.width, listHeight)

	// Render command input at bottom
	commandInput := components.CommandInput(m.commandInput, m.width)

	return statusBar + "\n" + notesList + "\n" + commandInput
}

// Run starts the Bubbletea program with the given model.
// It returns an error if the program fails to start or run.
func Run(client *mpv.Client, db *sql.DB, videoPath string) error {
	model := NewModel(client, db, videoPath)
	// Load notes and tackles for the current video
	model.loadNotesAndTackles()
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// loadNotesAndTackles loads notes and tackles from the database for the current video.
func (m *Model) loadNotesAndTackles() {
	if m.db == nil {
		return
	}

	var items []components.ListItem

	// Load notes
	rows, err := m.db.Query(`
		SELECT id, timestamp_seconds, text, category, player, team
		FROM notes
		WHERE video_path = ?
		ORDER BY timestamp_seconds ASC
	`, m.videoPath)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item components.ListItem
			var text, category, player, team sql.NullString
			err := rows.Scan(&item.ID, &item.TimestampSeconds, &text, &category, &player, &team)
			if err == nil {
				item.Type = components.ItemTypeNote
				item.Text = text.String
				item.Category = category.String
				item.Player = player.String
				item.Team = team.String
				items = append(items, item)
			}
		}
	}

	// Load tackles
	tackleRows, err := m.db.Query(`
		SELECT id, timestamp_seconds, player, team, outcome, notes, star
		FROM tackles
		WHERE video_path = ?
		ORDER BY timestamp_seconds ASC
	`, m.videoPath)
	if err == nil {
		defer tackleRows.Close()
		for tackleRows.Next() {
			var item components.ListItem
			var player, team, outcome, notes sql.NullString
			var star int
			err := tackleRows.Scan(&item.ID, &item.TimestampSeconds, &player, &team, &outcome, &notes, &star)
			if err == nil {
				item.Type = components.ItemTypeTackle
				item.Player = player.String
				item.Team = team.String
				item.Starred = star == 1
				// Build text from player, outcome, and notes
				if player.Valid && player.String != "" {
					item.Text = player.String
					if outcome.Valid && outcome.String != "" {
						item.Text += " - " + outcome.String
					}
				} else if outcome.Valid && outcome.String != "" {
					item.Text = outcome.String
				}
				if notes.Valid && notes.String != "" {
					if item.Text != "" {
						item.Text += ": " + notes.String
					} else {
						item.Text = notes.String
					}
				}
				items = append(items, item)
			}
		}
	}

	// Sort all items by timestamp
	sort.Slice(items, func(i, j int) bool {
		return items[i].TimestampSeconds < items[j].TimestampSeconds
	})

	m.notesList.Items = items
	m.notesList.SelectedIndex = 0
	m.notesList.ScrollOffset = 0
}
