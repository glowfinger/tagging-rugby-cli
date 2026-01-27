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

// stepSizes defines the available step sizes for seek operations.
// Users can cycle through these with < and > keys.
var stepSizes = []float64{0.1, 0.5, 1, 2, 5, 10, 30}

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
	// showHelp indicates if the help overlay is visible
	showHelp bool
	// statsView holds the state for the stats view
	statsView components.StatsViewState
	// overlayEnabled indicates if the mpv overlay is enabled
	overlayEnabled bool
	// noteInput holds the state for the quick note input
	noteInput components.NoteInputState
	// tackleInput holds the state for the quick tackle input
	tackleInput components.TackleInputState
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
		// Update overlay if enabled
		if m.overlayEnabled {
			m.updateOverlay()
		}
		// Continue ticking
		return m, tickCmd()

	case clearResultMsg:
		// Clear the command result message
		m.commandInput.ClearResult()
		return m, nil

	case tea.KeyMsg:
		// Handle help overlay - any key dismisses it
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Handle stats view input
		if m.statsView.Active {
			return m.handleStatsViewInput(msg)
		}

		// Handle note input mode
		if m.noteInput.Active {
			return m.handleNoteInput(msg)
		}

		// Handle tackle input mode
		if m.tackleInput.Active {
			return m.handleTackleInput(msg)
		}

		// Handle command mode input
		if m.commandInput.Active {
			return m.handleCommandInput(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "?":
			// Toggle help overlay
			m.showHelp = true
			return m, nil
		case "s", "S":
			// Open stats view
			m.loadTackleStats()
			m.statsView.Active = true
			return m, nil
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
		case " ":
			// Space toggles play/pause
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.TogglePause()
			}
			return m, nil
		case "m", "M":
			// M toggles mute
			if m.client != nil && m.client.IsConnected() {
				muted, err := m.client.GetMute()
				if err == nil {
					_ = m.client.SetMute(!muted)
				}
			}
			return m, nil
		case "h", "H":
			// H steps backward by current step size
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.SeekRelative(-m.statusBar.StepSize)
			}
			return m, nil
		case "l", "L":
			// L steps forward by current step size
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.SeekRelative(m.statusBar.StepSize)
			}
			return m, nil
		case "<":
			// Decrease step size
			m.decreaseStepSize()
			return m, nil
		case ">":
			// Increase step size
			m.increaseStepSize()
			return m, nil
		case "j", "J":
			// J moves selection to previous note/tackle in list
			m.notesList.MoveUp()
			return m, nil
		case "k", "K":
			// K moves selection to next note/tackle in list
			m.notesList.MoveDown()
			return m, nil
		case "enter":
			// Enter on selected item seeks mpv to that timestamp
			return m.jumpToSelectedItem()
		case "o", "O":
			// O toggles overlay on/off
			m.overlayEnabled = !m.overlayEnabled
			m.statusBar.OverlayEnabled = m.overlayEnabled
			if !m.overlayEnabled {
				// Hide overlay when disabled
				if m.client != nil && m.client.IsConnected() {
					_ = m.client.HideOverlay(1)
				}
			}
			return m, nil
		case "n", "N":
			// N opens quick note input prompt
			return m.openNoteInput()
		case "t", "T":
			// T opens quick tackle input prompt
			return m.openTackleInput()
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
				return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
					return clearResultMsg{}
				})
			}
			// Handle special return values that open input prompts
			if result == "OPEN_NOTE_INPUT" {
				m.commandInput.Clear()
				return m.openNoteInput()
			}
			if result == "OPEN_TACKLE_INPUT" {
				m.commandInput.Clear()
				return m.openTackleInput()
			}
			m.commandInput.SetResult(result, false)
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

