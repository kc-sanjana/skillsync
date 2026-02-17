package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
)

var (
	ErrSessionNotFound     = errors.New("coding session not found")
	ErrNotSessionParticipant = errors.New("user is not a participant in this session")
	ErrAlreadyRated        = errors.New("you have already rated this user for this session")
	ErrCannotRateSelf      = errors.New("you cannot rate yourself")
	ErrAlreadyGaveFeedback = errors.New("you have already submitted feedback for this session")
	ErrInvalidRating       = errors.New("ratings must be between 1 and 5")
)

// SessionFeedbackInput is the input DTO for SubmitSessionFeedback.
type SessionFeedbackInput struct {
	Enjoyed          bool     `json:"enjoyed"`
	LearnedSomething bool     `json:"learned_something"`
	WouldPairAgain   bool     `json:"would_pair_again"`
	Strengths        []string `json:"strengths"`
	Improvements     []string `json:"improvements"`
	Rating           int      `json:"rating" validate:"required,min=1,max=5"`
	FeedbackText     string   `json:"feedback_text"`
}

type ReputationService struct {
	db *gorm.DB
}

func NewReputationService(db *gorm.DB) *ReputationService {
	return &ReputationService{db: db}
}

// ---------------------------------------------------------------------------
// SubmitRating
// ---------------------------------------------------------------------------

func (s *ReputationService) SubmitRating(
	raterID, ratedID string, sessionID uint,
	overallRating, codeQuality, communication, helpfulness, reliability int,
	comment string,
) error {
	if raterID == ratedID {
		return ErrCannotRateSelf
	}

	// Validate ranges.
	for _, v := range []int{overallRating, codeQuality, communication, helpfulness, reliability} {
		if v < 1 || v > 5 {
			return ErrInvalidRating
		}
	}

	// Verify the session exists.
	var session domain.CodingSession
	if err := s.db.Preload("Match").First(&session, "id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to fetch session: %w", err)
	}

	// Verify both rater and rated are participants.
	match := session.Match
	participants := map[string]bool{match.User1ID: true, match.User2ID: true}
	if !participants[raterID] || !participants[ratedID] {
		return ErrNotSessionParticipant
	}

	// Check for duplicate rating.
	var exists int64
	s.db.Model(&domain.Rating{}).
		Where("rater_id = ? AND rated_id = ? AND session_id = ?", raterID, ratedID, sessionID).
		Count(&exists)
	if exists > 0 {
		return ErrAlreadyRated
	}

	rating := domain.Rating{
		RaterID:             raterID,
		RatedID:             ratedID,
		SessionID:           sessionID,
		OverallRating:       overallRating,
		CodeQualityRating:   codeQuality,
		CommunicationRating: communication,
		HelpfulnessRating:   helpfulness,
		ReliabilityRating:   reliability,
		Comment:             comment,
	}
	if err := s.db.Create(&rating).Error; err != nil {
		return fmt.Errorf("failed to save rating: %w", err)
	}

	// Recalculate the rated user's reputation in the background.
	go func() {
		if _, err := s.CalculateUserReputation(ratedID); err != nil {
			log.Error().Err(err).Str("user_id", ratedID).Msg("failed to recalculate reputation after rating")
		}
	}()

	return nil
}

// ---------------------------------------------------------------------------
// SubmitSessionFeedback
// ---------------------------------------------------------------------------

