package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("Hub clients map is nil")
	}
	if hub.broadcast == nil {
		t.Error("Hub broadcast channel is nil")
	}
	if hub.register == nil {
		t.Error("Hub register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("Hub unregister channel is nil")
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	event := &LogEvent{
		Type:      "test",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "info",
		Message:   "test message",
	}

	// Should not panic even with no clients
	hub.Broadcast(event)

	time.Sleep(10 * time.Millisecond)
}

func TestHub_BroadcastLog(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	data := map[string]interface{}{
		"key": "value",
	}

	hub.BroadcastLog("info", "test", "test message", data)

	time.Sleep(10 * time.Millisecond)
}

func TestHub_GetClientCount(t *testing.T) {
	hub := NewHub()
	if hub.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", hub.GetClientCount())
	}
}

func TestHub_ServeWS(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect as WebSocket client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait for connection to be registered
	time.Sleep(50 * time.Millisecond)

	if hub.GetClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", hub.GetClientCount())
	}

	// Broadcast a test message
	hub.BroadcastLog("info", "test", "Hello from server", nil)

	// Read the message
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var event LogEvent
	if err := json.Unmarshal(message, &event); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if event.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", event.Type)
	}
	if event.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", event.Level)
	}
	if event.Message != "Hello from server" {
		t.Errorf("Expected message 'Hello from server', got '%s'", event.Message)
	}
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect multiple clients
	var conns []*websocket.Conn
	for i := 0; i < 3; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		defer conn.Close()
		conns = append(conns, conn)
	}

	// Wait for connections to be registered
	time.Sleep(50 * time.Millisecond)

	if hub.GetClientCount() != 3 {
		t.Errorf("Expected 3 clients, got %d", hub.GetClientCount())
	}

	// Broadcast a message
	testMessage := "broadcast to all"
	hub.BroadcastLog("info", "broadcast", testMessage, nil)

	// All clients should receive the message
	for i, conn := range conns {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Client %d failed to read message: %v", i, err)
		}

		var event LogEvent
		if err := json.Unmarshal(message, &event); err != nil {
			t.Fatalf("Client %d failed to unmarshal: %v", i, err)
		}

		if event.Message != testMessage {
			t.Errorf("Client %d: expected '%s', got '%s'", i, testMessage, event.Message)
		}
	}
}

func TestLogEvent_MarshalJSON(t *testing.T) {
	event := &LogEvent{
		Type:      "test",
		Timestamp: "2025-11-04T12:00:00Z",
		Level:     "info",
		Message:   "test message",
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal LogEvent: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled["type"] != "test" {
		t.Errorf("Expected type 'test', got '%v'", unmarshaled["type"])
	}
	if unmarshaled["message"] != "test message" {
		t.Errorf("Expected message 'test message', got '%v'", unmarshaled["message"])
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if hub.GetClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", hub.GetClientCount())
	}

	// Close the connection
	conn.Close()

	// Wait for unregister
	time.Sleep(100 * time.Millisecond)

	if hub.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients after disconnect, got %d", hub.GetClientCount())
	}
}
