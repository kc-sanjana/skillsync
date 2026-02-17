package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/repository"
	ws "github.com/yourusername/skillsync/internal/websocket"
	"github.com/yourusername/skillsync/pkg/auth"
)

type WebSocketHandler struct {
	hub         *ws.Hub
	messageRepo *repository.MessageRepository
	jwt         *auth.JWTManager
}

func NewWebSocketHandler(hub *ws.Hub, mr *repository.MessageRepository, jwt *auth.JWTManager) *WebSocketHandler {
	return &WebSocketHandler{hub: hub, messageRepo: mr, jwt: jwt}
}

func (h *WebSocketHandler) HandleConnection(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing token"})
	}

	claims, err := h.jwt.Validate(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	conn, err := ws.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := ws.NewClient(h.hub, conn, claims.UserID, h.messageRepo)
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	return nil
}
