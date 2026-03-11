package ipc

import (
	"encoding/json"
	"testing"
)

func TestClientRoundTrip(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	// Handler: echo method and data back
	go func() {
		for req := range reqCh {
			data, _ := json.Marshal(map[string]string{"echo": req.Method})
			req.Reply(Response{Type: req.Method, Data: data})
		}
	}()

	client := NewClient(sock)

	resp, err := client.Send(map[string]string{"type": "get_status"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if resp.Type != "get_status" {
		t.Fatalf("expected type=get_status, got %s", resp.Type)
	}
	if resp.Data == nil {
		t.Fatal("expected data, got nil")
	}
}

func TestClientIsConnected(t *testing.T) {
	sock := tempSocket(t)
	srv := NewServer(sock)

	client := NewClient(sock)

	// Before server starts
	if client.IsConnected() {
		t.Fatal("expected not connected before server start")
	}

	reqCh, err := srv.Start()
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	// Drain requests so connections don't block
	go func() {
		for req := range reqCh {
			req.Reply(Response{Type: "ok"})
		}
	}()

	// After server starts
	if !client.IsConnected() {
		t.Fatal("expected connected after server start")
	}

	// After server stops
	srv.Stop()
	if client.IsConnected() {
		t.Fatal("expected not connected after server stop")
	}
}

func TestClientConnectionRefused(t *testing.T) {
	client := NewClient("/tmp/mark-nonexistent-socket-for-test.sock")

	if client.IsConnected() {
		t.Fatal("expected not connected for non-existent socket")
	}

	_, err := client.Send(map[string]string{"type": "ping"})
	if err == nil {
		t.Fatal("expected error for non-existent socket")
	}
}
