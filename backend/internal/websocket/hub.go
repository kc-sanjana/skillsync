package websocket

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Hub maintains the set of active clients and broadcasts messages to clients
// that belong to the same match.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *OutboundMessage
}

// OutboundMessage wraps a payload with the target match so the hub can route
// it to the right clients.
type OutboundMessage struct {
	MatchID uint
	Data    []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *OutboundMessage, 256),
	}
}

// Run starts the hub's event loop. Call as a goroutine: go hub.Run()
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Info().
				Str("user_id", client.UserID).
				Uint("match_id", client.MatchID).
				Msg("ws client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Info().
				Str("user_id", client.UserID).
				Uint("match_id", client.MatchID).
				Msg("ws client unregistered")

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.MatchID != msg.MatchID {
					continue
				}
				select {
				case client.send <- msg.Data:
				default:
					// Client's send buffer is full; drop it.
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToMatch sends a message to every client connected to a match.
func (h *Hub) BroadcastToMatch(matchID uint, data []byte) {
	h.broadcast <- &OutboundMessage{MatchID: matchID, Data: data}
}

// Register queues a client for registration.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister queues a client for removal.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// OnlineUsersForMatch returns the user IDs currently connected to a match.
func (h *Hub) OnlineUsersForMatch(matchID uint) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var ids []string
	seen := make(map[string]bool)
	for client := range h.clients {
		if client.MatchID == matchID && !seen[client.UserID] {
			ids = append(ids, client.UserID)
			seen[client.UserID] = true
		}
	}
	return ids
}
