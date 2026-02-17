package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
)

type ReputationHandler struct {
	reputationService *service.ReputationService
}

func NewReputationHandler(rs *service.ReputationService) *ReputationHandler {
	return &ReputationHandler{reputationService: rs}
}

func (h *ReputationHandler) GetMyReputation(c echo.Context) error {
	userID := c.Get("user_id").(string)

	rep, err := h.reputationService.GetReputation(c.Request().Context(), userID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch reputation")
	}

	return success(c, http.StatusOK, rep)
}

func (h *ReputationHandler) SubmitRating(c echo.Context) error {
	raterID := c.Get("user_id").(string)

	var input struct {
		MatchID       string `json:"match_id"`
		RatedUserID   string `json:"rated_user_id"`
		Score         int    `json:"score"`
		Communication int    `json:"communication"`
		Knowledge     int    `json:"knowledge"`
		Helpfulness   int    `json:"helpfulness"`
		Comment       string `json:"comment"`
	}
	if err := c.Bind(&input); err != nil {
		return fail(c, http.StatusBadRequest, "Invalid request body")
	}

	rating, err := h.reputationService.SubmitRating(c.Request().Context(), service.RatingInput{
		MatchID:       input.MatchID,
		RaterID:       raterID,
		RatedUserID:   input.RatedUserID,
		Score:         input.Score,
		Communication: input.Communication,
		Knowledge:     input.Knowledge,
		Helpfulness:   input.Helpfulness,
		Comment:       input.Comment,
	})
	if err != nil {
		return fail(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, http.StatusCreated, rating)
}

func (h *ReputationHandler) Leaderboard(c echo.Context) error {
	limit := 20

	entries, err := h.reputationService.GetLeaderboard(c.Request().Context(), limit)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch leaderboard")
	}

	return successPaginated(c, http.StatusOK, entries, len(entries), 1, limit)
}
