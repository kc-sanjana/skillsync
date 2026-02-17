package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
	"github.com/yourusername/skillsync/pkg/auth"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
}

func NewOAuthHandler(os *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: os}
}

func frontendURL() string {
	u := os.Getenv("FRONTEND_URL")
	if u == "" {
		return "http://localhost:5173"
	}
	return u
}

// generateState creates a random hex string for CSRF protection.
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// setStateCookie stores the OAuth state parameter in a short-lived cookie.
func setStateCookie(c echo.Context, name, value string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(5 * time.Minute / time.Second),
	})
}

// ---------------------------------------------------------------------------
// Google
// ---------------------------------------------------------------------------

func (h *OAuthHandler) GoogleLogin(c echo.Context) error {
	state := generateState()
	setStateCookie(c, "oauth_state_google", state)
	return c.Redirect(http.StatusTemporaryRedirect, h.oauthService.GetGoogleLoginURL(state))
}

func (h *OAuthHandler) GoogleCallback(c echo.Context) error {
	// Validate state.
	cookie, err := c.Cookie("oauth_state_google")
	if err != nil || cookie.Value != c.QueryParam("state") {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=invalid_state")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=no_code")
	}

	user, err := h.oauthService.HandleGoogleCallback(code)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=oauth_failed")
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/dashboard?token="+token)
}

// ---------------------------------------------------------------------------
// GitHub
// ---------------------------------------------------------------------------

func (h *OAuthHandler) GitHubLogin(c echo.Context) error {
	state := generateState()
	setStateCookie(c, "oauth_state_github", state)
	return c.Redirect(http.StatusTemporaryRedirect, h.oauthService.GetGitHubLoginURL(state))
}

func (h *OAuthHandler) GitHubCallback(c echo.Context) error {
	cookie, err := c.Cookie("oauth_state_github")
	if err != nil || cookie.Value != c.QueryParam("state") {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=invalid_state")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=no_code")
	}

	user, err := h.oauthService.HandleGitHubCallback(code)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=oauth_failed")
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, frontendURL()+"/dashboard?token="+token)
}
