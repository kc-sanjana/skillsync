package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
)

type MatchHandler struct {
	matchService *service.MatchService
}

func NewMatchHandler(ms *service.MatchService) *MatchHandler {
	return &MatchHandler{matchService: ms}
}

func (h *MatchHandler) Create(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var input struct {
		TargetUserID string `json:"target_user_id"`
		SkillOffered string `json:"skill_offered"`
		SkillWanted  string `json:"skill_wanted"`
	}
	if err := c.Bind(&input); err != nil {
		return fail(c, http.StatusBadRequest, "Invalid request body")
	}

	match, err := h.matchService.CreateWithUsers(c.Request().Context(), userID, input.TargetUserID, input.SkillOffered, input.SkillWanted)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, http.StatusCreated, match)
}

func (h *MatchHandler) List(c echo.Context) error {
	userID := c.Get("user_id").(string)

	matches, err := h.matchService.ListByUserWithUsers(c.Request().Context(), userID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch matches")
	}

	return success(c, http.StatusOK, matches)
}

func (h *MatchHandler) GetByID(c echo.Context) error {
	id := c.Param("id")

	match, err := h.matchService.GetByIDWithUsers(c.Request().Context(), id)
	if err != nil {
		return fail(c, http.StatusNotFound, "Match not found")
	}

	return success(c, http.StatusOK, match)
}

func (h *MatchHandler) UpdateStatus(c echo.Context) error {
	id := c.Param("id")
	userID := c.Get("user_id").(string)

	var input struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&input); err != nil {
		return fail(c, http.StatusBadRequest, "Invalid request body")
	}

	match, err := h.matchService.UpdateStatus(c.Request().Context(), id, userID, input.Status)
	if err != nil {
		return fail(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, http.StatusOK, match)
}
