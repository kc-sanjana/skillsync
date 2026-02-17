package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/middleware"
	"github.com/yourusername/skillsync/internal/service"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

type SubmitRatingRequest struct {
	RatedID             string `json:"rated_id" validate:"required"`
	SessionID           uint   `json:"session_id" validate:"required"`
	OverallRating       int    `json:"overall_rating" validate:"required,min=1,max=5"`
	CodeQualityRating   int    `json:"code_quality_rating" validate:"required,min=1,max=5"`
	CommunicationRating int    `json:"communication_rating" validate:"required,min=1,max=5"`
	HelpfulnessRating   int    `json:"helpfulness_rating" validate:"required,min=1,max=5"`
	ReliabilityRating   int    `json:"reliability_rating" validate:"required,min=1,max=5"`
	Comment             string `json:"comment"`
}

type SubmitFeedbackRequest struct {
	Enjoyed          bool     `json:"enjoyed"`
	LearnedSomething bool     `json:"learned_something"`
	WouldPairAgain   bool     `json:"would_pair_again"`
	Strengths        []string `json:"strengths"`
	Improvements     []string `json:"improvements"`
	Rating           int      `json:"rating" validate:"required,min=1,max=5"`
	FeedbackText     string   `json:"feedback_text"`
}

type LeaderboardEntry struct {
	Rank       int                    `json:"rank"`
	User       *domain.User           `json:"user"`
	Reputation *domain.UserReputation `json:"reputation"`
}

type LeaderboardResponse struct {
	Category string              `json:"category"`
	Entries  []*LeaderboardEntry `json:"entries"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type ReputationHandler struct {
	repService *service.ReputationService
	db         *gorm.DB
}

func NewReputationHandler(rs *service.ReputationService, db *gorm.DB) *ReputationHandler {
	return &ReputationHandler{repService: rs, db: db}
}

// SubmitRating handles POST /api/ratings
func (h *ReputationHandler) SubmitRating(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req SubmitRatingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err = h.repService.SubmitRating(
		userID,
		req.RatedID,
		req.SessionID,
		req.OverallRating,
		req.CodeQualityRating,
		req.CommunicationRating,
		req.HelpfulnessRating,
		req.ReliabilityRating,
		req.Comment,
	)
	if err != nil {
		switch err {
		case service.ErrCannotRateSelf:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case service.ErrInvalidRating:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case service.ErrSessionNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case service.ErrNotSessionParticipant:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case service.ErrAlreadyRated:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to submit rating"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "rating submitted"})
}

// SubmitSessionFeedback handles POST /api/sessions/:id/feedback
func (h *ReputationHandler) SubmitSessionFeedback(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid session id"})
	}

	var req SubmitFeedbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	input := service.SessionFeedbackInput{
		Enjoyed:          req.Enjoyed,
		LearnedSomething: req.LearnedSomething,
		WouldPairAgain:   req.WouldPairAgain,
		Strengths:        req.Strengths,
		Improvements:     req.Improvements,
		Rating:           req.Rating,
		FeedbackText:     req.FeedbackText,
	}

	err = h.repService.SubmitSessionFeedback(uint(sessionID), userID, input)
	if err != nil {
		switch err {
		case service.ErrSessionNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case service.ErrNotSessionParticipant:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case service.ErrAlreadyGaveFeedback:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		case service.ErrInvalidRating:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to submit feedback"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "feedback submitted"})
}

// GetUserReputation handles GET /api/users/:id/reputation
func (h *ReputationHandler) GetUserReputation(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
	}

	var rep domain.UserReputation
	if err := h.db.Where("user_id = ?", id).First(&rep).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "reputation not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch reputation"})
	}

	// Decode skill credibility scores for a richer response.
	var skillScores map[string]interface{}
	if len(rep.SkillCredibilityScores) > 0 {
		json.Unmarshal(rep.SkillCredibilityScores, &skillScores)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"reputation":              rep,
		"skill_credibility_scores": skillScores,
	})
}

// GetMyRatings handles GET /api/ratings/received
func (h *ReputationHandler) GetMyRatings(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var ratings []domain.Rating
	if err := h.db.Preload("Rater").Preload("Session").
		Where("rated_id = ?", userID).
		Order("created_at DESC").
		Find(&ratings).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch ratings"})
	}

	// Calculate a quick summary.
	var totalOverall, totalCode, totalComm, totalHelp, totalReli int
	for _, r := range ratings {
		totalOverall += r.OverallRating
		totalCode += r.CodeQualityRating
		totalComm += r.CommunicationRating
		totalHelp += r.HelpfulnessRating
		totalReli += r.ReliabilityRating
	}

	count := len(ratings)
	var summary map[string]interface{}
	if count > 0 {
		summary = map[string]interface{}{
			"total_ratings":          count,
			"avg_overall":            float64(totalOverall) / float64(count),
			"avg_code_quality":       float64(totalCode) / float64(count),
			"avg_communication":      float64(totalComm) / float64(count),
			"avg_helpfulness":        float64(totalHelp) / float64(count),
			"avg_reliability":        float64(totalReli) / float64(count),
		}
	} else {
		summary = map[string]interface{}{
			"total_ratings":          0,
			"avg_overall":            0,
			"avg_code_quality":       0,
			"avg_communication":      0,
			"avg_helpfulness":        0,
			"avg_reliability":        0,
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ratings": ratings,
		"summary": summary,
	})
}

// GetLeaderboard handles GET /api/leaderboard?category=overall&limit=20
func (h *ReputationHandler) GetLeaderboard(c echo.Context) error {
	category := c.QueryParam("category")
	if category == "" {
		category = "overall"
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	contributors, err := h.repService.GetTopContributors(category, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch leaderboard"})
	}

	entries := make([]*LeaderboardEntry, len(contributors))
	for i, c := range contributors {
		entries[i] = &LeaderboardEntry{
			Rank:       i + 1,
			User:       &c.User,
			Reputation: c.Reputation,
		}
	}

	return c.JSON(http.StatusOK, LeaderboardResponse{
		Category: category,
		Entries:  entries,
	})
}
