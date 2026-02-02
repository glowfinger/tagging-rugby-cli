package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/mpv"
	"github.com/user/tagging-rugby-cli/tui/components"
	"github.com/user/tagging-rugby-cli/tui/styles"
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
		// Refresh stats for column 3 periodically (every tick is fine, query is fast)
		m.loadTackleStatsForPanel()
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
	case "esc":
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
	case "esc":
		// Cancel note input
		m.noteInput.Clear()
		return m, nil

	case "enter":
		// Save note if text is not empty
		text, category, _, _, timestamp := m.noteInput.GetNote()
		if text == "" {
			m.commandInput.SetResult("Note text cannot be empty", true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		// Get video duration for video child record
		duration, _ := m.client.GetDuration()

		// Build children
		children := db.NoteChildren{
			Timings: []db.NoteTiming{
				{Start: timestamp, End: timestamp},
			},
			Videos: []db.NoteVideo{
				{Path: m.videoPath, Duration: duration, StoppedAt: timestamp},
			},
			Details: []db.NoteDetail{
				{Type: "text", Note: text},
			},
		}

		// Use category from input, default to empty
		if category == "" {
			category = "note"
		}

		// Save note with children
		noteID, err := db.InsertNoteWithChildren(m.db, category, children)
		if err != nil {
			m.noteInput.Clear()
			m.commandInput.SetResult("Error: "+err.Error(), true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

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
	case "esc":
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
		player, _, attemptStr, outcome, _, notes, _, star, timestamp := m.tackleInput.GetTackle()

		// Parse attempt as integer
		var attempt int
		fmt.Sscanf(attemptStr, "%d", &attempt)

		// Get video duration for video child record
		duration, _ := m.client.GetDuration()

		// Build children
		children := db.NoteChildren{
			Timings: []db.NoteTiming{
				{Start: timestamp, End: timestamp},
			},
			Videos: []db.NoteVideo{
				{Path: m.videoPath, Duration: duration, StoppedAt: timestamp},
			},
			Tackles: []db.NoteTackle{
				{Player: player, Attempt: attempt, Outcome: outcome},
			},
		}

		// Add detail if notes were provided
		if notes != "" {
			children.Details = []db.NoteDetail{
				{Type: "text", Note: notes},
			}
		}

		// Add highlight if starred
		if star {
			children.Highlights = []db.NoteHighlight{
				{Type: "star"},
			}
		}

		noteID, err := db.InsertNoteWithChildren(m.db, "tackle", children)
		if err != nil {
			m.tackleInput.Clear()
			m.commandInput.SetResult("Error: "+err.Error(), true)
			return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
				return clearResultMsg{}
			})
		}

		// Clear tackle input and reload list
		m.tackleInput.Clear()
		m.loadNotesAndTackles()

		// Show confirmation
		starSymbol := ""
		if star {
			starSymbol = " ★"
		}
		m.commandInput.SetResult(fmt.Sprintf("Tackle %d recorded: %s %s%s", noteID, player, outcome, starSymbol), false)
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
func (m *Model) addNote(text, category, _, _ string) (string, error) {
	timestamp, err := m.client.GetTimePos()
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %w", err)
	}

	duration, _ := m.client.GetDuration()

	children := db.NoteChildren{
		Timings: []db.NoteTiming{
			{Start: timestamp, End: timestamp},
		},
		Videos: []db.NoteVideo{
			{Path: m.videoPath, Duration: duration, StoppedAt: timestamp},
		},
	}

	if text != "" {
		children.Details = []db.NoteDetail{
			{Type: "text", Note: text},
		}
	}

	if category == "" {
		category = "note"
	}

	noteID, err := db.InsertNoteWithChildren(m.db, category, children)
	if err != nil {
		return "", fmt.Errorf("failed to insert note: %w", err)
	}

	// Reload notes list
	m.loadNotesAndTackles()

	return fmt.Sprintf("Note %d added at %s", noteID, formatTimeString(timestamp)), nil
}

// countNotes counts notes for the current video.
func (m *Model) countNotes() (int, error) {
	rows, err := m.db.Query(db.SelectNotesWithVideoSQL, m.videoPath)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

// gotoNote seeks to a note's timestamp.
func (m *Model) gotoNote(noteID int64) (string, error) {
	// Check note exists
	note, err := db.SelectNoteByID(m.db, noteID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("note %d not found", noteID)
	}
	if err != nil {
		return "", err
	}

	// Get timing for the note
	timings, err := db.SelectNoteTimingByNote(m.db, noteID)
	if err != nil || len(timings) == 0 {
		return "", fmt.Errorf("note %d has no timing data", noteID)
	}

	timestamp := timings[0].Start
	if err := m.client.Seek(timestamp); err != nil {
		return "", err
	}

	// Get detail text if available
	details, _ := db.SelectNoteDetailsByNote(m.db, noteID)
	textStr := ""
	if len(details) > 0 {
		textStr = details[0].Note
		if len(textStr) > 30 {
			textStr = textStr[:27] + "..."
		}
	}

	return fmt.Sprintf("Jumped to note %d [%s]: %s", note.ID, note.Category, textStr), nil
}

// addClip adds a clip to the database.
func (m *Model) addClip(start, end float64, description string) (int64, error) {
	clipDuration := end - start

	children := db.NoteChildren{
		Timings: []db.NoteTiming{
			{Start: start, End: end},
		},
		Videos: []db.NoteVideo{
			{Path: m.videoPath, StoppedAt: start},
		},
		Clips: []db.NoteClip{
			{Name: description, Duration: clipDuration},
		},
	}

	return db.InsertNoteWithChildren(m.db, "clip", children)
}

// countClips counts clip notes for the current video.
func (m *Model) countClips() (int, error) {
	rows, err := m.db.Query(
		"SELECT n.id FROM notes n INNER JOIN note_videos nv ON nv.note_id = n.id WHERE nv.path = ? AND n.category = 'clip'",
		m.videoPath,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

// playClip seeks to a clip note and sets A-B loop using its timing.
func (m *Model) playClip(noteID int64) (string, error) {
	// Check note exists
	_, err := db.SelectNoteByID(m.db, noteID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("note %d not found", noteID)
	}
	if err != nil {
		return "", err
	}

	// Get timing for the clip
	timings, err := db.SelectNoteTimingByNote(m.db, noteID)
	if err != nil || len(timings) == 0 {
		return "", fmt.Errorf("note %d has no timing data", noteID)
	}

	startSec := timings[0].Start
	endSec := timings[0].End

	if err := m.client.Seek(startSec); err != nil {
		return "", err
	}
	if err := m.client.SetABLoop(startSec, endSec); err != nil {
		return "", err
	}

	duration := endSec - startSec
	return fmt.Sprintf("Playing clip %d (%.1fs loop)", noteID, duration), nil
}

// addTackle adds a tackle at the current timestamp.
func (m *Model) addTackle(player, _ string, attempt int, outcome string) (string, error) {
	// Validate outcome
	validOutcomes := map[string]bool{"missed": true, "completed": true, "possible": true, "other": true}
	if !validOutcomes[outcome] {
		return "", fmt.Errorf("invalid outcome '%s': must be missed, completed, possible, or other", outcome)
	}

	timestamp, err := m.client.GetTimePos()
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %w", err)
	}

	duration, _ := m.client.GetDuration()

	children := db.NoteChildren{
		Timings: []db.NoteTiming{
			{Start: timestamp, End: timestamp},
		},
		Videos: []db.NoteVideo{
			{Path: m.videoPath, Duration: duration, StoppedAt: timestamp},
		},
		Tackles: []db.NoteTackle{
			{Player: player, Attempt: attempt, Outcome: outcome},
		},
	}

	noteID, err := db.InsertNoteWithChildren(m.db, "tackle", children)
	if err != nil {
		return "", fmt.Errorf("failed to insert tackle: %w", err)
	}

	// Reload notes list
	m.loadNotesAndTackles()

	return fmt.Sprintf("Tackle %d recorded: %s %s", noteID, player, outcome), nil
}

// countTackles counts tackle notes for the current video.
func (m *Model) countTackles() (int, error) {
	rows, err := m.db.Query(
		"SELECT n.id FROM notes n INNER JOIN note_videos nv ON nv.note_id = n.id WHERE nv.path = ? AND n.category = 'tackle'",
		m.videoPath,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	return count, rows.Err()
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

// minTerminalWidth is the minimum terminal width for the three-column layout.
const minTerminalWidth = 80

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

	// Minimum width warning
	if m.width > 0 && m.width < minTerminalWidth {
		warningStyle := lipgloss.NewStyle().
			Foreground(styles.Pink).
			Bold(true)
		hintStyle := lipgloss.NewStyle().
			Foreground(styles.Lavender).
			Italic(true)
		return warningStyle.Render(fmt.Sprintf("Terminal too narrow (%d cols)", m.width)) + "\n" +
			hintStyle.Render(fmt.Sprintf("Minimum width: %d columns", minTerminalWidth)) + "\n" +
			hintStyle.Render("Please resize your terminal.")
	}

	// Render status bar at top (full width)
	statusBar := components.StatusBar(m.statusBar, m.width)

	// Check if note input is active — show as single-column overlay
	if m.noteInput.Active {
		controlsDisplay := components.ControlsDisplay(m.width)
		noteInput := components.NoteInput(m.noteInput, m.width, m.noteInput.Timestamp)
		return statusBar + "\n" + controlsDisplay + "\n" + noteInput
	}

	// Check if tackle input is active — show as single-column overlay
	if m.tackleInput.Active {
		controlsDisplay := components.ControlsDisplay(m.width)
		tackleInput := components.TackleInput(m.tackleInput, m.width, m.tackleInput.Timestamp)
		return statusBar + "\n" + controlsDisplay + "\n" + tackleInput
	}

	// --- Three-column layout ---
	// Compute column widths (equal thirds)
	// Account for 2 vertical border characters between columns
	usableWidth := m.width - 2
	colWidth := usableWidth / 3
	col3Width := usableWidth - colWidth*2 // last column gets remainder

	// Available height for columns (total height minus status bar line, timeline 2 lines, and command input line)
	colHeight := m.height - 5
	if colHeight < 5 {
		colHeight = 5
	}

	// Column 1: Controls, playback status, current tag detail card, command input
	col1Content := m.renderColumn1(colWidth, colHeight)

	// Column 2: Scrollable list of all tags/events
	col2Content := m.renderColumn2(colWidth, colHeight)

	// Column 3: Live stats summary, bar graph, top players leaderboard
	col3Content := m.renderColumn3(col3Width, colHeight)

	// Border style for vertical separators
	borderStr := lipgloss.NewStyle().
		Foreground(styles.Purple).
		Render("│")

	// Join columns horizontally with borders
	// Build each row by combining column lines side by side
	col1Lines := strings.Split(col1Content, "\n")
	col2Lines := strings.Split(col2Content, "\n")
	col3Lines := strings.Split(col3Content, "\n")

	// Pad all columns to the same height
	for len(col1Lines) < colHeight {
		col1Lines = append(col1Lines, "")
	}
	for len(col2Lines) < colHeight {
		col2Lines = append(col2Lines, "")
	}
	for len(col3Lines) < colHeight {
		col3Lines = append(col3Lines, "")
	}

	var rows []string
	for i := 0; i < colHeight; i++ {
		c1 := padToWidth(col1Lines[i], colWidth)
		c2 := padToWidth(col2Lines[i], colWidth)
		c3 := padToWidth(col3Lines[i], col3Width)
		rows = append(rows, c1+borderStr+c2+borderStr+c3)
	}

	columnsView := strings.Join(rows, "\n")

	// Render timeline progress bar below columns (full width)
	timeline := components.Timeline(m.statusBar.TimePos, m.statusBar.Duration, m.notesList.Items, m.width)

	// Render command input at bottom (full width)
	commandInput := components.CommandInput(m.commandInput, m.width)

	return statusBar + "\n" + columnsView + "\n" + timeline + "\n" + commandInput
}

// padToWidth pads a string with spaces to the specified width.
// If the string is wider (due to ANSI sequences), it is returned as-is.
func padToWidth(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// renderColumn1 renders Column 1: Controls, playback status, tag detail card.
func (m *Model) renderColumn1(width, height int) string {
	var lines []string

	// Section header
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)
	lines = append(lines, headerStyle.Render(" Controls"))
	lines = append(lines, "")

	// Compact controls display (vertical list for column)
	shortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)
	nameStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	controls := []struct {
		emoji, name, key string
	}{
		{"\u23ef\ufe0f", "Play/Pause", "Space"},
		{"\u23ea", "Back", "H"},
		{"\u23e9", "Forward", "L"},
		{"\u23ee", "Prev Item", "J"},
		{"\u23ed", "Next Item", "K"},
		{"\U0001F507", "Mute", "M"},
		{"\u2796", "Step-", "<"},
		{"\u2795", "Step+", ">"},
		{"\U0001F4DD", "Overlay", "O"},
		{"\U0001F4CA", "Stats", "S"},
		{"\u2753", "Help", "?"},
		{"\U0001F6AA", "Quit", "q"},
	}

	for _, c := range controls {
		line := fmt.Sprintf(" %s %s %s",
			c.emoji,
			nameStyle.Render(fmt.Sprintf("%-10s", c.name)),
			shortcutStyle.Render("["+c.key+"]"))
		lines = append(lines, line)
	}
	lines = append(lines, "")

	// Playback status card
	statusHeader := lipgloss.NewStyle().
		Foreground(styles.Pink).
		Bold(true)
	lines = append(lines, statusHeader.Render(" Playback"))

	infoStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)

	playState := "▶ Playing"
	if m.statusBar.Paused {
		playState = "⏸ Paused"
	}
	lines = append(lines, infoStyle.Render(" "+playState))
	lines = append(lines, infoStyle.Render(fmt.Sprintf(" Time: %s / %s",
		formatTimeString(m.statusBar.TimePos),
		formatTimeString(m.statusBar.Duration))))
	lines = append(lines, infoStyle.Render(fmt.Sprintf(" Step: %s", formatStepSize(m.statusBar.StepSize))))

	if m.statusBar.Muted {
		lines = append(lines, infoStyle.Render(" 🔇 Muted"))
	}
	if m.statusBar.OverlayEnabled {
		lines = append(lines, infoStyle.Render(" 📺 Overlay On"))
	}
	lines = append(lines, "")

	// Current tag detail card (selected item)
	item := m.notesList.GetSelectedItem()
	if item != nil {
		detailHeader := lipgloss.NewStyle().
			Foreground(styles.Pink).
			Bold(true)
		lines = append(lines, detailHeader.Render(" Selected Tag"))

		detailStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
		dimStyle := lipgloss.NewStyle().Foreground(styles.Lavender)

		typeStr := "Note"
		if item.Type == components.ItemTypeTackle {
			typeStr = "Tackle"
		}
		starStr := ""
		if item.Starred {
			starStr = " ★"
		}
		lines = append(lines, detailStyle.Render(fmt.Sprintf(" #%d %s%s", item.ID, typeStr, starStr)))
		lines = append(lines, dimStyle.Render(fmt.Sprintf(" @ %s", formatTimeString(item.TimestampSeconds))))
		if item.Category != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" [%s]", item.Category)))
		}
		if item.Player != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" Player: %s", item.Player)))
		}
		if item.Team != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf(" Team: %s", item.Team)))
		}
		if item.Text != "" {
			text := item.Text
			maxTextW := width - 3
			if maxTextW < 10 {
				maxTextW = 10
			}
			if len(text) > maxTextW {
				text = text[:maxTextW-3] + "..."
			}
			lines = append(lines, detailStyle.Render(" "+text))
		}
	}

	content := strings.Join(lines, "\n")

	// Apply column style
	colStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(styles.DeepPurple)

	return colStyle.Render(content)
}

