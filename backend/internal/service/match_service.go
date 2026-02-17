package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
)

var (
	ErrMatchRequestExists  = errors.New("a pending match request already exists between these users")
	ErrMatchExists         = errors.New("an active match already exists between these users")
	ErrSelfMatch           = errors.New("cannot match with yourself")
	ErrRequestNotFound     = errors.New("match request not found")
	ErrNotRequestReceiver  = errors.New("only the receiver can accept or reject this request")
	ErrRequestNotPending   = errors.New("match request is no longer pending")
)

// MatchSuggestion is returned by FindMatches.
type MatchSuggestion struct {
	User                *domain.User    `json:"user"`
	MatchScore          float64         `json:"match_score"`
	AIInsights          *PairingInsights `json:"ai_insights,omitempty"`
	CommonSkills        []string        `json:"common_skills"`
	ComplementarySkills []string        `json:"complementary_skills"`
}

type MatchService struct {
	db      *gorm.DB
	claude  *ClaudeService
}

func NewMatchService(db *gorm.DB, claude *ClaudeService) *MatchService {
	return &MatchService{db: db, claude: claude}
}

// ---------------------------------------------------------------------------
// CalculateCompatibility
// ---------------------------------------------------------------------------

func (s *MatchService) CalculateCompatibility(user1ID, user2ID string) (float64, error) {
	if user1ID == user2ID {
		return 0, ErrSelfMatch
	}

	// Load both users with skills.
	var u1, u2 domain.User
	if err := s.db.Preload("Skills.Skill").First(&u1, "id = ?", user1ID).Error; err != nil {
		return 0, fmt.Errorf("user1 not found: %w", err)
	}
	if err := s.db.Preload("Skills.Skill").First(&u2, "id = ?", user2ID).Error; err != nil {
		return 0, fmt.Errorf("user2 not found: %w", err)
	}

	// Load reputations.
	var rep1, rep2 domain.UserReputation
	s.db.Where("user_id = ?", user1ID).First(&rep1)
	s.db.Where("user_id = ?", user2ID).First(&rep2)

	skillSim := skillSimilarity(u1.Skills, u2.Skills)
	goalsAlign := goalsAlignment(u1, u2)
	compSkills := complementaryScore(u1.Skills, u2.Skills)
	repCompat := reputationCompatibility(rep1, rep2)

	score := skillSim*0.40 + goalsAlign*0.30 + compSkills*0.20 + repCompat*0.10

	return math.Round(score*100) / 100, nil
}

// ---------------------------------------------------------------------------
// FindMatches
// ---------------------------------------------------------------------------

func (s *MatchService) FindMatches(userID string, limit int) ([]*MatchSuggestion, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Load the requesting user.
	var user domain.User
	if err := s.db.Preload("Skills.Skill").First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// IDs to exclude: self + existing active matches + pending outbound requests.
	excludeIDs := []string{userID}

	var matchedIDs []string
	s.db.Model(&domain.Match{}).
		Where("(user1_id = ? OR user2_id = ?) AND status = ?", userID, userID, domain.MatchActive).
		Select("CASE WHEN user1_id = ? THEN user2_id ELSE user1_id END", userID).
		Scan(&matchedIDs)
	excludeIDs = append(excludeIDs, matchedIDs...)

	var pendingIDs []string
	s.db.Model(&domain.MatchRequest{}).
		Where("sender_id = ? AND status = ?", userID, domain.RequestPending).
		Pluck("receiver_id", &pendingIDs)
	excludeIDs = append(excludeIDs, pendingIDs...)

	// Candidate pool: up to 5x the limit so we can score and rank.
	var candidates []domain.User
	s.db.Preload("Skills.Skill").
		Where("id NOT IN ?", excludeIDs).
		Limit(limit * 5).
		Find(&candidates)

	// Score every candidate.
	type scored struct {
		user  *domain.User
		score float64
	}
	results := make([]scored, 0, len(candidates))
	for i := range candidates {
		sc, err := s.CalculateCompatibility(userID, candidates[i].ID)
		if err != nil {
			continue
		}
		results = append(results, scored{user: &candidates[i], score: sc})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })

	if len(results) > limit {
		results = results[:limit]
	}

	// Build suggestions; enrich top 3 with AI insights.
	suggestions := make([]*MatchSuggestion, len(results))
	for i, r := range results {
		common, comp := classifySkills(user.Skills, r.user.Skills)
		suggestion := &MatchSuggestion{
			User:                r.user,
			MatchScore:          r.score,
			CommonSkills:        common,
			ComplementarySkills: comp,
		}

		if i < 3 && s.claude != nil {
			insights, err := s.claude.GeneratePairingInsights(user, *r.user, user.Skills, r.user.Skills)
			if err != nil {
				log.Warn().Err(err).Str("candidate", r.user.ID).Msg("failed to generate AI insights")
			} else {
				suggestion.AIInsights = insights
			}
		}

		suggestions[i] = suggestion
	}

	return suggestions, nil
}

