package mpv

import (
	"errors"
	"net"
	"sync"
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
)

// Client is an mpv IPC client that communicates via Unix socket.
type Client struct {
	socketPath string
	conn       net.Conn
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
