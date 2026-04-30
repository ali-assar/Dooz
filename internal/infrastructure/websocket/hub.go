package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"

	gorillaws "github.com/gorilla/websocket"
)

type MessageType string

const (
	TypeMove              MessageType = "move"
	TypeGameStart         MessageType = "game_start"
	TypeGameEnd           MessageType = "game_end"
	TypeChat              MessageType = "chat"
	TypePresence          MessageType = "presence"
	TypeMatchFound        MessageType = "match_found"
	TypeChallengeReceived MessageType = "challenge_received"
	TypeChallengeAccepted MessageType = "challenge_accepted"
	TypeChallengeRejected MessageType = "challenge_rejected"
	TypeChallengeCanceled MessageType = "challenge_canceled"
	TypeError             MessageType = "error"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	UserID string
	Conn   *gorillaws.Conn
	Send   chan []byte
}

type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
	mu         sync.RWMutex
	logger     *slog.Logger
}

type broadcastMsg struct {
	userIDs []string
	data    []byte
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan broadcastMsg, 256),
		logger:     logger.With("layer", "WebSocketHub"),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.logger.Info("client connected", "userID", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.UserID]; ok && existing == client {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Info("client disconnected", "userID", client.UserID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			for _, userID := range msg.userIDs {
				if client, ok := h.clients[userID]; ok {
					select {
					case client.Send <- msg.data:
					default:
						// slow client; drop
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToUsers broadcasts a typed message to a list of userIDs.
func (h *Hub) SendToUsers(userIDs []string, msgType MessageType, payload interface{}) {
	msg := Message{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal ws message", "error", err)
		return
	}
	h.broadcast <- broadcastMsg{userIDs: userIDs, data: data}
}

// IsOnline returns true if the user has an active WebSocket connection.
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// ConnectedUserIDs returns all currently connected user IDs.
func (h *Hub) ConnectedUserIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ids := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		ids = append(ids, userID)
	}
	return ids
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(gorillaws.TextMessage, msg); err != nil {
			return
		}
	}
}
