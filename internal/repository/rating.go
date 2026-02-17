package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/skillsync/internal/domain"
)

type RatingRepository struct {
	db *sql.DB
}

func NewRatingRepository(db *sql.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Create(ctx context.Context, rating *domain.Rating) error {
	query := `INSERT INTO ratings (match_id, rater_id, rated_user_id, score, communication, knowledge, helpfulness, comment)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		rating.MatchID, rating.RaterID, rating.RatedUserID, rating.Score,
		rating.Communication, rating.Knowledge, rating.Helpfulness, rating.Comment,
	).Scan(&rating.ID, &rating.CreatedAt)
}

func (r *RatingRepository) FindByMatchAndRater(ctx context.Context, matchID, raterID string) (*domain.Rating, error) {
	var rating domain.Rating
	query := `SELECT id, match_id, rater_id, rated_user_id, score, created_at
	          FROM ratings WHERE match_id = $1 AND rater_id = $2`
	err := r.db.QueryRowContext(ctx, query, matchID, raterID).Scan(
		&rating.ID, &rating.MatchID, &rating.RaterID, &rating.RatedUserID,
		&rating.Score, &rating.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *RatingRepository) GetReputation(ctx context.Context, userID string) (*domain.Reputation, error) {
	var rep domain.Reputation
	query := `SELECT
	            $1 as user_id,
	            COALESCE(AVG(score), 0) as overall_score,
	            COUNT(*) as total_ratings,
	            COUNT(DISTINCT match_id) as total_sessions,
	            COALESCE(AVG(communication), 0) as avg_communication,
	            COALESCE(AVG(knowledge), 0) as avg_knowledge,
	            COALESCE(AVG(helpfulness), 0) as avg_helpfulness
	          FROM ratings WHERE rated_user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&rep.UserID, &rep.OverallScore, &rep.TotalRatings, &rep.TotalSessions,
		&rep.AvgCommunication, &rep.AvgKnowledge, &rep.AvgHelpfulness,
	)
	if err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *RatingRepository) GetRecentByUser(ctx context.Context, userID string, limit int) ([]domain.Rating, error) {
	query := `SELECT id, match_id, rater_id, rated_user_id, score,
	            COALESCE(communication, 0), COALESCE(knowledge, 0), COALESCE(helpfulness, 0),
	            COALESCE(comment, ''), created_at
	          FROM ratings WHERE rated_user_id = $1
	          ORDER BY created_at DESC LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := make([]domain.Rating, 0)
	for rows.Next() {
		var rt domain.Rating
		if err := rows.Scan(&rt.ID, &rt.MatchID, &rt.RaterID, &rt.RatedUserID,
			&rt.Score, &rt.Communication, &rt.Knowledge, &rt.Helpfulness,
			&rt.Comment, &rt.CreatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, rt)
	}
	return ratings, nil
}

func (r *RatingRepository) GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error) {
	query := `SELECT u.id, u.username, COALESCE(u.avatar_url, ''), COALESCE(u.reputation_score, 0),
	            COUNT(DISTINCT r.match_id) as total_sessions
	          FROM users u
	          LEFT JOIN ratings r ON r.rated_user_id = u.id
	          GROUP BY u.id, u.username, u.avatar_url, u.reputation_score
	          ORDER BY u.reputation_score DESC
	          LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.LeaderboardEntry, 0)
	rank := 1
	for rows.Next() {
		var e domain.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.AvatarURL, &e.OverallScore, &e.TotalSessions); err != nil {
			return nil, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	return entries, nil
}
