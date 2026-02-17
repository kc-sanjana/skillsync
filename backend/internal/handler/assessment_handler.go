package handler

import (
	"encoding/json"
	"net/http"
	"strings"
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

type SubmitCodeRequest struct {
	Code        string `json:"code" validate:"required"`
	Language    string `json:"language" validate:"required"`
	ChallengeID string `json:"challenge_id" validate:"required"`
}

type SubmitCodeResponse struct {
	Assessment *domain.Assessment          `json:"assessment"`
	Analysis   *service.CodeAnalysisResult `json:"analysis"`
}

type GetHintRequest struct {
	Code     string `json:"code" validate:"required"`
	Language string `json:"language" validate:"required"`
	Problem  string `json:"problem" validate:"required"`
}

type GetHintResponse struct {
	Hint string `json:"hint"`
}

type ProjectSuggestionsRequest struct {
	Skills     string `query:"skills" validate:"required"`
	SkillLevel string `query:"level"`
}

type ProjectSuggestionsResponse struct {
	Projects []*service.ProjectSuggestion `json:"projects"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type AssessmentHandler struct {
	claudeService *service.ClaudeService
	db            *gorm.DB
}

func NewAssessmentHandler(cs *service.ClaudeService, db *gorm.DB) *AssessmentHandler {
	return &AssessmentHandler{claudeService: cs, db: db}
}

// SubmitCode handles POST /api/assessments
func (h *AssessmentHandler) SubmitCode(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req SubmitCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	// Run AI analysis.
	analysis, err := h.claudeService.AnalyzeCode(req.Code, req.Language)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "code analysis failed"})
	}

	// Persist the assessment.
	assessment := domain.Assessment{
		UserID:        userID,
		ChallengeID:   req.ChallengeID,
		CodeSubmitted: req.Code,
		Language:      req.Language,
		AIScore:       float64(analysis.Score),
		SkillLevel:    analysis.SkillLevel,
		AIFeedback:    marshalJSONB(analysis),
		CompletedAt:   time.Now(),
	}
	if err := h.db.Create(&assessment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save assessment"})
	}

	return c.JSON(http.StatusCreated, SubmitCodeResponse{
		Assessment: &assessment,
		Analysis:   analysis,
	})
}

// GetHint handles POST /api/assessments/hint
func (h *AssessmentHandler) GetHint(c echo.Context) error {
	if _, err := middleware.ExtractUserID(c); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req GetHintRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	hint, err := h.claudeService.GenerateHint(req.Code, req.Language, req.Problem)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate hint"})
	}

	return c.JSON(http.StatusOK, GetHintResponse{Hint: hint})
}

// GetAssessmentHistory handles GET /api/assessments/history
func (h *AssessmentHandler) GetAssessmentHistory(c echo.Context) error {
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var assessments []domain.Assessment
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&assessments).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch assessment history"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"assessments": assessments,
		"total":       len(assessments),
	})
}

// GetProjectSuggestions handles GET /api/projects/suggestions?skills=go,python&level=intermediate
func (h *AssessmentHandler) GetProjectSuggestions(c echo.Context) error {
	if _, err := middleware.ExtractUserID(c); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	skillsParam := c.QueryParam("skills")
	if skillsParam == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "skills query parameter is required"})
	}

	skills := strings.Split(skillsParam, ",")
	level := c.QueryParam("level")
	if level == "" {
		level = "intermediate"
	}

	projects, err := h.claudeService.SuggestProjects(skills, level)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate project suggestions"})
	}

	return c.JSON(http.StatusOK, ProjectSuggestionsResponse{Projects: projects})
}

// marshalJSONB marshals any value into a domain.JSONB, falling back to "{}".
func marshalJSONB(v interface{}) domain.JSONB {
	data, err := json.Marshal(v)
	if err != nil {
		return domain.JSONB("{}")
	}
	return domain.JSONB(data)
}
