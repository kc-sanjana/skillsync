package service

import (
	"context"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
)

type PairingInsightsService struct {
	claudeService *ClaudeService
	sessionRepo   *repository.SessionRepository
	matchRepo     *repository.MatchRepository
}

func NewPairingInsightsService(cs *ClaudeService, sr *repository.SessionRepository, mr *repository.MatchRepository) *PairingInsightsService {
	return &PairingInsightsService{claudeService: cs, sessionRepo: sr, matchRepo: mr}
}

func (s *PairingInsightsService) Analyze(ctx context.Context, matchID string) (*domain.PairingInsight, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	// Fetch both users via the match repo's user loader
	userA, err := s.matchRepo.GetUserByID(ctx, match.UserAID)
	if err != nil {
		return nil, err
	}
	userB, err := s.matchRepo.GetUserByID(ctx, match.UserBID)
	if err != nil {
		return nil, err
	}

	return s.claudeService.GeneratePairingInsights(ctx, userA, userB, match)
}
