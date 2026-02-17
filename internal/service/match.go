package service

import (
	"context"
	"errors"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
)

type MatchService struct {
	matchRepo    *repository.MatchRepository
	userRepo     *repository.UserRepository
	claudeService *ClaudeService
}

func NewMatchService(mr *repository.MatchRepository, ur *repository.UserRepository, cs *ClaudeService) *MatchService {
	return &MatchService{matchRepo: mr, userRepo: ur, claudeService: cs}
}

func (s *MatchService) Create(ctx context.Context, userAID, userBID, skillOffered, skillWanted string) (*domain.Match, error) {
	if userAID == userBID {
		return nil, errors.New("cannot match with yourself")
	}

	userA, err := s.userRepo.FindByID(ctx, userAID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userB, err := s.userRepo.FindByID(ctx, userBID)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	score := calculateMatchScore(userA, userB, skillOffered, skillWanted)

	match := &domain.Match{
		UserAID:      userAID,
		UserBID:      userBID,
		SkillOffered: skillOffered,
		SkillWanted:  skillWanted,
		Status:       "pending",
		MatchScore:   score,
	}

	if err := s.matchRepo.Create(ctx, match); err != nil {
		return nil, err
	}

	return match, nil
}

func (s *MatchService) ListByUser(ctx context.Context, userID string) ([]domain.Match, error) {
	return s.matchRepo.ListByUser(ctx, userID)
}

func (s *MatchService) GetByID(ctx context.Context, id string) (*domain.Match, error) {
	return s.matchRepo.FindByID(ctx, id)
}

func (s *MatchService) UpdateStatus(ctx context.Context, matchID, userID, status string) (*domain.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, errors.New("match not found")
	}

	if match.UserBID != userID && match.UserAID != userID {
		return nil, errors.New("not authorized to update this match")
	}

	validTransitions := map[string][]string{
		"pending":  {"accepted", "rejected"},
		"accepted": {"completed"},
	}

	allowed := false
	for _, valid := range validTransitions[match.Status] {
		if valid == status {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("invalid status transition")
	}

	match.Status = status
	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}

	return match, nil
}

// MatchWithUsers is the response format the frontend expects.
type MatchWithUsers struct {
	ID           string       `json:"id"`
	User1        *domain.User `json:"user1"`
	User2        *domain.User `json:"user2"`
	SkillOffered string       `json:"skill_offered"`
	SkillWanted  string       `json:"skill_wanted"`
	Status       string       `json:"status"`
	MatchScore   float64      `json:"match_score"`
	CreatedAt    any          `json:"created_at"`
	UpdatedAt    any          `json:"updated_at"`
}

func (s *MatchService) enrichMatch(ctx context.Context, m *domain.Match) (*MatchWithUsers, error) {
	user1, _ := s.userRepo.FindByID(ctx, m.UserAID)
	user2, _ := s.userRepo.FindByID(ctx, m.UserBID)
	return &MatchWithUsers{
		ID:           m.ID,
		User1:        user1,
		User2:        user2,
		SkillOffered: m.SkillOffered,
		SkillWanted:  m.SkillWanted,
		Status:       m.Status,
		MatchScore:   m.MatchScore,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func (s *MatchService) CreateWithUsers(ctx context.Context, userAID, userBID, skillOffered, skillWanted string) (*MatchWithUsers, error) {
	match, err := s.Create(ctx, userAID, userBID, skillOffered, skillWanted)
	if err != nil {
		return nil, err
	}
	return s.enrichMatch(ctx, match)
}

func (s *MatchService) ListByUserWithUsers(ctx context.Context, userID string) ([]MatchWithUsers, error) {
	matches, err := s.matchRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]MatchWithUsers, 0, len(matches))
	for i := range matches {
		enriched, err := s.enrichMatch(ctx, &matches[i])
		if err != nil {
			continue
		}
		result = append(result, *enriched)
	}
	return result, nil
}

func (s *MatchService) GetByIDWithUsers(ctx context.Context, id string) (*MatchWithUsers, error) {
	match, err := s.matchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.enrichMatch(ctx, match)
}

func calculateMatchScore(a, b *domain.User, offered, wanted string) float64 {
	score := 50.0

	for _, s := range a.SkillsTeach {
		if s == offered {
			score += 15
			break
		}
	}
	for _, s := range b.SkillsLearn {
		if s == offered {
			score += 15
			break
		}
	}
	for _, s := range b.SkillsTeach {
		if s == wanted {
			score += 10
			break
		}
	}

	score += (a.ReputationScore + b.ReputationScore) / 2 * 0.1

	if score > 100 {
		score = 100
	}
	return score
}
