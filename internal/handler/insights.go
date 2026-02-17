package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
)

type InsightsHandler struct {
	pairingService *service.PairingInsightsService
}

func NewInsightsHandler(ps *service.PairingInsightsService) *InsightsHandler {
	return &InsightsHandler{pairingService: ps}
}

func (h *InsightsHandler) GetPairingInsights(c echo.Context) error {
	matchID := c.Param("matchId")

	insights, err := h.pairingService.Analyze(c.Request().Context(), matchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate insights"})
	}

	return success(c, http.StatusOK, insights)
}
