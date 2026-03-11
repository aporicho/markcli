package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
)

func tempSocket(t *testing.T) string {
	t.Helper()
	// Use /tmp directly — macOS limits Unix socket paths to ~104 bytes,
	// and t.TempDir() paths are often too long.
	f, err := os.CreateTemp("/tmp", "mark-test-*.sock")
	if err != nil {
		t.Fatalf("create temp socket: %v", err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path)
	t.Cleanup(func() { os.Remove(path) })
	return path
}

func TestSocketPath(t *testing.T) {
	p := SocketPath()
	if !strings.HasPrefix(p, "/tmp/mark-") {
		t.Fatalf("unexpected prefix: %s", p)
	}
	if !strings.HasSuffix(p, ".sock") {
		t.Fatalf("unexpected suffix: %s", p)
	}
}

func TestServerStartStop(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	_, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Socket file should exist
	if _, err := os.Stat(sock); err != nil {
		t.Fatalf("socket file missing after Start: %v", err)
	}

	srv.Stop()

	// Socket file should be cleaned up
	if _, err := os.Stat(sock); !os.IsNotExist(err) {
		t.Fatalf("socket file not cleaned up after Stop")
	}

	// Double stop should not panic
	srv.Stop()
}

func TestServerRoundTrip(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	// Handler goroutine
	go func() {
		for req := range reqCh {
			req.Reply(Response{Type: req.Method, Message: "ok"})
		}
	}()

	// Connect and send a request
	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	fmt.Fprintln(conn, `{"type":"get_status","id":"1"}`)

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		t.Fatal("no response")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.Type != "get_status" {
		t.Fatalf("expected type=get_status, got %s", resp.Type)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected message=ok, got %s", resp.Message)
	}
}

func TestServerBadJSON(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	// Handler (shouldn't receive anything for bad JSON)
	go func() {
		for req := range reqCh {
			req.Reply(Response{Type: req.Method, Message: "ok"})
		}
	}()

	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// Send bad JSON
	fmt.Fprintln(conn, `not json at all`)

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		t.Fatal("no response for bad JSON")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Type != "error" {
		t.Fatalf("expected error response, got type=%s", resp.Type)
	}
	if !strings.Contains(resp.Message, "invalid JSON") {
		t.Fatalf("expected 'invalid JSON' in message, got: %s", resp.Message)
	}

	// Server still works: send a valid request
	fmt.Fprintln(conn, `{"type":"ping"}`)
	if !scanner.Scan() {
		t.Fatal("no response after bad JSON recovery")
	}

	var resp2 Response
	json.Unmarshal(scanner.Bytes(), &resp2)
	if resp2.Type != "ping" {
		t.Fatalf("expected ping, got %s", resp2.Type)
	}
}

func TestServerConcurrentConnections(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	// Handler: echo method back
	go func() {
		for req := range reqCh {
			req.Reply(Response{Type: req.Method, Message: "ok"})
		}
	}()

	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("unix", sock)
			if err != nil {
				errs <- fmt.Errorf("client %d dial: %w", id, err)
				return
			}
			defer conn.Close()

			method := fmt.Sprintf("test_%d", id)
			fmt.Fprintf(conn, `{"type":"%s"}`+"\n", method)

			scanner := bufio.NewScanner(conn)
			if !scanner.Scan() {
				errs <- fmt.Errorf("client %d: no response", id)
				return
			}

			var resp Response
			if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
				errs <- fmt.Errorf("client %d unmarshal: %w", id, err)
				return
			}
			if resp.Type != method {
				errs <- fmt.Errorf("client %d: expected %s, got %s", id, method, resp.Type)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestStaleSocketCleanup(t *testing.T) {
	sock := tempSocket(t)

	// Create a stale socket file
	if err := os.WriteFile(sock, []byte("stale"), 0o600); err != nil {
		t.Fatalf("create stale file: %v", err)
	}

	srv := NewServer(sock)
	_, err := srv.Start()
	if err != nil {
		t.Fatalf("Start with stale socket: %v", err)
	}
	srv.Stop()
}

func TestServerMissingType(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	go func() {
		for req := range reqCh {
			req.Reply(Response{Type: req.Method, Message: "ok"})
		}
	}()

	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// Valid JSON but missing type
	fmt.Fprintln(conn, `{"foo":"bar"}`)

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		t.Fatal("no response")
	}

	var resp Response
	json.Unmarshal(scanner.Bytes(), &resp)
	if resp.Type != "error" {
		t.Fatalf("expected error, got %s", resp.Type)
	}
	if !strings.Contains(resp.Message, "missing") {
		t.Fatalf("expected 'missing' in message, got: %s", resp.Message)
	}
}
