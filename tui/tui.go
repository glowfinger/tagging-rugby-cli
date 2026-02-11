package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/user/tagging-rugby-cli/db"
	"github.com/user/tagging-rugby-cli/mpv"
	"github.com/user/tagging-rugby-cli/tui/components"
	"github.com/user/tagging-rugby-cli/tui/forms"
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

// speedSizes defines the available playback speed levels.
// Users can cycle through these with [ ] { } keys.
var speedSizes = []float64{0.5, 0.75, 1.0, 1.25, 1.5, 2.0}

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
	// noteForm is the huh form for note input (nil when inactive)
	noteForm *huh.Form
	// noteFormResult holds the bound values for the note form
	noteFormResult forms.NoteFormResult
	// noteFormTimestamp is the timestamp captured when the note form was opened
	noteFormTimestamp float64
	// tackleForm is the huh form for tackle input (nil when inactive)
	tackleForm *huh.Form
	// tackleFormResult holds the bound values for the tackle form
	tackleFormResult forms.TackleFormResult
	// tackleFormTimestamp is the timestamp captured when the tackle form was opened
	tackleFormTimestamp float64
	// confirmDiscardForm is shown when user presses Esc on a form with data (nil when inactive)
	confirmDiscardForm *huh.Form
	// confirmDiscard holds the confirm result (true = discard, false = go back)
	confirmDiscard bool
	// confirmDiscardTarget tracks which form triggered the confirm ("note" or "tackle")
	confirmDiscardTarget string
	// tackleSortColumn tracks the sort column for the tackle stats table in column 3
	tackleSortColumn components.SortColumn
}

// NewModel creates a new TUI model with the given mpv client, database connection, and video path.
func NewModel(client *mpv.Client, db *sql.DB, videoPath string) *Model {
	return &Model{
		client:    client,
		db:        db,
		videoPath: videoPath,
		statusBar: components.StatusBarState{
			StepSize: defaultStepSize,
			Speed:    1.0,
		},
		tackleSortColumn: components.SortByTotal,
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
	// Delegate all messages to active huh form (it needs non-key messages too)
	if m.confirmDiscardForm != nil || m.noteForm != nil || m.tackleForm != nil {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			if _, isTick := msg.(tickMsg); !isTick {
				if _, isClear := msg.(clearResultMsg); !isClear {
					if _, isResize := msg.(tea.WindowSizeMsg); !isResize {
						if m.confirmDiscardForm != nil {
							return m.handleConfirmDiscardUpdate(msg)
						}
						if m.noteForm != nil {
							return m.handleNoteFormUpdate(msg)
						}
						return m.handleTackleFormUpdate(msg)
					}
				}
			}
		}
	}

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

		// Handle confirm discard dialog (huh form)
		if m.confirmDiscardForm != nil {
			return m.handleConfirmDiscardUpdate(msg)
		}

		// Handle note form mode (huh form)
		if m.noteForm != nil {
			return m.handleNoteFormUpdate(msg)
		}

		// Handle tackle form mode (huh form)
		if m.tackleForm != nil {
			return m.handleTackleFormUpdate(msg)
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
		case "ctrl+c":
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
		case "ctrl+h":
			// Ctrl+H steps backward by one frame
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.FrameBackStep()
			}
			return m, nil
		case "ctrl+l":
			// Ctrl+L steps forward by one frame
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.FrameStep()
			}
			return m, nil
		case "h", "H", "left":
			// H / Left arrow steps backward by current step size
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.SeekRelative(-m.statusBar.StepSize)
			}
			return m, nil
		case "l", "L", "right":
			// L / Right arrow steps forward by current step size
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.SeekRelative(m.statusBar.StepSize)
			}
			return m, nil
		case "<", ",":
			// < / , decreases step size
			m.decreaseStepSize()
			return m, nil
		case ">", ".":
			// > / . increases step size
			m.increaseStepSize()
			return m, nil
		case "j", "J", "up":
			// J / Up arrow moves selection to previous note/tackle in list
			m.notesList.MoveUp()
			return m, nil
		case "k", "K", "down":
			// K / Down arrow moves selection to next note/tackle in list
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
		case "[", "{":
			// [ / { decreases playback speed
			m.decreaseSpeed()
			return m, nil
		case "]", "}":
			// ] / } increases playback speed
			m.increaseSpeed()
			return m, nil
		case "\\":
			// Backslash resets speed to 1.0x
			if m.client != nil && m.client.IsConnected() {
				_ = m.client.SetSpeed(1.0)
			}
			return m, nil
		case "x", "X":
			// X cycles through tackle stats sort columns
			m.cycleTackleSortColumn()
			return m, nil
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

