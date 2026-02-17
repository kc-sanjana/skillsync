package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (64 KB).
	maxMessageSize = 64 * 1024
)

// Client is a middleman between a single WebSocket connection and the Hub.
type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	UserID  string
	MatchID uint
	DB      *gorm.DB
	send    chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, matchID uint, db *gorm.DB) *Client {
	return &Client{
		Hub:     hub,
		Conn:    conn,
		UserID:  userID,
		MatchID: matchID,
		DB:      db,
		send:    make(chan []byte, 256),
	}
}

// ---------------------------------------------------------------------------
// Wire protocol
// ---------------------------------------------------------------------------

// InboundMessage is what the client sends over the socket.
type InboundMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ChatPayload is the data field for a "chat_message".
type ChatPayload struct {
	Content    string `json:"content"`
	ReceiverID string `json:"receiver_id"`
}

// TypingPayload is the data field for a "typing_indicator".
type TypingPayload struct {
	IsTyping bool `json:"is_typing"`
}

// CodeChangePayload is the data field for a "code_change".
type CodeChangePayload struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Cursor   int    `json:"cursor"`
}

// OutboundChatMessage is what gets broadcast for chat messages.
type OutboundChatMessage struct {
	Type      string         `json:"type"`
	Message   *domain.Message `json:"message"`
	Timestamp time.Time      `json:"timestamp"`
}

// OutboundTypingMessage is what gets broadcast for typing indicators.
type OutboundTypingMessage struct {
	Type     string `json:"type"`
	UserID   string `json:"user_id"`
	IsTyping bool   `json:"is_typing"`
}

// OutboundCodeChange is what gets broadcast for code changes.
type OutboundCodeChange struct {
	Type     string `json:"type"`
	UserID   string `json:"user_id"`
	Code     string `json:"code"`
	Language string `json:"language"`
	Cursor   int    `json:"cursor"`
}

// ---------------------------------------------------------------------------
// ReadPump
// ---------------------------------------------------------------------------

// ReadPump pumps messages from the WebSocket connection to the hub.
// It runs in its own goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Warn().Err(err).Str("user_id", c.UserID).Msg("ws unexpected close")
			}
			return
		}

		var msg InboundMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Warn().Err(err).Msg("ws bad json from client")
			continue
		}

		c.HandleMessage(msg.Type, msg.Data)
	}
}

// ---------------------------------------------------------------------------
// WritePump
// ---------------------------------------------------------------------------

// WritePump pumps messages from the hub to the WebSocket connection.
// A goroutine running WritePump is started for each client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain any queued messages into the same write.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// HandleMessage
// ---------------------------------------------------------------------------

func (c *Client) HandleMessage(msgType string, data json.RawMessage) {
	switch msgType {
	case "chat_message":
		c.handleChat(data)
	case "typing_indicator":
		c.handleTyping(data)
	case "code_change":
		c.handleCodeChange(data)
	default:
		log.Warn().Str("type", msgType).Msg("ws unknown message type")
	}
}

func (c *Client) handleChat(data json.RawMessage) {
	var payload ChatPayload
	if err := json.Unmarshal(data, &payload); err != nil || payload.Content == "" {
		return
	}

	// Determine receiver: for a 2-person match, the receiver is the other user.
	receiverID := payload.ReceiverID
	if receiverID == "" {
		var match domain.Match
		if err := c.DB.First(&match, c.MatchID).Error; err != nil {
			log.Error().Err(err).Msg("ws cannot find match")
			return
		}
		if match.User1ID == c.UserID {
			receiverID = match.User2ID
		} else {
			receiverID = match.User1ID
		}
	}

	// Persist to database.
	msg := domain.Message{
		SenderID:   c.UserID,
		ReceiverID: receiverID,
		MatchID:    c.MatchID,
		Content:    payload.Content,
	}
	if err := c.DB.Create(&msg).Error; err != nil {
		log.Error().Err(err).Msg("ws failed to persist message")
		return
	}

	// Preload sender for the broadcast payload.
	c.DB.Preload("Sender").First(&msg, msg.ID)

	out := OutboundChatMessage{
		Type:      "chat_message",
		Message:   &msg,
		Timestamp: msg.CreatedAt,
	}
	outBytes, _ := json.Marshal(out)
	c.Hub.BroadcastToMatch(c.MatchID, outBytes)
}

func (c *Client) handleTyping(data json.RawMessage) {
	var payload TypingPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	out := OutboundTypingMessage{
		Type:     "typing_indicator",
		UserID:   c.UserID,
		IsTyping: payload.IsTyping,
	}
	outBytes, _ := json.Marshal(out)
	c.Hub.BroadcastToMatch(c.MatchID, outBytes)
}

func (c *Client) handleCodeChange(data json.RawMessage) {
	var payload CodeChangePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	out := OutboundCodeChange{
		Type:     "code_change",
		UserID:   c.UserID,
		Code:     payload.Code,
		Language: payload.Language,
		Cursor:   payload.Cursor,
	}
	outBytes, _ := json.Marshal(out)
	c.Hub.BroadcastToMatch(c.MatchID, outBytes)
}
