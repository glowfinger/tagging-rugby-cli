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
		return nil, fmt.Errorf("mpv: failed to send command: %w", err)
	}

	// Read response lines until we get our request_id
	for {
		line, err := c.reader.ReadBytes('\n')
		if err != nil {
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