// openNoteInput opens the huh note form.
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

	// Initialize huh note form
	m.noteFormResult = forms.NoteFormResult{}
	m.noteFormTimestamp = timestamp
	m.noteForm = forms.NewNoteForm(timestamp, &m.noteFormResult)

	return m, m.noteForm.Init()
}

// handleNoteFormUpdate delegates messages to the huh note form and handles completion.
func (m *Model) handleNoteFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.noteForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.noteForm = f
	}

	// Check if form was completed or cancelled
	if m.noteForm.State == huh.StateCompleted {
		return m.saveNoteFromForm()
	}
	if m.noteForm.State == huh.StateAborted {
		// If form has data, show confirm discard dialog
		if m.noteFormResult.HasData() {
			return m.openConfirmDiscard("note")
		}
		m.noteForm = nil
		return m, nil
	}

	return m, cmd
}

// saveNoteFromForm saves the note data from the completed huh form.
func (m *Model) saveNoteFromForm() (tea.Model, tea.Cmd) {
	result := m.noteFormResult
	timestamp := m.noteFormTimestamp

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
			{Type: "text", Note: result.Text},
		},
	}

	// Use category from input, default to "note"
	category := result.Category
	if category == "" {
		category = "note"
	}

	// Save note with children
	noteID, err := db.InsertNoteWithChildren(m.db, category, children)
	m.noteForm = nil

	if err != nil {
		m.commandInput.SetResult("Error: "+err.Error(), true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Reload list and show confirmation
	m.loadNotesAndTackles()
	m.commandInput.SetResult(fmt.Sprintf("Note %d added at %s", noteID, formatTimeString(timestamp)), false)
	return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

// openTackleInput opens the huh tackle wizard form.
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

	// Initialize huh tackle form
	m.tackleFormResult = forms.TackleFormResult{}
	m.tackleFormTimestamp = timestamp
	m.tackleForm = forms.NewTackleForm(timestamp, &m.tackleFormResult)

	return m, m.tackleForm.Init()
}

// handleTackleFormUpdate delegates messages to the huh tackle form and handles completion.
func (m *Model) handleTackleFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.tackleForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.tackleForm = f
	}

	// Check if form was completed or cancelled
	if m.tackleForm.State == huh.StateCompleted {
		return m.saveTackleFromForm()
	}
	if m.tackleForm.State == huh.StateAborted {
		// If form has data, show confirm discard dialog
		if m.tackleFormResult.HasData() {
			return m.openConfirmDiscard("tackle")
		}
		m.tackleForm = nil
		return m, nil
	}

	return m, cmd
}

// openConfirmDiscard opens a confirm dialog when user presses Esc on a form with data.
// The target parameter indicates which form triggered the confirm ("note" or "tackle").
func (m *Model) openConfirmDiscard(target string) (tea.Model, tea.Cmd) {
	m.confirmDiscard = false
	m.confirmDiscardTarget = target
	m.confirmDiscardForm = forms.NewConfirmDiscardForm(&m.confirmDiscard)
	return m, m.confirmDiscardForm.Init()
}

