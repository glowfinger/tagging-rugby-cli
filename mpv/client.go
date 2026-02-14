package mpv

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

const (
	// DefaultSocketPath is the default Unix socket path for mpv IPC.
	DefaultSocketPath = "/tmp/tagging-rugby-mpv.sock"
)

var (
	// ErrNotConnected is returned when attempting operations on a disconnected client.
	ErrNotConnected = errors.New("mpv: not connected")
	// ErrSocketNotFound is returned when the socket file doesn't exist.
	ErrSocketNotFound = errors.New("mpv: socket not found - is mpv running with --input-ipc-server?")
	// requestID is a global counter for generating unique request IDs.
	requestID uint64
)

// ipcRequest represents a JSON IPC request to mpv.
type ipcRequest struct {
	Command   []interface{} `json:"command"`
	RequestID uint64        `json:"request_id"`
}

// ipcResponse represents a JSON IPC response from mpv.
type ipcResponse struct {
	Data      interface{} `json:"data"`
	RequestID uint64      `json:"request_id"`
	Error     string      `json:"error"`
}

// Client is an mpv IPC client that communicates via Unix socket.
type Client struct {
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	mu         sync.Mutex
}

// NewClient creates a new mpv IPC client.
// If socketPath is empty, DefaultSocketPath is used.
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Client{
		socketPath: socketPath,
	}
}

// Connect establishes a connection to the mpv IPC socket.
// Returns an error if the socket doesn't exist or connection fails.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		// Check if it's a "no such file" error
		if errors.Is(err, net.UnknownNetworkError("unix")) {
			return ErrSocketNotFound
		}
		// For other connection errors (including file not found)
		return ErrSocketNotFound
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

// Close closes the connection to mpv.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	return err
}

// IsConnected returns true if the client is connected to mpv.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// SocketPath returns the socket path this client is configured to use.
func (c *Client) SocketPath() string {
	return c.socketPath
}

// GetProperty retrieves the value of an mpv property.
// The property name should be the mpv property name (e.g., "time-pos", "duration", "pause").
func (c *Client) GetProperty(name string) (interface{}, error) {
	return c.sendCommand("get_property", name)
}

// SetProperty sets the value of an mpv property.
// The property name should be the mpv property name (e.g., "pause", "speed").
func (c *Client) SetProperty(name string, value interface{}) error {
	_, err := c.sendCommand("set_property", name, value)
	return err
}

// GetTimePos returns the current playback position in seconds.
func (c *Client) GetTimePos() (float64, error) {
	result, err := c.GetProperty("time-pos")
	if err != nil {
		return 0, err
	}
	return toFloat64(result)
}

// GetDuration returns the total duration of the video in seconds.
func (c *Client) GetDuration() (float64, error) {
	result, err := c.GetProperty("duration")
	if err != nil {
		return 0, err
	}
	return toFloat64(result)
}

// GetPaused returns true if playback is paused.
func (c *Client) GetPaused() (bool, error) {
	result, err := c.GetProperty("pause")
	if err != nil {
		return false, err
	}
	paused, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("mpv: unexpected pause value type: %T", result)
	}
	return paused, nil
}

// Play resumes playback by setting pause to false.
func (c *Client) Play() error {
	return c.SetProperty("pause", false)
}

// Pause pauses playback by setting pause to true.
func (c *Client) Pause() error {
	return c.SetProperty("pause", true)
}

// TogglePause toggles the pause state.
func (c *Client) TogglePause() error {
	paused, err := c.GetPaused()
	if err != nil {
		return err
	}
	return c.SetProperty("pause", !paused)
}

// Seek performs an absolute seek to the specified position in seconds.
func (c *Client) Seek(seconds float64) error {
	_, err := c.sendCommand("seek", seconds, "absolute")
	return err
}

// SeekRelative performs a relative seek by the specified number of seconds.
// Positive values seek forward, negative values seek backward.
func (c *Client) SeekRelative(seconds float64) error {
	_, err := c.sendCommand("seek", seconds, "relative")
	return err
}

// SetSpeed sets the playback speed multiplier.
// 1.0 is normal speed, 0.5 is half speed, 2.0 is double speed.
func (c *Client) SetSpeed(multiplier float64) error {
	return c.SetProperty("speed", multiplier)
}

