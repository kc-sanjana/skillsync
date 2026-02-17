package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure properly in production
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	UserID      string
	Send        chan []byte
	messageRepo *repository.MessageRepository
}

type IncomingMessage struct {
	Type    string `json:"type"`    // join_room, leave_room, message
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, mr *repository.MessageRepository) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		UserID:      userID,
		Send:        make(chan []byte, 256),
		messageRepo: mr,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "join_room":
			c.hub.JoinRoom(msg.RoomID, c)
		case "leave_room":
			c.hub.LeaveRoom(msg.RoomID, c)
		case "message":
			dbMsg := &domain.Message{
				MatchID:  msg.RoomID,
				SenderID: c.UserID,
				Content:  msg.Content,
				Type:     "text",
			}
			if err := c.messageRepo.Create(context.Background(), dbMsg); err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}

			outgoing, _ := json.Marshal(map[string]any{
				"type":       "message",
				"id":         dbMsg.ID,
				"room_id":    dbMsg.MatchID,
				"sender_id":  dbMsg.SenderID,
				"content":    dbMsg.Content,
				"created_at": dbMsg.CreatedAt,
			})
			c.hub.Broadcast <- &RoomMessage{
				RoomID:  msg.RoomID,
				Message: outgoing,
				Sender:  c.UserID,
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