// handleConfirmDiscardUpdate delegates messages to the confirm discard form.
func (m *Model) handleConfirmDiscardUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.confirmDiscardForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.confirmDiscardForm = f
	}

	if m.confirmDiscardForm.State == huh.StateCompleted {
		m.confirmDiscardForm = nil
		if m.confirmDiscard {
			// User chose to discard — close the underlying form
			if m.confirmDiscardTarget == "note" {
				m.noteForm = nil
			} else {
				m.tackleForm = nil
			}
			return m, nil
		}
		// User chose to go back — reopen the form from saved state
		if m.confirmDiscardTarget == "note" {
			m.noteForm = forms.NewNoteForm(m.noteFormTimestamp, &m.noteFormResult)
			return m, m.noteForm.Init()
		}
		m.tackleForm = forms.NewTackleForm(m.tackleFormTimestamp, &m.tackleFormResult)
		return m, m.tackleForm.Init()
	}

	if m.confirmDiscardForm.State == huh.StateAborted {
		// Esc on confirm dialog — treat as "go back" to form
		m.confirmDiscardForm = nil
		if m.confirmDiscardTarget == "note" {
			m.noteForm = forms.NewNoteForm(m.noteFormTimestamp, &m.noteFormResult)
			return m, m.noteForm.Init()
		}
		m.tackleForm = forms.NewTackleForm(m.tackleFormTimestamp, &m.tackleFormResult)
		return m, m.tackleForm.Init()
	}

	return m, cmd
}