// GetSpeed returns the current playback speed multiplier.
func (c *Client) GetSpeed() (float64, error) {
	result, err := c.GetProperty("speed")
	if err != nil {
		return 0, err
	}
	return toFloat64(result)
}

// FrameStep advances playback by one frame and pauses.
func (c *Client) FrameStep() error {
	_, err := c.sendCommand("frame-step")
	return err
}

// FrameBackStep steps backward by one frame and pauses.
func (c *Client) FrameBackStep() error {
	_, err := c.sendCommand("frame-back-step")
	return err
}

// SetMute sets the mute state.
func (c *Client) SetMute(muted bool) error {
	return c.SetProperty("mute", muted)
}

// GetMute returns true if audio is muted.
func (c *Client) GetMute() (bool, error) {
	result, err := c.GetProperty("mute")
	if err != nil {
		return false, err
	}
	muted, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("mpv: unexpected mute value type: %T", result)
	}
	return muted, nil
}

// SetABLoop sets the A-B loop points for looping playback between start and end times.
// Both start and end are in seconds.
func (c *Client) SetABLoop(start, end float64) error {
	if err := c.SetProperty("ab-loop-a", start); err != nil {
		return err
	}
	return c.SetProperty("ab-loop-b", end)
}

// ClearABLoop clears the A-B loop by setting both loop points to "no".
func (c *Client) ClearABLoop() error {
	if err := c.SetProperty("ab-loop-a", "no"); err != nil {
		return err
	}
	return c.SetProperty("ab-loop-b", "no")
}

// ShowOverlay displays text on the mpv video using osd-overlay.
// The overlayID identifies the overlay (use 1 for notes overlay).
// The text is displayed with ASS formatting support for styling.
func (c *Client) ShowOverlay(overlayID int, text string) error {
	// osd-overlay command format: osd-overlay <id> <format> <data>
	// format "ass-events" allows ASS styling
	_, err := c.sendCommand("osd-overlay", overlayID, "ass-events", text)
	return err
}

// HideOverlay removes a displayed overlay.
// The overlayID must match the ID used in ShowOverlay.
func (c *Client) HideOverlay(overlayID int) error {
	// To hide an overlay, set data to empty string with "none" format
	_, err := c.sendCommand("osd-overlay", overlayID, "none", "")
	return err
}

// toFloat64 converts an interface{} to float64.
// JSON numbers from mpv are typically decoded as float64.
func toFloat64(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("mpv: unexpected numeric value type: %T", v)
	}
}

// sendCommand sends a JSON IPC command to mpv and returns the result.
// The command is formatted as {"command": [command, args...], "request_id": <id>}
// and sent as newline-terminated JSON over the socket.
func (c *Client) sendCommand(command string, args ...interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, ErrNotConnected
	}

	// Build command array: [command, arg1, arg2, ...]
	cmdArray := make([]interface{}, 0, len(args)+1)
	cmdArray = append(cmdArray, command)
	cmdArray = append(cmdArray, args...)

	// Generate unique request ID
	reqID := atomic.AddUint64(&requestID, 1)

	// Create request
	req := ipcRequest{
		Command:   cmdArray,
		RequestID: reqID,
	}

	// Encode to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("mpv: failed to marshal command: %w", err)
	}

	// Send newline-terminated JSON
	data = append(data, '\n')
	if _, err := c.conn.Write(data); err != nil {
		// Connection is broken — clean up so IsConnected() returns false
		c.conn.Close()
		c.conn = nil
		c.reader = nil
		return nil, fmt.Errorf("mpv: failed to send command: %w", err)
	}

	// Read response lines until we get our request_id
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			// Connection is broken — clean up so IsConnected() returns false
			c.conn.Close()
			c.conn = nil
			c.reader = nil
			return nil, fmt.Errorf("mpv: failed to read response: %w", err)
		}

		var resp ipcResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Skip malformed lines (could be events)
			continue
		}

		// Check if this is our response
		if resp.RequestID == reqID {
			if resp.Error != "" && resp.Error != "success" {
				return nil, fmt.Errorf("mpv: %s", resp.Error)
			}
			return resp.Data, nil
		}
		// If request_id doesn't match, it's probably an event - skip and keep reading
	}
}
