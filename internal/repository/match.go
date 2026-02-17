package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/skillsync/internal/domain"
)

type MatchRepository struct {
	db *sql.DB
}

func NewMatchRepository(db *sql.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(ctx context.Context, match *domain.Match) error {
	query := `INSERT INTO matches (user_a_id, user_b_id, skill_offered, skill_wanted, status, match_score, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	          RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		match.UserAID, match.UserBID, match.SkillOffered, match.SkillWanted, match.Status, match.MatchScore,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)
}

func (r *MatchRepository) FindByID(ctx context.Context, id string) (*domain.Match, error) {
	var m domain.Match
	query := `SELECT id, user_a_id, user_b_id, skill_offered, skill_wanted, status, match_score, created_at, updated_at
	          FROM matches WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.UserAID, &m.UserBID, &m.SkillOffered, &m.SkillWanted,
		&m.Status, &m.MatchScore, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MatchRepository) ListByUser(ctx context.Context, userID string) ([]domain.Match, error) {
	query := `SELECT id, user_a_id, user_b_id, skill_offered, skill_wanted, status, match_score, created_at, updated_at
	          FROM matches WHERE user_a_id = $1 OR user_b_id = $1
	          ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		var m domain.Match
		if err := rows.Scan(&m.ID, &m.UserAID, &m.UserBID, &m.SkillOffered, &m.SkillWanted,
			&m.Status, &m.MatchScore, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (r *MatchRepository) Update(ctx context.Context, match *domain.Match) error {
	query := `UPDATE matches SET status=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, match.Status, match.ID)
	return err
}

func (r *MatchRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM matches WHERE user_a_id = $1 OR user_b_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *MatchRepository) CountCompletedByUser(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM matches WHERE (user_a_id = $1 OR user_b_id = $1) AND status = 'completed'`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *MatchRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var u domain.User
	query := `SELECT id, email, username, full_name, skills_teach, skills_learn, skill_level, reputation_score
	          FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID, &u.Email, &u.Username, &u.FullName,
		&u.SkillsTeach, &u.SkillsLearn, &u.SkillLevel, &u.ReputationScore,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