// saveTackleFromForm saves the tackle data from the completed huh form.
func (m *Model) saveTackleFromForm() (tea.Model, tea.Cmd) {
	result := m.tackleFormResult
	timestamp := m.tackleFormTimestamp

	// Parse attempt as integer
	var attempt int
	fmt.Sscanf(result.Attempt, "%d", &attempt)

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
			{Player: result.Player, Attempt: attempt, Outcome: result.Outcome},
		},
	}

	// Add followed detail if provided (maps to note_detail type="followed")
	if result.Followed != "" {
		children.Details = append(children.Details, db.NoteDetail{
			Type: "followed", Note: result.Followed,
		})
	}

	// Add notes detail if provided (maps to note_detail type="notes")
	if result.Notes != "" {
		children.Details = append(children.Details, db.NoteDetail{
			Type: "notes", Note: result.Notes,
		})
	}

	// Add zone if provided (maps to note_zones)
	if result.Zone != "" {
		children.Zones = []db.NoteZone{
			{Horizontal: result.Zone},
		}
	}

	// Add highlight if starred (maps to note_highlights type="star")
	if result.Star {
		children.Highlights = []db.NoteHighlight{
			{Type: "star"},
		}
	}

	// Category is always "tackle" — auto-set, not a form field
	noteID, err := db.InsertNoteWithChildren(m.db, "tackle", children)
	m.tackleForm = nil

	if err != nil {
		m.commandInput.SetResult("Error: "+err.Error(), true)
		return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		})
	}

	// Reload list and show confirmation
	m.loadNotesAndTackles()
	starSymbol := ""
	if result.Star {
		starSymbol = " ★"
	}
	m.commandInput.SetResult(fmt.Sprintf("Tackle %d recorded: %s %s%s", noteID, result.Player, result.Outcome, starSymbol), false)
	return m, tea.Tick(resultDisplayDuration, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
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

// cycleTackleSortColumn cycles to the next sort column for the tackle stats table.
// Order: Total -> Completed -> Missed -> Percentage -> Player -> Total (loops)
func (m *Model) cycleTackleSortColumn() {
	switch m.tackleSortColumn {
	case components.SortByTotal:
		m.tackleSortColumn = components.SortByCompleted
	case components.SortByCompleted:
		m.tackleSortColumn = components.SortByMissed
	case components.SortByMissed:
		m.tackleSortColumn = components.SortByPercentage
	case components.SortByPercentage:
		m.tackleSortColumn = components.SortByPlayer
	case components.SortByPlayer:
		m.tackleSortColumn = components.SortByTotal
	default:
		m.tackleSortColumn = components.SortByTotal
	}
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

// decreaseSpeed cycles to the previous (slower) playback speed.
func (m *Model) decreaseSpeed() {
	if m.client == nil || !m.client.IsConnected() {
		return
	}
	currentIndex := m.findSpeedIndex()
	if currentIndex > 0 {
		_ = m.client.SetSpeed(speedSizes[currentIndex-1])
	}
}

// increaseSpeed cycles to the next (faster) playback speed.
func (m *Model) increaseSpeed() {
	if m.client == nil || !m.client.IsConnected() {
		return
	}
	currentIndex := m.findSpeedIndex()
	if currentIndex < len(speedSizes)-1 {
		_ = m.client.SetSpeed(speedSizes[currentIndex+1])
	}
}

// findSpeedIndex finds the index of the current speed in the speedSizes array.
// If the current speed is not in the array, it returns the index of the closest value.
func (m *Model) findSpeedIndex() int {
	currentSpeed := m.statusBar.Speed
	if currentSpeed == 0 {
		currentSpeed = 1.0
	}
	for i, speed := range speedSizes {
		if currentSpeed == speed {
			return i
		}
	}
	// Find closest if not exact match
	for i, speed := range speedSizes {
		if currentSpeed < speed {
			if i == 0 {
				return 0
			}
			return i - 1
		}
	}
	return len(speedSizes) - 1
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

	// Get playback speed
	speed, err := m.client.GetSpeed()
	if err == nil {
		m.statusBar.Speed = speed
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

	// Narrow terminal: show standalone mini player with warning
	if m.width > 0 && m.width < minTerminalWidth {
		return components.RenderMiniPlayer(m.statusBar, 0, true)
	}

	// Render status bar at top (full width)
	statusBar := components.StatusBar(m.statusBar, m.width)

	// Check if confirm discard dialog is active — show it as overlay
	if m.confirmDiscardForm != nil {
		controlsDisplay := components.ControlsDisplay(m.width)
		confirmView := m.confirmDiscardForm.View()
		return statusBar + "\n" + controlsDisplay + "\n" + confirmView
	}

	// Check if note form is active — show huh form as overlay
	if m.noteForm != nil {
		controlsDisplay := components.ControlsDisplay(m.width)
		noteFormView := m.noteForm.View()
		return statusBar + "\n" + controlsDisplay + "\n" + noteFormView
	}

	// Check if tackle form is active — show huh wizard as overlay
	if m.tackleForm != nil {
		controlsDisplay := components.ControlsDisplay(m.width)
		tackleFormView := m.tackleForm.View()
		return statusBar + "\n" + controlsDisplay + "\n" + tackleFormView
	}

	// --- Three-column layout with responsive sizing ---
	// Column 1 is fixed at 30 characters. Column 3 shrinks first; hides at narrow widths.
	const (
		col1Width         = 30  // fixed width for column 1
		col3HideThreshold = 90  // below this width, hide column 3 entirely
		col3MinWidth      = 18  // minimum width for column 3 before hiding
	)

	// Available height for columns (total height minus status bar line, bordered timeline 6 lines, command input line, and newline separators)
	colHeight := m.height - 9
	if colHeight < 5 {
		colHeight = 5
	}

	showCol3 := m.width >= col3HideThreshold

	var columnsView string

	if showCol3 {
		// Three-column layout: col1 fixed at 30, no border characters
		usableWidth := m.width
		var col2Width, col3Width int

		if m.width >= 120 {
			// Wide: col1 = 30, col2 and col3 split the remaining space
			remaining := usableWidth - col1Width
			col2Width = remaining / 2
			col3Width = remaining - col2Width
		} else {
			// Medium (90-119): col1 = 30, col3 = col3MinWidth, col2 gets the rest
			col3Width = col3MinWidth
			col2Width = usableWidth - col1Width - col3Width
		}

		col1Content := m.renderColumn1(col1Width, colHeight)
		col2Content := m.renderColumn2(col2Width, colHeight)
		col3Content := m.renderColumn3(col3Width, colHeight)

		col1Lines := normalizeLines(strings.Split(col1Content, "\n"), colHeight)
		col2Lines := normalizeLines(strings.Split(col2Content, "\n"), colHeight)
		col3Lines := normalizeLines(strings.Split(col3Content, "\n"), colHeight)

		var rows []string
		for i := 0; i < colHeight; i++ {
			c1 := padToWidth(col1Lines[i], col1Width)
			c2 := padToWidth(col2Lines[i], col2Width)
			c3 := padToWidth(col3Lines[i], col3Width)
			rows = append(rows, c1+c2+c3)
		}
		columnsView = strings.Join(rows, "\n")
	} else {
		// Two-column layout: col1 fixed at 30, column 3 hidden, no border characters
		usableWidth := m.width
		col2Width := usableWidth - col1Width

		col1Content := m.renderColumn1(col1Width, colHeight)
		col2Content := m.renderColumn2(col2Width, colHeight)

		col1Lines := normalizeLines(strings.Split(col1Content, "\n"), colHeight)
		col2Lines := normalizeLines(strings.Split(col2Content, "\n"), colHeight)

		var rows []string
		for i := 0; i < colHeight; i++ {
			c1 := padToWidth(col1Lines[i], col1Width)
			c2 := padToWidth(col2Lines[i], col2Width)
			rows = append(rows, c1+c2)
		}
		columnsView = strings.Join(rows, "\n")
	}

	// Render timeline progress bar below columns (full width)
	timeline := components.Timeline(m.statusBar.TimePos, m.statusBar.Duration, m.notesList.Items, m.width)

	// Render command input at bottom (full width)
	commandInput := components.CommandInput(m.commandInput, m.width)

	return statusBar + "\n" + columnsView + "\n" + timeline + "\n" + commandInput
}

// padToWidth pads or truncates a string to exactly the specified width.
// Uses ansi.Truncate for ANSI-aware, grapheme-aware truncation that correctly
// handles double-width characters (emoji, East-Asian).
func padToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	currentWidth := lipgloss.Width(s)
	if currentWidth == width {
		return s
	}
	if currentWidth > width {
		s = ansi.Truncate(s, width, "")
		currentWidth = lipgloss.Width(s)
	}
	if currentWidth < width {
		return s + strings.Repeat(" ", width-currentWidth)
	}
	return s
}

// normalizeLines pads or truncates a slice of strings to exactly the given height.
func normalizeLines(lines []string, height int) []string {
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}

// renderColumn1 renders Column 1: Playback status, selected tag detail, controls.
func (m *Model) renderColumn1(width, height int) string {
	var lines []string

	// Bordered mini player card
	miniPlayer := components.RenderMiniPlayer(m.statusBar, width, false)
	lines = append(lines, strings.Split(miniPlayer, "\n")...)

	// Current tag detail card (selected item) — bordered box
	item := m.notesList.GetSelectedItem()
	if item != nil {
		detailStyle := lipgloss.NewStyle().Foreground(styles.LightLavender)
		dimStyle := lipgloss.NewStyle().Foreground(styles.Lavender)

		// Inner width = column width - 4 (2 border chars + 2 padding/space)
		innerW := width - 4
		if innerW < 10 {
			innerW = 10
		}

		typeStr := "Note"
		if item.Type == components.ItemTypeTackle {
			typeStr = "Tackle"
		}
		starStr := ""
		if item.Starred {
			starStr = " ★"
		}

		var contentLines []string
		contentLines = append(contentLines, detailStyle.Render(fmt.Sprintf(" #%d %s%s", item.ID, typeStr, starStr)))
		contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" @ %s", formatTimeString(item.TimestampSeconds))))
		if item.Category != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" [%s]", item.Category)))
		}
		if item.Player != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" Player: %s", item.Player)))
		}
		if item.Team != "" {
			contentLines = append(contentLines, dimStyle.Render(fmt.Sprintf(" Team: %s", item.Team)))
		}
		if item.Text != "" {
			text := item.Text
			if len(text) > innerW {
				text = text[:innerW-3] + "..."
			}
			contentLines = append(contentLines, detailStyle.Render(" "+text))
		}

		infoBox := components.RenderInfoBox("Selected Tag", contentLines, width)
		lines = append(lines, strings.Split(infoBox, "\n")...)
	}

	// Controls section: each group in a bordered container
	shortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Cyan).
		Bold(true)
	nameStyle := lipgloss.NewStyle().
		Foreground(styles.LightLavender)

	groups := components.GetControlGroups()
	for _, group := range groups {
		var contentLines []string
		for _, c := range group.Controls {
			// Format: [Key] Label
			line := " " + shortcutStyle.Render("["+c.Shortcut+"]") + " " + nameStyle.Render(c.Name)
			contentLines = append(contentLines, line)
		}
		infoBox := components.RenderInfoBox(group.Name, contentLines, width)
		lines = append(lines, strings.Split(infoBox, "\n")...)
	}

	return strings.Join(lines, "\n")
}

