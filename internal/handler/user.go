package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
	"github.com/yourusername/skillsync/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	ratingRepo  *repository.RatingRepository
	matchRepo   *repository.MatchRepository
}

func NewUserHandler(us *service.UserService, rr *repository.RatingRepository, mr *repository.MatchRepository) *UserHandler {
	return &UserHandler{userService: us, ratingRepo: rr, matchRepo: mr}
}

// UserProfileResponse is the enriched profile returned by GET /users/:id
type UserProfileResponse struct {
	domain.User
	AverageRating       *float64             `json:"average_rating"`
	ReputationBreakdown *ReputationBreakdown `json:"reputation_breakdown"`
	Skills              []SkillEntry         `json:"skills"`
	RecentRatings       []RatingEntry        `json:"recent_ratings"`
	TotalMatches        int                  `json:"total_matches"`
	SessionsCompleted   int                  `json:"sessions_completed"`
	SuccessRate         float64              `json:"success_rate"`
}

type ReputationBreakdown struct {
	CodeQuality   float64 `json:"code_quality"`
	Communication float64 `json:"communication"`
	Helpfulness   float64 `json:"helpfulness"`
	Reliability   float64 `json:"reliability"`
}

type SkillEntry struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	SkillID         string  `json:"skill_id"`
	CredibilityScore float64 `json:"credibility_score"`
	VerifiedByPeers bool    `json:"verified_by_peers"`
}

type RatingEntry struct {
	ID            string `json:"id"`
	SessionID     string `json:"session_id"`
	RaterID       string `json:"rater_id"`
	RateeID       string `json:"ratee_id"`
	OverallRating int    `json:"overall_rating"`
	Feedback      string `json:"feedback"`
}

func (h *UserHandler) List(c echo.Context) error {
	skill := c.QueryParam("skill")
	level := c.QueryParam("level")

	users, err := h.userService.List(c.Request().Context(), skill, level)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch users")
	}

	if users == nil {
		users = []domain.User{}
	}

	return successPaginated(c, http.StatusOK, users, len(users), 1, len(users)+1)
}

func (h *UserHandler) GetMe(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return fail(c, http.StatusUnauthorized, "Invalid token")
	}

	user, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch user")
	}

	if user == nil {
		return fail(c, http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetByID(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to fetch user")
	}
	if user == nil {
		return fail(c, http.StatusNotFound, "User not found")
	}

	profile := UserProfileResponse{
		User: *user,
	}

	// Reputation breakdown from ratings
	rep, err := h.ratingRepo.GetReputation(ctx, id)
	if err == nil && rep != nil && rep.TotalRatings > 0 {
		avgRating := rep.OverallScore
		profile.AverageRating = &avgRating

		// Map backend rating categories to frontend fields
		// Scale 1-5 ratings to 0-100 for the progress bars
		profile.ReputationBreakdown = &ReputationBreakdown{
			CodeQuality:   rep.AvgKnowledge * 20,    // knowledge → code quality (1-5 → 0-100)
			Communication: rep.AvgCommunication * 20,
			Helpfulness:   rep.AvgHelpfulness * 20,
			Reliability:   rep.OverallScore * 20,     // overall score as reliability proxy
		}
	}

	// Build skills list from skills_teach and skills_learn
	skills := make([]SkillEntry, 0)
	for i, s := range user.SkillsTeach {
		skills = append(skills, SkillEntry{
			ID:              user.ID + "_teach_" + s,
			UserID:          user.ID,
			SkillID:         s,
			CredibilityScore: 80, // default for teaching skills
			VerifiedByPeers: false,
		})
		_ = i
	}
	for _, s := range user.SkillsLearn {
		skills = append(skills, SkillEntry{
			ID:              user.ID + "_learn_" + s,
			UserID:          user.ID,
			SkillID:         s,
			CredibilityScore: 30, // default for learning skills
			VerifiedByPeers: false,
		})
	}
	profile.Skills = skills

	// Recent ratings
	recentRatings, err := h.ratingRepo.GetRecentByUser(ctx, id, 6)
	if err == nil {
		entries := make([]RatingEntry, 0, len(recentRatings))
		for _, r := range recentRatings {
			entries = append(entries, RatingEntry{
				ID:            r.ID,
				SessionID:     r.MatchID,
				RaterID:       r.RaterID,
				RateeID:       r.RatedUserID,
				OverallRating: r.Score,
				Feedback:      r.Comment,
			})
		}
		profile.RecentRatings = entries
	} else {
		profile.RecentRatings = make([]RatingEntry, 0)
	}

	// Match statistics
	totalMatches, err := h.matchRepo.CountByUser(ctx, id)
	if err == nil {
		profile.TotalMatches = totalMatches
	}

	completedMatches, err := h.matchRepo.CountCompletedByUser(ctx, id)
	if err == nil {
		profile.SessionsCompleted = completedMatches
		if totalMatches > 0 {
			profile.SuccessRate = float64(completedMatches) / float64(totalMatches)
		}
	}

	return success(c, http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var input service.UpdateProfileInput
	if err := c.Bind(&input); err != nil {
		return fail(c, http.StatusBadRequest, "Invalid request body")
	}

	user, err := h.userService.UpdateProfile(c.Request().Context(), userID, input)
	if err != nil {
		return fail(c, http.StatusInternalServerError, "Failed to update profile")
	}

	return success(c, http.StatusOK, user)
}