// renderColumn2 renders Column 2: Scrollable list of all tags/events.
func (m *Model) renderColumn2(width, height int) string {
	// Use a taller list that fills the column
	listHeight := height
	if listHeight < 3 {
		listHeight = 3
	}

	notesList := components.NotesList(m.notesList, width, listHeight, m.statusBar.TimePos)

	colStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(styles.DeepPurple)

	return colStyle.Render(notesList)
}

// renderColumn3 renders Column 3: Live stats summary, bar graph, top players leaderboard.
func (m *Model) renderColumn3(width, height int) string {
	content := components.StatsPanel(m.statsView.Stats, m.notesList.Items, width, height)

	colStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(styles.DeepPurple)

	return colStyle.Render(content)
}

// formatStepSize formats the step size for display.
func formatStepSize(stepSize float64) string {
	if stepSize < 1 {
		return fmt.Sprintf("%.1fs", stepSize)
	}
	return fmt.Sprintf("%.0fs", stepSize)
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
// Uses the normalized schema: queries notes joined with note_videos, note_timing, note_details, note_tackles, note_highlights.
func (m *Model) loadNotesAndTackles() {
	if m.db == nil {
		return
	}

	var items []components.ListItem

	// Query all notes for this video with timing info
	rows, err := m.db.Query(`
		SELECT n.id, n.category, COALESCE(nt.start, 0)
		FROM notes n
		INNER JOIN note_videos nv ON nv.note_id = n.id
		LEFT JOIN note_timing nt ON nt.note_id = n.id
		WHERE nv.path = ?
		ORDER BY nt.start ASC`, m.videoPath)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var noteID int64
		var category string
		var timestamp float64
		if err := rows.Scan(&noteID, &category, &timestamp); err != nil {
			continue
		}

		item := components.ListItem{
			ID:               noteID,
			TimestampSeconds: timestamp,
			Category:         category,
		}

		// Determine type based on category
		if category == "tackle" {
			item.Type = components.ItemTypeTackle
			// Load tackle details
			tackles, err := db.SelectNoteTacklesByNote(m.db, noteID)
			if err == nil && len(tackles) > 0 {
				t := tackles[0]
				item.Player = t.Player
				item.Text = t.Player
				if t.Outcome != "" {
					item.Text += " - " + t.Outcome
				}
			}
		} else {
			item.Type = components.ItemTypeNote
		}

		// Load detail text
		details, err := db.SelectNoteDetailsByNote(m.db, noteID)
		if err == nil && len(details) > 0 {
			if item.Type == components.ItemTypeTackle && item.Text != "" {
				// Append detail text to tackle display
				item.Text += ": " + details[0].Note
			} else {
				item.Text = details[0].Note
			}
		}

		// Check for star highlights
		highlights, err := db.SelectNoteHighlightsByNote(m.db, noteID)
		if err == nil {
			for _, h := range highlights {
				if h.Type == "star" {
					item.Starred = true
					break
				}
			}
		}

		items = append(items, item)
	}

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
	case "esc":
		// Clear all filters
		m.statsView.ClearFilters()
		return m, nil
	}
	return m, nil
}

