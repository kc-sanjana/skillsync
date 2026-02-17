package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/middleware"
	"github.com/yourusername/skillsync/internal/service"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

type UpdateUserRequest struct {
	FullName    *string `json:"full_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	GithubURL   *string `json:"github_url"`
	LinkedinURL *string `json:"linkedin_url"`
}

type AddSkillRequest struct {
	SkillName   string  `json:"skill_name" validate:"required"`
	Proficiency string  `json:"proficiency" validate:"required,oneof=beginner intermediate advanced"`
	Years       float64 `json:"years_experience" validate:"gte=0"`
}

type PaginatedUsersResponse struct {
	Users  interface{} `json:"users"`
	Total  int64       `json:"total"`
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
	Pages  int         `json:"pages"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

// GetUsers handles GET /api/users?skills=go,python&level=advanced&page=1&limit=20
func (h *UserHandler) GetUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var skills []string
	if s := c.QueryParam("skills"); s != "" {
		skills = strings.Split(s, ",")
	}
	level := c.QueryParam("level")

	offset := (page - 1) * limit
	users, total, err := h.userService.SearchUsers(skills, level, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to search users"})
	}

	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}

	return c.JSON(http.StatusOK, PaginatedUsersResponse{
		Users: users,
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: pages,
	})
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch user"})
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /api/users/:id (protected - owner only)
func (h *UserHandler) UpdateUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
	}

	authUserID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if authUserID != id {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you can only update your own profile"})
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	updates := make(map[string]interface{})
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.GithubURL != nil {
		updates["github_url"] = *req.GithubURL
	}
	if req.LinkedinURL != nil {
		updates["linkedin_url"] = *req.LinkedinURL
	}

	if err := h.userService.UpdateProfile(id, updates); err != nil {
		if err == service.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile"})
	}

	user, _ := h.userService.GetUser(id)
	return c.JSON(http.StatusOK, user)
}

// AddUserSkill handles POST /api/users/:id/skills (protected - owner only)
func (h *UserHandler) AddUserSkill(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
	}

	authUserID, err := middleware.ExtractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}
	if authUserID != id {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "you can only add skills to your own profile"})
	}

	var req AddSkillRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if err := h.userService.AddSkill(id, req.SkillName, req.Proficiency, req.Years); err != nil {
		switch err {
		case service.ErrSkillExists:
			return c.JSON(http.StatusConflict, ErrorResponse{Error: "skill already added"})
		case service.ErrInvalidLevel:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to add skill"})
		}
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "skill added"})
}

// GetUserReputation handles GET /api/users/:id/reputation
func (h *UserHandler) GetUserReputation(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
	}

	user, err := h.userService.GetUserWithReputation(id)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch reputation"})
	}

	return c.JSON(http.StatusOK, user.Reputation)
}
