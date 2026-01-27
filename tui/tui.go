package tui

import (
	"database/sql"
	"sort"
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
)

// tickMsg is a message sent on every tick interval to update playback status.
type tickMsg time.Time

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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
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

	// Calculate available height for notes list (minus status bar and footer)
	listHeight := m.height - 3 // 1 for status bar, 2 for footer
	if listHeight < 5 {
		listHeight = 5
	}

	// Render notes list
	notesList := components.NotesList(m.notesList, m.width, listHeight)

	// Footer with help hint
	footer := "\n Press q to quit"

	return statusBar + "\n" + notesList + footer
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
