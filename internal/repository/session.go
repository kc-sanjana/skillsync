package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/skillsync/internal/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO sessions (match_id, status)
	          VALUES ($1, $2)
	          RETURNING id, started_at`
	return r.db.QueryRowContext(ctx, query, session.MatchID, session.Status).Scan(&session.ID, &session.StartedAt)
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	var s domain.Session
	query := `SELECT id, match_id, started_at, ended_at, duration_min, notes, status
	          FROM sessions WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.MatchID, &s.StartedAt, &s.EndedAt, &s.DurationMin, &s.Notes, &s.Status,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) ListByMatch(ctx context.Context, matchID string) ([]domain.Session, error) {
	query := `SELECT id, match_id, started_at, ended_at, duration_min, notes, status
	          FROM sessions WHERE match_id = $1
	          ORDER BY started_at DESC`

	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(&s.ID, &s.MatchID, &s.StartedAt, &s.EndedAt, &s.DurationMin, &s.Notes, &s.Status); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *SessionRepository) End(ctx context.Context, id string, notes string) error {
	query := `UPDATE sessions SET ended_at=NOW(), duration_min=EXTRACT(EPOCH FROM (NOW()-started_at))/60, notes=$1, status='completed'
	          WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, notes, id)
	return err
}
