package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/middleware"
	ws "github.com/yourusername/skillsync/internal/websocket"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

type SendMessageRequest struct {
	MatchID    uint   `json:"match_id" validate:"required"`
	ReceiverID string `json:"receiver_id" validate:"required"`
	Content    string `json:"content" validate:"required"`
}

type MarkReadRequest struct {
	MessageIDs []uint `json:"message_ids" validate:"required,min=1"`
}

type MessageResponse struct {
	Messages []domain.Message `json:"messages"`
	Total    int64            `json:"total"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type MessageHandler struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewMessageHandler(db *gorm.DB, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{db: db, hub: hub}
}

// GetMessages handles GET /api/matches/:matchId/messages?page=1&limit=50
func (h *MessageHandler) GetMessages(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	matchID, err := strconv.ParseUint(c.Param("matchId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match id"})
	}

	// Verify participant.
	var match domain.Match
	if err := h.db.First(&match, uint(matchID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch match"})
	}
	if match.User1ID != userID && match.User2ID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not a participant in this match"})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int64
	h.db.Model(&domain.Message{}).Where("match_id = ?", uint(matchID)).Count(&total)

	var messages []domain.Message
	if err := h.db.Preload("Sender").
		Where("match_id = ?", uint(matchID)).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch messages"})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Messages: messages,
		Total:    total,
	})
}

// SendMessage handles POST /api/messages (REST fallback when WS is unavailable)
func (h *MessageHandler) SendMessage(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Verify the match and that the sender is a participant.
	var match domain.Match
	if err := h.db.First(&match, req.MatchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch match"})
	}
	if match.User1ID != userID && match.User2ID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not a participant in this match"})
	}

	msg := domain.Message{
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		MatchID:    req.MatchID,
		Content:    req.Content,
	}
	if err := h.db.Create(&msg).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to send message"})
	}

	// Preload sender for the response and broadcast.
	h.db.Preload("Sender").First(&msg, msg.ID)

	// Push through the WebSocket hub so connected clients get it in real time.
	if h.hub != nil {
		out := ws.OutboundChatMessage{
			Type:      "chat_message",
			Message:   &msg,
			Timestamp: msg.CreatedAt,
		}
		outBytes, _ := json.Marshal(out)
		h.hub.BroadcastToMatch(req.MatchID, outBytes)
	}

	return c.JSON(http.StatusCreated, msg)
}

// MarkMessagesRead handles PUT /api/messages/read
func (h *MessageHandler) MarkMessagesRead(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req MarkReadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Only mark messages where the authenticated user is the receiver.
	res := h.db.Model(&domain.Message{}).
		Where("id IN ? AND receiver_id = ? AND is_read = false", req.MessageIDs, userID).
		Update("is_read", true)
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark messages as read"})
	}

	// Broadcast read receipts through WebSocket so the sender can update UI.
	if h.hub != nil && res.RowsAffected > 0 {
		// Determine match_id from the first message.
		var sample domain.Message
		if h.db.Select("match_id").First(&sample, req.MessageIDs[0]).Error == nil {
			receipt := map[string]interface{}{
				"type":        "messages_read",
				"reader_id":   userID,
				"message_ids": req.MessageIDs,
				"timestamp":   time.Now(),
			}
			data, _ := json.Marshal(receipt)
			h.hub.BroadcastToMatch(sample.MatchID, data)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "messages marked as read",
		"updated_count": res.RowsAffected,
	})
}
