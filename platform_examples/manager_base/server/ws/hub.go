package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Validate origin in production
	},
}

// Client represents a WebSocket client
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	channels map[string]bool
	hub      *Hub
	mu       sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Channel string      `json:"channel"`
	Event   interface{} `json:"event"`
}

// SubscribeMessage represents a subscription message
type SubscribeMessage struct {
	Action   string   `json:"action"`
	Channels []string `json:"channels"`
}

// Hub maintains active WebSocket clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	eventCh    chan *types.ServiceEvent
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		eventCh:    make(chan *types.ServiceEvent, 256),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	go h.processEvents()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("[Hub] Client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[Hub] Client disconnected, total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// processEvents processes service events and broadcasts them
func (h *Hub) processEvents() {
	for event := range h.eventCh {
		msg := &Message{
			Channel: "events",
			Event:   event,
		}
		h.broadcast <- msg
	}
}

// broadcastMessage sends a message to all subscribed clients
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Hub] Failed to marshal message: %v", err)
		return
	}

	for client := range h.clients {
		if client.channels[message.Channel] {
			select {
			case client.send <- data:
			default:
				// Client channel is full, disconnect
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
}

// Events returns the event channel
func (h *Hub) Events() chan<- *types.ServiceEvent {
	return h.eventCh
}

// HandleConnection handles a new WebSocket connection
func (h *Hub) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Hub] WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		channels: make(map[string]bool),
		hub:      h,
	}

	h.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Hub] Read error: %v", err)
			}
			break
		}

		// Parse subscription messages
		var msg SubscribeMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Action == "subscribe" {
				c.mu.Lock()
				for _, ch := range msg.Channels {
					c.channels[ch] = true
				}
				c.mu.Unlock()
				log.Printf("[Hub] Client subscribed to channels: %v", msg.Channels)
			} else if msg.Action == "unsubscribe" {
				c.mu.Lock()
				for _, ch := range msg.Channels {
					delete(c.channels, ch)
				}
				c.mu.Unlock()
				log.Printf("[Hub] Client unsubscribed from channels: %v", msg.Channels)
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("[Hub] Write error: %v", err)
				return
			}

		case <-c.done():
			return
		}
	}
}

// done returns a channel that signals when the client is done
func (c *Client) done() chan struct{} {
	return nil // TODO: Implement proper cleanup signal
}

// Close closes the client's connection
func (c *Client) Close() {
	c.conn.Close()
}

// GetChannels returns the channels the client is subscribed to
func (c *Client) GetChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	channels := make([]string, 0, len(c.channels))
	for ch := range c.channels {
		channels = append(channels, ch)
	}
	return channels
}
