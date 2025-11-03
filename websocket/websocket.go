package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// LogEvent represents a log event sent to WebSocket clients
type LogEvent struct {
	Type      string                 `json:"type"`      // intercept, modify, forward, attack, error, info
	Timestamp string                 `json:"timestamp"` // ISO 8601 format
	Level     string                 `json:"level"`     // info, warn, error, debug
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	conn *websocket.Conn
	send chan *LogEvent
	hub  *Hub
}

// Hub manages all WebSocket client connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *LogEvent
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for demo purposes
			// In production, you should validate the origin
			return true
		},
	}
)

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *LogEvent, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the WebSocket hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[websocket] Client connected (total: %d)", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[websocket] Client disconnected (total: %d)", len(h.clients))
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- event:
				default:
					// Client's send channel is full, close it
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a log event to all connected clients
func (h *Hub) Broadcast(event *LogEvent) {
	select {
	case h.broadcast <- event:
	default:
		// Broadcast channel is full, skip this event
		log.Printf("[websocket] Warning: broadcast channel full, dropping event")
	}
}

// BroadcastLog is a convenience method to broadcast a log message
func (h *Hub) BroadcastLog(level, eventType, message string, data map[string]interface{}) {
	event := &LogEvent{
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   message,
		Data:      data,
	}
	h.Broadcast(event)
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeWS handles WebSocket connection requests
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[websocket] Upgrade error: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan *LogEvent, 256),
		hub:  h,
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[websocket] Read error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send log event as JSON
			if err := c.conn.WriteJSON(event); err != nil {
				log.Printf("[websocket] Write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendWelcomeMessage sends a welcome message to a newly connected client
func (c *Client) SendWelcomeMessage() {
	welcome := &LogEvent{
		Type:      "info",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "info",
		Message:   "Connected to SAGE Gateway WebSocket",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"server":  "sage-gateway-infected-for-demo",
		},
	}

	select {
	case c.send <- welcome:
	default:
	}
}

// MarshalJSON marshals LogEvent to JSON for pretty printing
func (e *LogEvent) MarshalJSON() ([]byte, error) {
	type Alias LogEvent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}
