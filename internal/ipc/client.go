package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client connects to a TUI's IPC server over a Unix socket.
type Client struct {
	socketPath string
}

// NewClient creates a Client targeting the given socket path.
func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

// IsConnected tries to connect to the socket (1s timeout) and returns whether it's reachable.
func (c *Client) IsConnected() bool {
	conn, err := net.DialTimeout("unix", c.socketPath, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Send sends an NDJSON request and waits for a response (3s timeout).
// Protocol: connect → write JSON + \n → read one line JSON → close.
func (c *Client) Send(request any) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("ipc: dial %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(3 * time.Second))

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("ipc: marshal request: %w", err)
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("ipc: write: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("ipc: read: %w", err)
		}
		return nil, fmt.Errorf("ipc: connection closed without response")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("ipc: unmarshal response: %w", err)
	}
	return &resp, nil
}
