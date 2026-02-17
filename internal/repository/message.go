package repository

import (
	"context"
	"database/sql"

	"github.com/yourusername/skillsync/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.Message) error {
	query := `INSERT INTO messages (match_id, sender_id, content, type)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		msg.MatchID, msg.SenderID, msg.Content, msg.Type,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *MessageRepository) ListByMatch(ctx context.Context, matchID string, limit, offset int) ([]domain.Message, error) {
	query := `SELECT id, match_id, sender_id, content, type, created_at
	          FROM messages WHERE match_id = $1
	          ORDER BY created_at ASC
	          LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, matchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.MatchID, &m.SenderID, &m.Content, &m.Type, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
