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
	"github.com/yourusername/skillsync/internal/service"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

type SendMatchRequestReq struct {
	ReceiverID string `json:"receiver_id" validate:"required"`
	Message    string `json:"message"`
}

type MatchInsightsResponse struct {
	Match    *domain.Match           `json:"match"`
	Insights *service.PairingInsights `json:"insights"`
}

type CollaborationSuggestionsResponse struct {
	Projects []*service.ProjectSuggestion `json:"projects"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type MatchHandler struct {
	matchService  *service.MatchService
	claudeService *service.ClaudeService
	db            *gorm.DB
}

func NewMatchHandler(ms *service.MatchService, cs *service.ClaudeService, db *gorm.DB) *MatchHandler {
	return &MatchHandler{matchService: ms, claudeService: cs, db: db}
}

// GetMatchSuggestions handles GET /api/matches/suggestions?limit=10
func (h *MatchHandler) GetMatchSuggestions(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	suggestions, err := h.matchService.FindMatches(userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to find matches"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"suggestions": suggestions,
		"total":       len(suggestions),
	})
}

// SendMatchRequest handles POST /api/matches/request
func (h *MatchHandler) SendMatchRequest(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req SendMatchRequestReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if err := h.matchService.CreateMatchRequest(userID, req.ReceiverID, req.Message); err != nil {
		switch err {
		case service.ErrSelfMatch:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case service.ErrMatchRequestExists:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		case service.ErrMatchExists:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to send match request"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "match request sent"})
}

// AcceptMatchRequest handles PUT /api/matches/request/:id/accept
func (h *MatchHandler) AcceptMatchRequest(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request id"})
	}

	match, err := h.matchService.AcceptMatchRequest(uint(requestID), userID)
	if err != nil {
		switch err {
		case service.ErrRequestNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case service.ErrNotRequestReceiver:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case service.ErrRequestNotPending:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to accept match request"})
		}
	}

	return c.JSON(http.StatusOK, match)
}

// RejectMatchRequest handles PUT /api/matches/request/:id/reject
func (h *MatchHandler) RejectMatchRequest(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request id"})
	}

	var req domain.MatchRequest
	if err := h.db.First(&req, uint(requestID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match request not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch request"})
	}

	if req.ReceiverID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only the receiver can reject this request"})
	}
	if req.Status != domain.RequestPending {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "match request is no longer pending"})
	}

	now := time.Now()
	if err := h.db.Model(&req).Updates(map[string]interface{}{
		"status":       domain.RequestRejected,
		"responded_at": now,
	}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to reject match request"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "match request rejected"})
}

// GetMyMatches handles GET /api/matches
func (h *MatchHandler) GetMyMatches(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	matches, err := h.matchService.GetUserMatches(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch matches"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"matches": matches,
		"total":   len(matches),
	})
}

// GetPendingRequests handles GET /api/matches/requests/pending
func (h *MatchHandler) GetPendingRequests(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var received []domain.MatchRequest
	h.db.Preload("Sender").
		Where("receiver_id = ? AND status = ?", userID, domain.RequestPending).
		Order("created_at DESC").
		Find(&received)

	var sent []domain.MatchRequest
	h.db.Preload("Receiver").
		Where("sender_id = ? AND status = ?", userID, domain.RequestPending).
		Order("created_at DESC").
		Find(&sent)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"received": received,
		"sent":     sent,
	})
}

// GetMatchInsights handles GET /api/matches/:id/insights
func (h *MatchHandler) GetMatchInsights(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	matchID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match id"})
	}
	var match domain.Match
	if err := h.db.Preload("User1.Skills.Skill").Preload("User2.Skills.Skill").
		First(&match, uint(matchID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch match"})
	}

	// Ensure the requesting user is a participant.
	if match.User1ID != userID && match.User2ID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not a participant in this match"})
	}

	// Try to decode stored insights first.
	var insights service.PairingInsights
	if len(match.AIInsights) > 0 {
		if err := json.Unmarshal(match.AIInsights, &insights); err == nil && insights.OverallReasoning != "" {
			return c.JSON(http.StatusOK, MatchInsightsResponse{
				Match:    &match,
				Insights: &insights,
			})
		}
	}

	// Generate fresh insights if none are stored.
	fresh, err := h.claudeService.GeneratePairingInsights(
		match.User1, match.User2,
		match.User1.Skills, match.User2.Skills,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate insights"})
	}

	// Persist for next time.
	data, _ := json.Marshal(fresh)
	h.db.Model(&match).Update("ai_insights", domain.JSONB(data))

	return c.JSON(http.StatusOK, MatchInsightsResponse{
		Match:    &match,
		Insights: fresh,
	})
}

// GetCollaborationSuggestions handles GET /api/matches/:id/suggestions
func (h *MatchHandler) GetCollaborationSuggestions(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	matchID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid match id"})
	}

	var match domain.Match
	if err := h.db.Preload("User1.Skills.Skill").Preload("User2.Skills.Skill").
		First(&match, uint(matchID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "match not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch match"})
	}

	if match.User1ID != userID && match.User2ID != userID {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you are not a participant in this match"})
	}

	// Collect combined skills from both users.
	skillSet := make(map[string]bool)
	var bestLevel string
	for _, s := range match.User1.Skills {
		skillSet[s.Skill.Name] = true
		bestLevel = highestLevel(bestLevel, string(s.ProficiencyLevel))
	}
	for _, s := range match.User2.Skills {
		skillSet[s.Skill.Name] = true
		bestLevel = highestLevel(bestLevel, string(s.ProficiencyLevel))
	}

	combined := make([]string, 0, len(skillSet))
	for name := range skillSet {
		combined = append(combined, name)
	}
	if bestLevel == "" {
		bestLevel = "intermediate"
	}

	projects, err := h.claudeService.SuggestProjects(combined, bestLevel)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate collaboration suggestions"})
	}

	return c.JSON(http.StatusOK, CollaborationSuggestionsResponse{Projects: projects})
}

// highestLevel returns the higher of two proficiency levels.
func highestLevel(a, b string) string {
	rank := map[string]int{"beginner": 1, "intermediate": 2, "advanced": 3}
	if rank[b] > rank[a] {
		return b
	}
	return a
}
