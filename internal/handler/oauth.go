package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/service"
	"github.com/yourusername/skillsync/pkg/auth"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
	jwt          *auth.JWTManager
}

func NewOAuthHandler(os *service.OAuthService, jwt *auth.JWTManager) *OAuthHandler {
	return &OAuthHandler{oauthService: os, jwt: jwt}
}

func oauthFrontendURL() string {
	u := os.Getenv("FRONTEND_URL")
	if u == "" {
		return "http://localhost:5173"
	}
	return u
}

func generateOAuthState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func setOAuthStateCookie(c echo.Context, name, value string) {
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
	state := generateOAuthState()
	setOAuthStateCookie(c, "oauth_state_google", state)
	return c.Redirect(http.StatusTemporaryRedirect, h.oauthService.GetGoogleLoginURL(state))
}

func (h *OAuthHandler) GoogleCallback(c echo.Context) error {
	cookie, err := c.Cookie("oauth_state_google")
	if err != nil || cookie.Value != c.QueryParam("state") {
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=invalid_state")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=no_code")
	}

	user, err := h.oauthService.HandleGoogleCallback(c.Request().Context(), code)
	if err != nil {
		log.Printf("[OAuth Google] callback error: %v", err)
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=oauth_failed")
	}

	token, err := h.jwt.Generate(user.ID, user.Email)
	if err != nil {
		log.Printf("[OAuth Google] token generation error: %v", err)
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/dashboard?token="+token)
}

// ---------------------------------------------------------------------------
// GitHub
// ---------------------------------------------------------------------------

func (h *OAuthHandler) GitHubLogin(c echo.Context) error {
	state := generateOAuthState()
	setOAuthStateCookie(c, "oauth_state_github", state)
	return c.Redirect(http.StatusTemporaryRedirect, h.oauthService.GetGitHubLoginURL(state))
}

func (h *OAuthHandler) GitHubCallback(c echo.Context) error {
	cookie, err := c.Cookie("oauth_state_github")
	if err != nil || cookie.Value != c.QueryParam("state") {
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=invalid_state")
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=no_code")
	}

	user, err := h.oauthService.HandleGitHubCallback(c.Request().Context(), code)
	if err != nil {
		log.Printf("[OAuth GitHub] callback error: %v", err)
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=oauth_failed")
	}

	token, err := h.jwt.Generate(user.ID, user.Email)
	if err != nil {
		log.Printf("[OAuth GitHub] token generation error: %v", err)
		return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/login?error=token_failed")
	}

	return c.Redirect(http.StatusTemporaryRedirect, oauthFrontendURL()+"/dashboard?token="+token)
}