func (s *ReputationService) SubmitSessionFeedback(sessionID uint, userID string, input SessionFeedbackInput) error {
	if input.Rating < 1 || input.Rating > 5 {
		return ErrInvalidRating
	}

	// Verify session exists and user is a participant.
	var session domain.CodingSession
	if err := s.db.Preload("Match").First(&session, "id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to fetch session: %w", err)
	}

	match := session.Match
	if match.User1ID != userID && match.User2ID != userID {
		return ErrNotSessionParticipant
	}

	// No duplicate feedback.
	var exists int64
	s.db.Model(&domain.SessionFeedback{}).
		Where("session_id = ? AND user_id = ?", sessionID, userID).
		Count(&exists)
	if exists > 0 {
		return ErrAlreadyGaveFeedback
	}

	strengthsJSON, _ := json.Marshal(input.Strengths)
	improvementsJSON, _ := json.Marshal(input.Improvements)

	fb := domain.SessionFeedback{
		SessionID:        sessionID,
		UserID:           userID,
		Enjoyed:          input.Enjoyed,
		LearnedSomething: input.LearnedSomething,
		WouldPairAgain:   input.WouldPairAgain,
		Strengths:        domain.JSONB(strengthsJSON),
		Improvements:     domain.JSONB(improvementsJSON),
		Rating:           input.Rating,
		FeedbackText:     input.FeedbackText,
	}
	if err := s.db.Create(&fb).Error; err != nil {
		return fmt.Errorf("failed to save session feedback: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CalculateUserReputation
// ---------------------------------------------------------------------------

func (s *ReputationService) CalculateUserReputation(userID string) (*domain.UserReputation, error) {
	// Aggregate all ratings received by the user.
	var stats struct {
		Count         int64
		AvgOverall    float64
		AvgCode       float64
		AvgComm       float64
		AvgHelp       float64
		AvgReliable   float64
	}
	s.db.Model(&domain.Rating{}).
		Where("rated_id = ?", userID).
		Select(`
			COUNT(*)                         AS count,
			COALESCE(AVG(overall_rating),0)       AS avg_overall,
			COALESCE(AVG(code_quality_rating),0)  AS avg_code,
			COALESCE(AVG(communication_rating),0) AS avg_comm,
			COALESCE(AVG(helpfulness_rating),0)    AS avg_help,
			COALESCE(AVG(reliability_rating),0)    AS avg_reliable
		`).
		Scan(&stats)

	// Normalize 1-5 averages to 0-100 scale.
	codeScore := normalize(stats.AvgCode)
	commScore := normalize(stats.AvgComm)
	helpScore := normalize(stats.AvgHelp)
	reliScore := normalize(stats.AvgReliable)

	// Weighted overall: code(30%) + communication(30%) + helpfulness(20%) + reliability(20%)
	overall := codeScore*0.30 + commScore*0.30 + helpScore*0.20 + reliScore*0.20

	// Count completed sessions.
	var completedSessions int64
	s.db.Model(&domain.CodingSession{}).
		Joins("JOIN matches ON matches.id = coding_sessions.match_id").
		Where("(matches.user1_id = ? OR matches.user2_id = ?) AND coding_sessions.ended_at IS NOT NULL",
			userID, userID).
		Count(&completedSessions)

	// Count successful matches (active matches with at least one completed session).
	var successfulMatches int64
	s.db.Model(&domain.Match{}).
		Where("(user1_id = ? OR user2_id = ?) AND status = ?", userID, userID, domain.MatchActive).
		Where(`id IN (
			SELECT match_id FROM coding_sessions WHERE ended_at IS NOT NULL
		)`).
		Count(&successfulMatches)

	// Calculate per-skill credibility scores.
	skillCredibility := s.calculateSkillCredibility(userID)
	credJSON, _ := json.Marshal(skillCredibility)

	// Upsert reputation row.
	var rep domain.UserReputation
	err := s.db.Where("user_id = ?", userID).First(&rep).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rep = domain.UserReputation{UserID: userID}
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch reputation: %w", err)
	}

	rep.OverallScore = math.Round(overall*100) / 100
	rep.CodeQualityScore = math.Round(codeScore*100) / 100
	rep.CommunicationScore = math.Round(commScore*100) / 100
	rep.HelpfulnessScore = math.Round(helpScore*100) / 100
	rep.ReliabilityScore = math.Round(reliScore*100) / 100
	rep.TotalRatings = int(stats.Count)
	rep.AverageRating = math.Round(stats.AvgOverall*100) / 100
	rep.CompletedSessions = int(completedSessions)
	rep.SuccessfulMatches = int(successfulMatches)
	rep.SkillCredibilityScores = domain.JSONB(credJSON)

	if rep.ID == 0 {
		err = s.db.Create(&rep).Error
	} else {
		err = s.db.Save(&rep).Error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save reputation: %w", err)
	}

	// Sync the denormalized score on the user row.
	s.db.Model(&domain.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"reputation_score": rep.OverallScore,
			"total_sessions":   rep.CompletedSessions,
		})

	// Award badges.
	s.awardBadges(userID, &rep)

	return &rep, nil
}

// ---------------------------------------------------------------------------
// GetTopContributors
// ---------------------------------------------------------------------------