// ---------------------------------------------------------------------------
// CreateMatchRequest
// ---------------------------------------------------------------------------

func (s *MatchService) CreateMatchRequest(senderID, receiverID string, message string) error {
	if senderID == receiverID {
		return ErrSelfMatch
	}

	// Check for existing pending request in either direction.
	var count int64
	s.db.Model(&domain.MatchRequest{}).
		Where(
			"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?",
			senderID, receiverID, receiverID, senderID, domain.RequestPending,
		).Count(&count)
	if count > 0 {
		return ErrMatchRequestExists
	}

	// Check for existing active match.
	s.db.Model(&domain.Match{}).
		Where(
			"((user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)) AND status = ?",
			senderID, receiverID, receiverID, senderID, domain.MatchActive,
		).Count(&count)
	if count > 0 {
		return ErrMatchExists
	}

	// Generate AI preview insights.
	var previewJSON domain.JSONB
	if s.claude != nil {
		var sender, receiver domain.User
		s.db.Preload("Skills.Skill").First(&sender, senderID)
		s.db.Preload("Skills.Skill").First(&receiver, receiverID)

		senderSkillNames := skillNames(sender.Skills)
		receiverSkillNames := skillNames(receiver.Skills)

		_, reasoning, err := s.claude.CalculateMatchScore(
			senderSkillNames, receiverSkillNames,
			sender.Bio, receiver.Bio,
		)
		if err != nil {
			log.Warn().Err(err).Msg("failed to generate AI preview for match request")
		} else {
			preview := map[string]string{"reasoning": reasoning}
			data, _ := json.Marshal(preview)
			previewJSON = domain.JSONB(data)
		}
	}
	if len(previewJSON) == 0 {
		previewJSON = domain.JSONB("{}")
	}

	req := domain.MatchRequest{
		SenderID:          senderID,
		ReceiverID:        receiverID,
		Status:            domain.RequestPending,
		Message:           message,
		AIPreviewInsights: previewJSON,
	}
	if err := s.db.Create(&req).Error; err != nil {
		return fmt.Errorf("failed to create match request: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// AcceptMatchRequest
// ---------------------------------------------------------------------------

func (s *MatchService) AcceptMatchRequest(requestID uint, userID string) (*domain.Match, error) {
	var req domain.MatchRequest
	if err := s.db.First(&req, "id = ?", requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRequestNotFound
		}
		return nil, fmt.Errorf("failed to fetch request: %w", err)
	}

	if req.ReceiverID != userID {
		return nil, ErrNotRequestReceiver
	}
	if req.Status != domain.RequestPending {
		return nil, ErrRequestNotPending
	}

	// Calculate compatibility score for the new match.
	score, _ := s.CalculateCompatibility(req.SenderID, req.ReceiverID)

	// Generate full AI insights.
	var insightsJSON domain.JSONB
	if s.claude != nil {
		var sender, receiver domain.User
		s.db.Preload("Skills.Skill").First(&sender, req.SenderID)
		s.db.Preload("Skills.Skill").First(&receiver, req.ReceiverID)

		insights, err := s.claude.GeneratePairingInsights(sender, receiver, sender.Skills, receiver.Skills)
		if err != nil {
			log.Warn().Err(err).Msg("failed to generate full AI insights on accept")
		} else {
			data, _ := json.Marshal(insights)
			insightsJSON = domain.JSONB(data)
		}
	}
	if len(insightsJSON) == 0 {
		insightsJSON = domain.JSONB("{}")
	}

	// Use a transaction: update request + create match.
	var match domain.Match
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&req).Updates(map[string]interface{}{
			"status":       domain.RequestAccepted,
			"responded_at": now,
		}).Error; err != nil {
			return err
		}

		match = domain.Match{
			User1ID:    req.SenderID,
			User2ID:    req.ReceiverID,
			MatchScore: score,
			AIInsights: insightsJSON,
			Status:     domain.MatchActive,
		}
		return tx.Create(&match).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to accept match request: %w", err)
	}

	// Re-load with relations.
	s.db.Preload("User1").Preload("User2").First(&match, match.ID)
	return &match, nil
}

