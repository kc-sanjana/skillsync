package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
)

type AssessmentHandler struct {
	claudeService *service.ClaudeService
	userService   *service.UserService
}

func NewAssessmentHandler(cs *service.ClaudeService, us *service.UserService) *AssessmentHandler {
	return &AssessmentHandler{claudeService: cs, userService: us}
}

type assessmentRequest struct {
	Skill   string   `json:"skill" validate:"required"`
	Answers []string `json:"answers" validate:"required"`
}

func (h *AssessmentHandler) Evaluate(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req assessmentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	assessment, err := h.claudeService.EvaluateSkill(c.Request().Context(), userID, req.Skill, req.Answers)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Assessment failed"})
	}

	if err := h.userService.UpdateSkillLevel(c.Request().Context(), userID, req.Skill, assessment.Level); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update skill level"})
	}

	return success(c, http.StatusOK, assessment)
}
