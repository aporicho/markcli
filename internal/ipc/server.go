package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

// SocketPath returns the default IPC socket path: /tmp/mark-{uid}.sock
func SocketPath() string {
	return fmt.Sprintf("/tmp/mark-%d.sock", os.Getuid())
}

// Server listens on a Unix socket and dispatches incoming NDJSON requests.
type Server struct {
	path     string
	listener net.Listener
	reqCh    chan Request
	done     chan struct{}
	wg       sync.WaitGroup
	once     sync.Once
}

// NewServer creates a Server bound to the given socket path.
func NewServer(socketPath string) *Server {
	return &Server{
		path:  socketPath,
		reqCh: make(chan Request, 64),
		done:  make(chan struct{}),
	}
}

// Start begins listening on the Unix socket and returns a channel of incoming requests.
// The caller should read from the channel and call req.Reply() for each request.
func (s *Server) Start() (<-chan Request, error) {
	// Clean up stale socket file
	os.Remove(s.path)

	ln, err := net.Listen("unix", s.path)
	if err != nil {
		return nil, fmt.Errorf("ipc: listen %s: %w", s.path, err)
	}
	s.listener = ln

	s.wg.Add(1)
	go s.acceptLoop()

	return s.reqCh, nil
}

// Stop closes the listener, waits for in-flight connections to drain,
// and removes the socket file.
func (s *Server) Stop() {
	s.once.Do(func() {
		close(s.done)
		s.listener.Close()
		s.wg.Wait()
		os.Remove(s.path)
	})
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

// rawRequest is the wire format: {"type":"method_name", ...params}
type rawRequest struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// Copy bytes — scanner reuses its internal buffer
		line := make([]byte, len(scanner.Bytes()))
		copy(line, scanner.Bytes())

		var raw rawRequest
		if err := json.Unmarshal(line, &raw); err != nil {
			errResp := Response{Type: "error", Message: fmt.Sprintf("invalid JSON: %s", err)}
			b, merr := json.Marshal(errResp)
			if merr != nil {
				b = []byte(`{"type":"error","message":"marshal failed"}`)
			}
			b = append(b, '\n')
			conn.Write(b)
			continue
		}

		if raw.Type == "" {
			errResp := Response{Type: "error", Message: "missing \"type\" field"}
			b, merr := json.Marshal(errResp)
			if merr != nil {
				b = []byte(`{"type":"error","message":"marshal failed"}`)
			}
			b = append(b, '\n')
			conn.Write(b)
			continue
		}

		req := NewRequest(raw.ID, raw.Type, json.RawMessage(line))

		// Send to channel or bail if server is stopping
		select {
		case s.reqCh <- req:
		case <-s.done:
			return
		}

		// Wait for handler to reply
		select {
		case resp := <-req.Chan():
			b, merr := json.Marshal(resp)
			if merr != nil {
				b = []byte(`{"type":"error","message":"marshal failed"}`)
			}
			b = append(b, '\n')
			conn.Write(b)
		case <-s.done:
			return
		}
	}
}