func (s *ReputationService) GetTopContributors(category string, limit int) ([]*UserWithReputation, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := s.db.Model(&domain.UserReputation{})

	// Allow filtering by score category.
	var orderCol string
	switch category {
	case "code_quality":
		orderCol = "code_quality_score"
	case "communication":
		orderCol = "communication_score"
	case "helpfulness":
		orderCol = "helpfulness_score"
	case "reliability":
		orderCol = "reliability_score"
	default:
		orderCol = "overall_score"
	}

	var reps []domain.UserReputation
	err := query.
		Where("total_ratings > 0").
		Order(orderCol + " DESC").
		Limit(limit).
		Find(&reps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top contributors: %w", err)
	}

	// Hydrate with user data.
	results := make([]*UserWithReputation, 0, len(reps))
	for i := range reps {
		var user domain.User
		if err := s.db.Preload("Skills.Skill").First(&user, reps[i].UserID).Error; err != nil {
			continue
		}
		rep := reps[i] // copy for safe pointer
		results = append(results, &UserWithReputation{
			User:       user,
			Reputation: &rep,
		})
	}

	return results, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// normalize converts a 1-5 average to a 0-100 score.
func normalize(avg float64) float64 {
	if avg <= 0 {
		return 0
	}
	return ((avg - 1) / 4) * 100
}

// SkillCredibilityScore stores the per-skill breakdown.
type SkillCredibilityScore struct {
	SkillName        string  `json:"skill_name"`
	AIAssessment     float64 `json:"ai_assessment"`
	PeerVerification float64 `json:"peer_verification"`
	SessionSuccess   float64 `json:"session_success"`
	Total            float64 `json:"total"`
}

// calculateSkillCredibility computes a credibility score for each of the
// user's skills:
//
//	(AI_assessments * 0.4) + (peer_verifications * 0.4) + (session_success * 0.2)
func (s *ReputationService) calculateSkillCredibility(userID string) map[string]SkillCredibilityScore {
	var userSkills []domain.UserSkill
	s.db.Preload("Skill").Where("user_id = ?", userID).Find(&userSkills)

	result := make(map[string]SkillCredibilityScore, len(userSkills))

	for _, us := range userSkills {
		// AI assessment component: average AI score from assessments in the
		// skill's language / challenge area (normalized to 0-100).
		var avgAI float64
		s.db.Model(&domain.Assessment{}).
			Where("user_id = ? AND language = ?", userID, us.Skill.Name).
			Select("COALESCE(AVG(ai_score), 0)").
			Scan(&avgAI)

		// Peer verification component: normalized count (cap at 10 verifications = 100).
		peerScore := math.Min(float64(us.VerifiedByPeers)/10*100, 100)

		// Session success component: average success_rating across all sessions
		// the user participated in (normalized 0-100, already 0-1 in DB * 100).
		var avgSession float64
		s.db.Model(&domain.CodingSession{}).
			Joins("JOIN matches ON matches.id = coding_sessions.match_id").
			Where("(matches.user1_id = ? OR matches.user2_id = ?) AND coding_sessions.ended_at IS NOT NULL",
				userID, userID).
			Select("COALESCE(AVG(coding_sessions.success_rating), 0)").
			Scan(&avgSession)
		sessionScore := avgSession * 100

		total := avgAI*0.4 + peerScore*0.4 + sessionScore*0.2
		total = math.Round(total*100) / 100

		result[us.Skill.Name] = SkillCredibilityScore{
			SkillName:        us.Skill.Name,
			AIAssessment:     math.Round(avgAI*100) / 100,
			PeerVerification: math.Round(peerScore*100) / 100,
			SessionSuccess:   math.Round(sessionScore*100) / 100,
			Total:            total,
		}
	}

	return result
}

// awardBadges checks badge criteria and updates the user's badges JSONB.
func (s *ReputationService) awardBadges(userID string, rep *domain.UserReputation) {
	type badge struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var badges []badge

	if rep.OverallScore >= 90 {
		badges = append(badges, badge{
			Name:        "Top Contributor",
			Description: "Maintained an overall reputation score above 90",
		})
	}
	if rep.CodeQualityScore >= 95 {
		badges = append(badges, badge{
			Name:        "Code Master",
			Description: "Achieved a code quality score above 95",
		})
	}
	if rep.CompletedSessions >= 50 {
		badges = append(badges, badge{
			Name:        "Session Guru",
			Description: "Completed over 50 pair-programming sessions",
		})
	}
	if rep.SuccessfulMatches >= 25 {
		badges = append(badges, badge{
			Name:        "Networking Pro",
			Description: "Successfully matched with over 25 developers",
		})
	}
	if rep.AverageRating >= 4.8 && rep.TotalRatings >= 10 {
		badges = append(badges, badge{
			Name:        "Highly Rated",
			Description: "Maintained a 4.8+ average rating with at least 10 reviews",
		})
	}

	if len(badges) == 0 {
		badges = []badge{}
	}

	data, _ := json.Marshal(badges)
	s.db.Model(&domain.User{}).Where("id = ?", userID).
		Update("badges", domain.JSONB(data))
}