// openNoteInput opens the quick note input prompt.
func (m *Model) openNoteInput() (tea.Model, tea.Cmd) {
	if m.client == nil || !m.client.IsConnected() {
		m.commandInput.SetResult("Not connected to mpv", true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Get current timestamp from mpv
	timestamp, err := m.client.GetTimePos()
	if err != nil {
		m.commandInput.SetResult("Failed to get timestamp: "+err.Error(), true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Initialize note input state
	m.noteInput.Clear()
	m.noteInput.Active = true
	m.noteInput.Timestamp = timestamp
	m.noteInput.CurrentField = components.NoteInputFieldText

	return m, nil
}

// handleNoteInput handles key events when in note input mode.
func (m *Model) handleNoteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		// Cancel note input
		m.noteInput.Clear()
		return m, nil

	case "enter":
		// Save note if text is not empty
		text, category, player, team, timestamp := m.noteInput.GetNote()
		if text == "" {
			m.commandInput.SetResult("Note text cannot be empty", true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		// Save note to database
		result, err := m.db.Exec(
			`INSERT INTO notes (video_path, timestamp_seconds, text, category, player, team) VALUES (?, ?, ?, ?, ?, ?)`,
			m.videoPath, timestamp, text, category, player, team,
		)
		if err != nil {
			m.noteInput.Clear()
			m.commandInput.SetResult("Error: "+err.Error(), true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		noteID, _ := result.LastInsertId()

		// Clear note input and reload list
		m.noteInput.Clear()
		m.loadNotesAndTackles()

		// Show confirmation
		m.commandInput.SetResult(fmt.Sprintf("Note %d added at %s", noteID, formatTimeString(timestamp)), false)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})

	case "tab":
		// Move to next field
		m.noteInput.NextField()
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.noteInput.PrevField()
		return m, nil

	case "backspace":
		// Delete last character
		m.noteInput.Backspace()
		return m, nil

	default:
		// Insert character if it's a printable rune
		if len(msg.String()) == 1 {
			m.noteInput.InsertChar(rune(msg.String()[0]))
		} else if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.noteInput.InsertChar(r)
			}
		}
		return m, nil
	}
}

// openTackleInput opens the quick tackle input prompt.
func (m *Model) openTackleInput() (tea.Model, tea.Cmd) {
	if m.client == nil || !m.client.IsConnected() {
		m.commandInput.SetResult("Not connected to mpv", true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Get current timestamp from mpv
	timestamp, err := m.client.GetTimePos()
	if err != nil {
		m.commandInput.SetResult("Failed to get timestamp: "+err.Error(), true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Initialize tackle input state
	m.tackleInput.Clear()
	m.tackleInput.Active = true
	m.tackleInput.Timestamp = timestamp
	m.tackleInput.CurrentField = components.TackleInputFieldPlayer

	return m, nil
}

// handleTackleInput handles key events when in tackle input mode.
func (m *Model) handleTackleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		// Cancel tackle input
		m.tackleInput.Clear()
		return m, nil

	case "enter":
		// Validate required fields
		if errMsg := m.tackleInput.ValidationError(); errMsg != "" {
			m.commandInput.SetResult(errMsg, true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		// Save tackle to database
		player, team, attemptStr, outcome, followed, notes, zone, star, timestamp := m.tackleInput.GetTackle()

		// Parse attempt as integer
		var attempt int
		fmt.Sscanf(attemptStr, "%d", &attempt)

		// Convert star bool to int
		starInt := 0
		if star {
			starInt = 1
		}

		result, err := m.db.Exec(
			`INSERT INTO tackles (video_path, timestamp_seconds, player, team, attempt, outcome, followed, notes, zone, star) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			m.videoPath, timestamp, player, team, attempt, outcome, followed, notes, zone, starInt,
		)
		if err != nil {
			m.tackleInput.Clear()
			m.commandInput.SetResult("Error: "+err.Error(), true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		tackleID, _ := result.LastInsertId()

		// Clear tackle input and reload list
		m.tackleInput.Clear()
		m.loadNotesAndTackles()

		// Show confirmation
		starSymbol := ""
		if star {
			starSymbol = " ★"
		}
		m.commandInput.SetResult(fmt.Sprintf("Tackle %d recorded: %s %s%s", tackleID, player, outcome, starSymbol), false)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})

	case "tab":
		// Move to next field
		m.tackleInput.NextField()
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.tackleInput.PrevField()
		return m, nil

	case "*", "s", "S":
		// Toggle star (when not in a text field or when * is pressed)
		if msg.String() == "*" || m.tackleInput.CurrentField == components.TackleInputFieldOutcome {
			m.tackleInput.ToggleStar()
			return m, nil
		}
		// For s/S in text fields, insert the character
		m.tackleInput.InsertChar(rune(msg.String()[0]))
		return m, nil

	case "left":
		// For outcome field, cycle to previous outcome
		if m.tackleInput.CurrentField == components.TackleInputFieldOutcome {
			m.tackleInput.PrevOutcome()
		}
		return m, nil

	case "right":
		// For outcome field, cycle to next outcome
		if m.tackleInput.CurrentField == components.TackleInputFieldOutcome {
			m.tackleInput.NextOutcome()
		}
		return m, nil

	case "backspace":
		// Delete last character
		m.tackleInput.Backspace()
		return m, nil

	default:
		// Insert character if it's a printable rune (except in outcome field)
		if m.tackleInput.CurrentField != components.TackleInputFieldOutcome {
			if len(msg.String()) == 1 {
				m.tackleInput.InsertChar(rune(msg.String()[0]))
			} else if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					m.tackleInput.InsertChar(r)
				}
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
	// Shorthand commands
	case "nn":
		return m.executeShorthandNoteCommand(args)
	case "nt":
		return m.executeShorthandTackleCommand(args)
	case "cs":
		return m.executeClipCommand([]string{"start"})
	case "ce":
		// Shorthand for clip end - args become the description
		return m.executeClipCommand(append([]string{"end"}, args...))
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

// executeShorthandNoteCommand handles the :nn shorthand command.
// With no args, it opens the quick note input prompt (same as N key).
// With args, it adds a note with the given text.
func (m *Model) executeShorthandNoteCommand(args []string) (string, error) {
	if len(args) == 0 {
		// No args - open quick note input prompt
		// Return special value to signal opening the input
		return "OPEN_NOTE_INPUT", nil
	}
	// With args - add note with the text
	text := strings.Join(args, " ")
	return m.addNote(text, "", "", "")
}

// executeShorthandTackleCommand handles the :nt shorthand command.
// With no args, it opens the quick tackle input prompt (same as T key).
// With 4 positional args, it adds a tackle: :nt <player> <team> <attempt> <outcome>
// With partial args, it shows a usage hint.
func (m *Model) executeShorthandTackleCommand(args []string) (string, error) {
	if len(args) == 0 {
		// No args - open quick tackle input prompt
		return "OPEN_TACKLE_INPUT", nil
	}
	if len(args) != 4 {
		return "", fmt.Errorf("usage: :nt <player> <team> <attempt> <outcome>")
	}
	// Parse positional args
	player := args[0]
	team := args[1]
	var attempt int
	if _, err := fmt.Sscanf(args[2], "%d", &attempt); err != nil || attempt < 1 {
		return "", fmt.Errorf("invalid attempt number: %s (must be a positive integer)", args[2])
	}
	outcome := args[3]
	return m.addTackle(player, team, attempt, outcome)
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

// jumpToSelectedItem seeks mpv to the selected item's timestamp and displays details.
func (m *Model) jumpToSelectedItem() (tea.Model, tea.Cmd) {
	item := m.notesList.GetSelectedItem()
	if item == nil {
		m.commandInput.SetResult("No item selected", true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	if m.client == nil || !m.client.IsConnected() {
		m.commandInput.SetResult("Not connected to mpv", true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Seek to the item's timestamp
	if err := m.client.Seek(item.TimestampSeconds); err != nil {
		m.commandInput.SetResult("Error: "+err.Error(), true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Build details message
	var typeStr string
	if item.Type == components.ItemTypeNote {
		typeStr = "note"
	} else {
		typeStr = "tackle"
	}

	// Build info string
	var info string
	if item.Text != "" {
		info = item.Text
		if len(info) > 40 {
			info = info[:37] + "..."
		}
	}
	if item.Player != "" && item.Type == components.ItemTypeTackle {
		if info != "" {
			info = item.Player + ": " + info
		} else {
			info = item.Player
		}
	}
	if item.Category != "" && item.Type == components.ItemTypeNote {
		if info != "" {
			info = "[" + item.Category + "] " + info
		} else {
			info = "[" + item.Category + "]"
		}
	}

	starStr := ""
	if item.Starred {
		starStr = " ★"
	}

	result := fmt.Sprintf("Jumped to %s %d%s: %s", typeStr, item.ID, starStr, info)
	m.commandInput.SetResult(result, false)
	return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

// decreaseStepSize cycles to the previous (smaller) step size.
func (m *Model) decreaseStepSize() {
	currentIndex := m.findStepSizeIndex()
	if currentIndex > 0 {
		m.statusBar.StepSize = stepSizes[currentIndex-1]
	}
}

// increaseStepSize cycles to the next (larger) step size.
func (m *Model) increaseStepSize() {
	currentIndex := m.findStepSizeIndex()
	if currentIndex < len(stepSizes)-1 {
		m.statusBar.StepSize = stepSizes[currentIndex+1]
	}
}

// findStepSizeIndex finds the index of the current step size in the stepSizes array.
// If the current step size is not in the array, it returns the index of the closest value.
func (m *Model) findStepSizeIndex() int {
	for i, size := range stepSizes {
		if m.statusBar.StepSize == size {
			return i
		}
	}
	// Find closest if not exact match
	for i, size := range stepSizes {
		if m.statusBar.StepSize < size {
			if i == 0 {
				return 0
			}
			return i - 1
		}
	}
	return len(stepSizes) - 1
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

// overlayProximitySeconds is how close (in seconds) a note must be to current timestamp to display.
const overlayProximitySeconds = 2.0

// overlayID is the ID used for the notes overlay in mpv.
const overlayID = 1

// updateOverlay displays notes near the current timestamp on the mpv video.
func (m *Model) updateOverlay() {
	if m.client == nil || !m.client.IsConnected() {
		return
	}

	// Get current playback position
	timePos := m.statusBar.TimePos

	// Find notes within proximity of current timestamp
	var nearbyNotes []components.ListItem
	for _, item := range m.notesList.Items {
		// Only show notes (not tackles) in overlay
		if item.Type != components.ItemTypeNote {
			continue
		}
		// Check if note is within proximity
		diff := timePos - item.TimestampSeconds
		if diff >= 0 && diff <= overlayProximitySeconds {
			nearbyNotes = append(nearbyNotes, item)
		}
	}

	// If no notes nearby, hide overlay
	if len(nearbyNotes) == 0 {
		_ = m.client.HideOverlay(overlayID)
		return
	}

	// Build overlay text with ASS formatting for semi-transparent background
	// ASS format: {\pos(x,y)\an7\1c&HFFFFFF&\3c&H000000&\bord2\shad0\alpha&H40&}text
	// Using position at bottom-left with some margin, anchor point 7 (bottom-left)
	var overlayText strings.Builder
	for _, note := range nearbyNotes {
		// Build note display: category, player/team, text
		var parts []string
		if note.Category != "" {
			parts = append(parts, "["+note.Category+"]")
		}
		if note.Player != "" || note.Team != "" {
			playerTeam := ""
			if note.Player != "" && note.Team != "" {
				playerTeam = note.Player + " (" + note.Team + ")"
			} else if note.Player != "" {
				playerTeam = note.Player
			} else {
				playerTeam = note.Team
			}
			parts = append(parts, playerTeam)
		}
		if note.Text != "" {
			parts = append(parts, note.Text)
		}

		noteDisplay := strings.Join(parts, " - ")
		if noteDisplay == "" {
			noteDisplay = "(empty note)"
		}

		// ASS styling: position at bottom, semi-transparent box background
		// \an1 = bottom-left alignment
		// \pos(20, h-80) = position 20px from left, 80px from bottom (we'll use percent)
		// \bord0 = no border
		// \shad0 = no shadow
		// \3c&H000000& = box color (black)
		// \4c&H000000& = shadow color (black)
		// \4a&H80& = shadow/box alpha (semi-transparent)
		// \1c&HFFFFFF& = primary fill color (white)
		// Using simple format with box enabled via \be1 (blur edges) and \bord
		overlayText.WriteString(fmt.Sprintf("{\\an7\\pos(20,20)\\fs24\\1c&HFFFFFF&\\3c&H201a1a&\\bord3\\shad0}%s\\N", noteDisplay))
	}

	// Show the overlay
	_ = m.client.ShowOverlay(overlayID, overlayText.String())
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

	// If help overlay is active, show it instead of normal view
	if m.showHelp {
		return components.HelpOverlay(m.width, m.height)
	}

	// If stats view is active, show it instead of normal view
	if m.statsView.Active {
		return components.StatsView(m.statsView, m.width, m.height)
	}

	// Render status bar at top
	statusBar := components.StatusBar(m.statusBar, m.width)

	// Check if note input is active
	if m.noteInput.Active {
		// Show note input overlay instead of normal view
		noteInput := components.NoteInput(m.noteInput, m.width, m.noteInput.Timestamp)
		return statusBar + "\n" + noteInput
	}

	// Check if tackle input is active
	if m.tackleInput.Active {
		// Show tackle input overlay instead of normal view
		tackleInput := components.TackleInput(m.tackleInput, m.width, m.tackleInput.Timestamp)
		return statusBar + "\n" + tackleInput
	}

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

// handleStatsViewInput handles key events when the stats view is active.
func (m *Model) handleStatsViewInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode input first
	if m.statsView.FilterMode {
		return m.handleStatsFilterInput(msg)
	}

	switch msg.String() {
	case "backspace":
		// Return to main view
		m.statsView.Active = false
		return m, nil
	case "tab":
		// Cycle sort column
		m.statsView.NextSortColumn()
		return m, nil
	case "v", "V":
		// Toggle between current video / all videos
		m.statsView.AllVideos = !m.statsView.AllVideos
		m.loadTackleStats()
		return m, nil
	case "j", "J":
		// Move selection up
		m.statsView.MoveUp()
		return m, nil
	case "k", "K":
		// Move selection down
		m.statsView.MoveDown()
		return m, nil
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "?":
		// Show help overlay
		m.showHelp = true
		return m, nil
	case "/":
		// Enter filter mode
		m.statsView.FilterMode = true
		m.statsView.FilterInput = ""
		return m, nil
	case "escape":
		// Clear all filters
		m.statsView.ClearFilters()
		return m, nil
	}
	return m, nil
}

// handleStatsFilterInput handles key events when in filter input mode.
func (m *Model) handleStatsFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		// Exit filter mode (but keep existing filters)
		m.statsView.FilterMode = false
		m.statsView.FilterInput = ""
		return m, nil
	case "enter":
		// Apply filter and exit filter mode
		if m.statsView.FilterInput != "" {
			m.statsView.ToggleFilter(m.statsView.FilterInput)
		}
		m.statsView.FilterMode = false
		m.statsView.FilterInput = ""
		return m, nil
	case "backspace":
		// Delete last character
		if len(m.statsView.FilterInput) > 0 {
			m.statsView.FilterInput = m.statsView.FilterInput[:len(m.statsView.FilterInput)-1]
		}
		return m, nil
	default:
		// Add character to filter input
		if len(msg.String()) == 1 {
			m.statsView.FilterInput += msg.String()
		} else if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.statsView.FilterInput += string(r)
			}
		}
		return m, nil
	}
}

// loadTackleStats loads tackle statistics from the database.
func (m *Model) loadTackleStats() {
	if m.db == nil {
		return
	}

	// Build query based on whether we want all videos or just current video
	var query string
	var args []interface{}

	if m.statsView.AllVideos {
		query = `
			SELECT
				player,
				COUNT(*) as total,
				SUM(CASE WHEN outcome = 'completed' THEN 1 ELSE 0 END) as completed,
				SUM(CASE WHEN outcome = 'missed' THEN 1 ELSE 0 END) as missed,
				SUM(CASE WHEN outcome = 'possible' THEN 1 ELSE 0 END) as possible,
				SUM(CASE WHEN outcome = 'other' THEN 1 ELSE 0 END) as other,
				SUM(CASE WHEN star = 1 THEN 1 ELSE 0 END) as starred
			FROM tackles
			GROUP BY player
		`
	} else {
		query = `
			SELECT
				player,
				COUNT(*) as total,
				SUM(CASE WHEN outcome = 'completed' THEN 1 ELSE 0 END) as completed,
				SUM(CASE WHEN outcome = 'missed' THEN 1 ELSE 0 END) as missed,
				SUM(CASE WHEN outcome = 'possible' THEN 1 ELSE 0 END) as possible,
				SUM(CASE WHEN outcome = 'other' THEN 1 ELSE 0 END) as other,
				SUM(CASE WHEN star = 1 THEN 1 ELSE 0 END) as starred
			FROM tackles
			WHERE video_path = ?
			GROUP BY player
		`
		args = append(args, m.videoPath)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	var stats []components.PlayerStats
	for rows.Next() {
		var stat components.PlayerStats
		err := rows.Scan(
			&stat.Player,
			&stat.Total,
			&stat.Completed,
			&stat.Missed,
			&stat.Possible,
			&stat.Other,
			&stat.Starred,
		)
		if err == nil {
			// Calculate completion percentage
			if stat.Completed+stat.Missed > 0 {
				stat.Percentage = float64(stat.Completed) / float64(stat.Completed+stat.Missed) * 100
			}
			stats = append(stats, stat)
		}
	}

	m.statsView.Stats = stats
	m.statsView.SelectedIndex = 0
	m.statsView.ScrollOffset = 0
	m.statsView.SortStats()
}
