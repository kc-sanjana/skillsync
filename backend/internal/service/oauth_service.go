package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
)

type OAuthService struct {
	db          *gorm.DB
	userService *UserService
}

func NewOAuthService(db *gorm.DB, userService *UserService) *OAuthService {
	return &OAuthService{db: db, userService: userService}
}

// ---------------------------------------------------------------------------
// Google
// ---------------------------------------------------------------------------

func (s *OAuthService) GetGoogleLoginURL(state string) string {
	params := url.Values{
		"client_id":     {os.Getenv("GOOGLE_CLIENT_ID")},
		"redirect_uri":  {os.Getenv("OAUTH_REDIRECT_BASE") + "/api/auth/google/callback"},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (s *OAuthService) HandleGoogleCallback(code string) (*domain.User, error) {
	// Exchange code for token.
	tokenResp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("GOOGLE_CLIENT_ID")},
		"client_secret": {os.Getenv("GOOGLE_CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("OAUTH_REDIRECT_BASE") + "/api/auth/google/callback"},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("google token exchange failed: %w", err)
	}
	defer tokenResp.Body.Close()

	body, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google token exchange returned %d: %s", tokenResp.StatusCode, string(body))
	}

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to parse google token response: %w", err)
	}

	// Fetch user profile.
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var profile struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse google profile: %w", err)
	}

	return s.userService.FindOrCreateOAuthUser("google", profile.ID, profile.Email, profile.Name, profile.Picture)
}

// ---------------------------------------------------------------------------
// GitHub
// ---------------------------------------------------------------------------

func (s *OAuthService) GetGitHubLoginURL(state string) string {
	params := url.Values{
		"client_id":    {os.Getenv("GITHUB_CLIENT_ID")},
		"redirect_uri": {os.Getenv("OAUTH_REDIRECT_BASE") + "/api/auth/github/callback"},
		"scope":        {"user:email read:user"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (s *OAuthService) HandleGitHubCallback(code string) (*domain.User, error) {
	// Exchange code for token.
	data := url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("GITHUB_CLIENT_ID")},
		"client_secret": {os.Getenv("GITHUB_CLIENT_SECRET")},
		"redirect_uri":  {os.Getenv("OAUTH_REDIRECT_BASE") + "/api/auth/github/callback"},
	}

	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenData struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to parse github token response: %w", err)
	}
	if tokenData.Error != "" {
		return nil, fmt.Errorf("github token error: %s", tokenData.Error)
	}

	// Fetch user profile.
	profileReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	profileReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
	profileResp, err := http.DefaultClient.Do(profileReq)
	if err != nil {
		return nil, fmt.Errorf("github user request failed: %w", err)
	}
	defer profileResp.Body.Close()

	body, _ = io.ReadAll(profileResp.Body)
	var profile struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse github profile: %w", err)
	}

	// GitHub may not return email in profile — fetch from emails endpoint.
	email := profile.Email
	if email == "" {
		email, _ = s.fetchGitHubPrimaryEmail(tokenData.AccessToken)
	}

	name := profile.Name
	if name == "" {
		name = profile.Login
	}

	return s.userService.FindOrCreateOAuthUser("github", fmt.Sprintf("%d", profile.ID), email, name, profile.AvatarURL)
}

func (s *OAuthService) fetchGitHubPrimaryEmail(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
