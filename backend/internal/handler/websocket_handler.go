package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/pkg/auth"
	ws "github.com/yourusername/skillsync/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, restrict this to your frontend's origin.
		return true
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
	db  *gorm.DB
}

func NewWebSocketHandler(hub *ws.Hub, db *gorm.DB) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, db: db}
}

// HandleWebSocket handles GET /ws?token=xxx&match_id=1
//
// Flow:
//  1. Read token + match_id from query params
//  2. Validate JWT
//  3. Verify user is a participant in the match
//  4. Upgrade to WebSocket
//  5. Create Client, register with Hub, start read/write pumps
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	// --- authenticate via query param (WebSocket can't send headers) ---
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing token query parameter"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
	}
	userID := claims.UserID

	// --- parse match_id ---
	matchIDStr := c.QueryParam("match_id")
	matchID64, err2 := strconv.ParseUint(matchIDStr, 10, 64)
	if err2 != nil || matchID64 == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid or missing match_id"})
	}
	matchID := uint(matchID64)

	// --- verify the match exists and the user is a participant ---
	var match domain.Match
	if err := h.db.First(&match, "id = ?", matchID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch match"})
	}
	if match.User1ID != userID && match.User2ID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not a participant in this match"})
	}
	if match.Status != domain.MatchActive {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "match is not active"})
	}

	// --- upgrade to WebSocket ---
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Error().Err(err).Msg("ws upgrade failed")
		return nil // Upgrade already wrote an HTTP error
	}

	client := ws.NewClient(h.hub, conn, userID, matchID, h.db)
	h.hub.Register(client)

	// Start pumps in their own goroutines.
	go client.WritePump()
	go client.ReadPump()

	return nil
}