// ---------------------------------------------------------------------------
// GetUserMatches
// ---------------------------------------------------------------------------

func (s *MatchService) GetUserMatches(userID string) ([]*domain.Match, error) {
	var matches []*domain.Match
	err := s.db.
		Preload("User1").Preload("User2").
		Where("(user1_id = ? OR user2_id = ?) AND status = ?", userID, userID, domain.MatchActive).
		Order("created_at DESC").
		Find(&matches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch matches: %w", err)
	}
	return matches, nil
}

// ---------------------------------------------------------------------------
// Scoring helpers
// ---------------------------------------------------------------------------

// skillSimilarity returns 0-100 based on how many skills overlap relative to
// the total distinct skill count (Jaccard similarity).
func skillSimilarity(s1, s2 []domain.UserSkill) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 50 // neutral when neither user has skills listed
	}

	set1 := make(map[uint]bool, len(s1))
	for _, sk := range s1 {
		set1[sk.SkillID] = true
	}
	set2 := make(map[uint]bool, len(s2))
	for _, sk := range s2 {
		set2[sk.SkillID] = true
	}

	var intersection int
	for id := range set1 {
		if set2[id] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 50
	}
	return (float64(intersection) / float64(union)) * 100
}

// goalsAlignment returns 0-100 based on bio length overlap as a rough proxy
// for goal compatibility. (In production this would call the AI service.)
func goalsAlignment(u1, u2 domain.User) float64 {
	if u1.Bio == "" && u2.Bio == "" {
		return 50
	}
	if u1.Bio == "" || u2.Bio == "" {
		return 30
	}
	// Rough heuristic: both bios present gives a baseline. In a production
	// system, the AI-powered CalculateMatchScore would handle this.
	return 65
}

// complementaryScore rewards users whose skills do NOT overlap, so they can
// teach each other.
func complementaryScore(s1, s2 []domain.UserSkill) float64 {
	if len(s1) == 0 || len(s2) == 0 {
		return 30
	}

	set1 := make(map[uint]bool, len(s1))
	for _, sk := range s1 {
		set1[sk.SkillID] = true
	}

	var unique int
	for _, sk := range s2 {
		if !set1[sk.SkillID] {
			unique++
		}
	}

	total := len(s2)
	return (float64(unique) / float64(total)) * 100
}

// reputationCompatibility returns 0-100 based on how close two users'
// reputation scores are (closer = better pairing experience).
func reputationCompatibility(r1, r2 domain.UserReputation) float64 {
	diff := math.Abs(r1.OverallScore - r2.OverallScore)
	if diff > 100 {
		diff = 100
	}
	return 100 - diff
}

// classifySkills splits skills into common and complementary lists.
func classifySkills(s1, s2 []domain.UserSkill) (common, complementary []string) {
	nameByID := make(map[uint]string)
	set1 := make(map[uint]bool)

	for _, sk := range s1 {
		set1[sk.SkillID] = true
		nameByID[sk.SkillID] = sk.Skill.Name
	}
	for _, sk := range s2 {
		nameByID[sk.SkillID] = sk.Skill.Name
	}

	seen := make(map[uint]bool)
	for _, sk := range s2 {
		if set1[sk.SkillID] {
			if !seen[sk.SkillID] {
				common = append(common, nameByID[sk.SkillID])
				seen[sk.SkillID] = true
			}
		} else {
			if !seen[sk.SkillID] {
				complementary = append(complementary, nameByID[sk.SkillID])
				seen[sk.SkillID] = true
			}
		}
	}
	// Also add user1's unique skills as complementary.
	for _, sk := range s1 {
		if !seen[sk.SkillID] {
			set2Has := false
			for _, s := range s2 {
				if s.SkillID == sk.SkillID {
					set2Has = true
					break
				}
			}
			if !set2Has {
				complementary = append(complementary, nameByID[sk.SkillID])
				seen[sk.SkillID] = true
			}
		}
	}

	if common == nil {
		common = []string{}
	}
	if complementary == nil {
		complementary = []string{}
	}
	return common, complementary
}

// skillNames extracts bare skill name strings from UserSkill relations.
func skillNames(skills []domain.UserSkill) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Skill.Name
	}
	return names
}
