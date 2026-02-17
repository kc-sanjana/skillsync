package service

import (
	"context"
	"errors"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
)

type ReputationService struct {
	ratingRepo *repository.RatingRepository
	userRepo   *repository.UserRepository
}

func NewReputationService(rr *repository.RatingRepository, ur *repository.UserRepository) *ReputationService {
	return &ReputationService{ratingRepo: rr, userRepo: ur}
}

type RatingInput struct {
	MatchID       string
	RaterID       string
	RatedUserID   string
	Score         int
	Communication int
	Knowledge     int
	Helpfulness   int
	Comment       string
}

func (s *ReputationService) SubmitRating(ctx context.Context, input RatingInput) (*domain.Rating, error) {
	if input.Score < 1 || input.Score > 5 {
		return nil, errors.New("score must be between 1 and 5")
	}
	if input.RaterID == input.RatedUserID {
		return nil, errors.New("cannot rate yourself")
	}

	existing, _ := s.ratingRepo.FindByMatchAndRater(ctx, input.MatchID, input.RaterID)
	if existing != nil {
		return nil, errors.New("you have already rated this session")
	}

	rating := &domain.Rating{
		MatchID:       input.MatchID,
		RaterID:       input.RaterID,
		RatedUserID:   input.RatedUserID,
		Score:         input.Score,
		Communication: input.Communication,
		Knowledge:     input.Knowledge,
		Helpfulness:   input.Helpfulness,
		Comment:       input.Comment,
	}

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		return nil, err
	}

	if err := s.recalculateReputation(ctx, input.RatedUserID); err != nil {
		return nil, err
	}

	return rating, nil
}

func (s *ReputationService) GetReputation(ctx context.Context, userID string) (*domain.Reputation, error) {
	return s.ratingRepo.GetReputation(ctx, userID)
}

func (s *ReputationService) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	return s.ratingRepo.GetLeaderboard(ctx, limit)
}

func (s *ReputationService) recalculateReputation(ctx context.Context, userID string) error {
	rep, err := s.ratingRepo.GetReputation(ctx, userID)
	if err != nil {
		return err
	}

	badge := "newcomer"
	switch {
	case rep.TotalSessions >= 50 && rep.OverallScore >= 4.5:
		badge = "mentor"
	case rep.TotalSessions >= 20 && rep.OverallScore >= 4.0:
		badge = "expert"
	case rep.TotalSessions >= 5 && rep.OverallScore >= 3.5:
		badge = "rising_star"
	}

	return s.userRepo.UpdateReputation(ctx, userID, rep.OverallScore, badge)
}