// renderColumn2 renders Column 2: Scrollable list of all tags/events in a bordered box.
func (m *Model) renderColumn2(width, height int) string {
	// List height shrinks to account for box chrome: 3 tab header lines + 1 bottom border = 4
	listHeight := height - 4
	if listHeight < 3 {
		listHeight = 3
	}

	// Inner width for the table content (column width - 2 for border chars)
	innerWidth := width - 2
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Render the notes list at inner width
	tableContent := components.NotesList(m.notesList, innerWidth, listHeight, m.statusBar.TimePos)

	// Build bordered box with tab-style "Notes" header
	headerStyle := lipgloss.NewStyle().Foreground(styles.Pink).Bold(true)
	borderColor := styles.Purple
	topBorderStyle := lipgloss.NewStyle().Foreground(borderColor)
	sideStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Tab header: ╭─ Notes ─────╮
	headerText := headerStyle.Render(" Notes ")
	headerTextWidth := lipgloss.Width(headerText)
	topLeft := topBorderStyle.Render("╭─")
	topRight := topBorderStyle.Render("╮")
	fillWidth := innerWidth - 2 - headerTextWidth - 1 + 2 // innerWidth - topLeftChars - headerWidth - topRightChar + border adjustment
	if fillWidth < 0 {
		fillWidth = 0
	}
	topLine := topLeft + headerText + topBorderStyle.Render(strings.Repeat("─", fillWidth)) + topRight

	var lines []string
	lines = append(lines, topLine)

	// Wrap each table line in side borders
	tableLines := strings.Split(tableContent, "\n")
	for _, tl := range tableLines {
		lineWidth := lipgloss.Width(tl)
		pad := innerWidth - lineWidth
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, sideStyle.Render("│")+tl+strings.Repeat(" ", pad)+sideStyle.Render("│"))
	}

	// Pad to fill column height (account for top line + bottom line)
	targetContentLines := height - 2 // total height minus top border and bottom border
	for len(lines) < targetContentLines+1 { // +1 because lines already includes top line
		lines = append(lines, sideStyle.Render("│")+strings.Repeat(" ", innerWidth)+sideStyle.Render("│"))
	}

	// Bottom border: ╰──────────────╯
	bottomLine := topBorderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	lines = append(lines, bottomLine)

	return strings.Join(lines, "\n")
}

// renderColumn3 renders Column 3: Live stats summary, bar graph, top players leaderboard.
func (m *Model) renderColumn3(width, height int) string {
	return components.StatsPanel(m.statsView.Stats, m.notesList.Items, width, height, m.tackleSortColumn)
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
	case "ctrl+c":
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