// handleStatsFilterInput handles key events when in filter input mode.
func (m *Model) handleStatsFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
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

// tackleStatsAllVideosQuery aggregates tackle stats across all videos.
const tackleStatsAllVideosQuery = `
SELECT
    ntk.player,
    COUNT(*) AS total,
    SUM(CASE WHEN ntk.outcome = 'completed' THEN 1 ELSE 0 END) AS completed,
    SUM(CASE WHEN ntk.outcome = 'missed' THEN 1 ELSE 0 END) AS missed,
    SUM(CASE WHEN ntk.outcome = 'possible' THEN 1 ELSE 0 END) AS possible,
    SUM(CASE WHEN ntk.outcome = 'other' THEN 1 ELSE 0 END) AS other,
    SUM(CASE WHEN nh.type = 'star' THEN 1 ELSE 0 END) AS starred
FROM note_tackles ntk
INNER JOIN notes n ON n.id = ntk.note_id
LEFT JOIN note_highlights nh ON nh.note_id = n.id AND nh.type = 'star'
GROUP BY ntk.player
ORDER BY total DESC`

// tackleStatsByVideoQuery aggregates tackle stats for a specific video.
const tackleStatsByVideoQuery = `
SELECT
    ntk.player,
    COUNT(*) AS total,
    SUM(CASE WHEN ntk.outcome = 'completed' THEN 1 ELSE 0 END) AS completed,
    SUM(CASE WHEN ntk.outcome = 'missed' THEN 1 ELSE 0 END) AS missed,
    SUM(CASE WHEN ntk.outcome = 'possible' THEN 1 ELSE 0 END) AS possible,
    SUM(CASE WHEN ntk.outcome = 'other' THEN 1 ELSE 0 END) AS other,
    SUM(CASE WHEN nh.type = 'star' THEN 1 ELSE 0 END) AS starred
FROM note_tackles ntk
INNER JOIN notes n ON n.id = ntk.note_id
INNER JOIN note_videos nv ON nv.note_id = n.id
LEFT JOIN note_highlights nh ON nh.note_id = n.id AND nh.type = 'star'
WHERE nv.path = ?
GROUP BY ntk.player
ORDER BY total DESC`

// loadTackleStats loads tackle statistics from the database.
func (m *Model) loadTackleStats() {
	if m.db == nil {
		return
	}

	var query string
	var args []interface{}

	if m.statsView.AllVideos {
		query = tackleStatsAllVideosQuery
	} else {
		query = tackleStatsByVideoQuery
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

// loadTackleStatsForPanel refreshes tackle stats for the live stats panel (column 3).
// Unlike loadTackleStats, this does not reset selection/scroll state.
func (m *Model) loadTackleStatsForPanel() {
	if m.db == nil {
		return
	}

	rows, err := m.db.Query(tackleStatsByVideoQuery, m.videoPath)
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
			if stat.Completed+stat.Missed > 0 {
				stat.Percentage = float64(stat.Completed) / float64(stat.Completed+stat.Missed) * 100
			}
			stats = append(stats, stat)
		}
	}

	// Only update stats if the stats view is not actively being used (to avoid interfering)
	if !m.statsView.Active {
		m.statsView.Stats = stats
	}
}
