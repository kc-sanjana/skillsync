package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
	"github.com/yourusername/skillsync/pkg/auth"
)

type AuthHandler struct {
	userService *service.UserService
	jwt         *auth.JWTManager
}

func NewAuthHandler(us *service.UserService, jwt *auth.JWTManager) *AuthHandler {
	return &AuthHandler{userService: us, jwt: jwt}
}

type registerRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Username string   `json:"username" validate:"required,min=3,max=30"`
	Password string   `json:"password" validate:"required,min=8"`
	FullName string   `json:"full_name" validate:"required"`
	SkillsTeach []string `json:"skills_teach"`
	SkillsLearn []string `json:"skills_learn"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type authResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	user, err := h.userService.Register(c.Request().Context(), service.RegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		FullName:    req.FullName,
		SkillsTeach: req.SkillsTeach,
		SkillsLearn: req.SkillsLearn,
	})
	if err != nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
	}

	token, err := h.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	user, err := h.userService.Authenticate(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, err := h.jwt.Generate(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	oldToken := c.Request().Header.Get("Authorization")
	if oldToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing token"})
	}

	newToken, err := h.jwt.Refresh(oldToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": newToken})
}
