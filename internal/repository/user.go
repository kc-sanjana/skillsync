package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/yourusername/skillsync/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, username, password_hash, full_name, skills_teach, skills_learn, skill_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Email, user.Username, user.PasswordHash, user.FullName,
		pq.Array(user.SkillsTeach), pq.Array(user.SkillsLearn), user.SkillLevel,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	var fullName, bio, avatarURL, skillLevel sql.NullString
	var reputationScore sql.NullFloat64
	var isOnline sql.NullBool
	var lastActiveAt, createdAt, updatedAt sql.NullTime

	query := `SELECT id, email, username, full_name, bio, avatar_url, skills_teach, skills_learn,
	          skill_level, reputation_score, is_online, last_active_at, created_at, updated_at
	          FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &fullName, &bio, &avatarURL,
		pq.Array(&user.SkillsTeach), pq.Array(&user.SkillsLearn), &skillLevel, &reputationScore,
		&isOnline, &lastActiveAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	user.FullName = fullName.String
	user.Bio = bio.String
	user.AvatarURL = avatarURL.String
	user.SkillLevel = skillLevel.String
	if reputationScore.Valid {
		user.ReputationScore = reputationScore.Float64
	}
	user.IsOnline = isOnline.Valid && isOnline.Bool
	if lastActiveAt.Valid {
		user.LastActiveAt = lastActiveAt.Time
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	var passwordHash, fullName, skillLevel sql.NullString
	var reputationScore sql.NullFloat64
	query := `SELECT id, email, username, password_hash, full_name, skill_level, reputation_score
	          FROM users WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &passwordHash,
		&fullName, &skillLevel, &reputationScore,
	)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = passwordHash.String
	user.FullName = fullName.String
	user.SkillLevel = skillLevel.String
	if reputationScore.Valid {
		user.ReputationScore = reputationScore.Float64
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, skill, level string) ([]domain.User, error) {
	query := `SELECT id, email, username, full_name, bio, avatar_url, skills_teach, skills_learn,
	          skill_level, reputation_score, is_online, created_at
	          FROM users WHERE 1=1`
	args := []any{}
	argIdx := 1

	if skill != "" {
		query += ` AND ($` + string(rune('0'+argIdx)) + ` = ANY(skills_teach) OR $` + string(rune('0'+argIdx)) + ` = ANY(skills_learn))`
		args = append(args, skill)
		argIdx++
	}
	if level != "" {
		query += ` AND skill_level = $` + string(rune('0'+argIdx))
		args = append(args, level)
	}

	query += ` ORDER BY reputation_score DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var fullName, bio, avatarURL, skillLevel sql.NullString
		var reputationScore sql.NullFloat64
		var isOnline sql.NullBool
		var createdAt sql.NullTime
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Username, &fullName, &bio, &avatarURL,
			pq.Array(&u.SkillsTeach), pq.Array(&u.SkillsLearn), &skillLevel, &reputationScore,
			&isOnline, &createdAt,
		); err != nil {
			return nil, err
		}
		u.FullName = fullName.String
		u.Bio = bio.String
		u.AvatarURL = avatarURL.String
		u.SkillLevel = skillLevel.String
		if reputationScore.Valid {
			u.ReputationScore = reputationScore.Float64
		}
		u.IsOnline = isOnline.Valid && isOnline.Bool
		if createdAt.Valid {
			u.CreatedAt = createdAt.Time
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET full_name=$1, bio=$2, avatar_url=$3, skills_teach=$4, skills_learn=$5, updated_at=NOW()
	          WHERE id=$6`
	_, err := r.db.ExecContext(ctx, query,
		user.FullName, user.Bio, user.AvatarURL, pq.Array(user.SkillsTeach), pq.Array(user.SkillsLearn), user.ID,
	)
	return err
}

func (r *UserRepository) UpdateSkillLevel(ctx context.Context, userID, skill, level string) error {
	query := `UPDATE users SET skill_level=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, level, userID)
	return err
}

func (r *UserRepository) UpdateReputation(ctx context.Context, userID string, score float64, badge string) error {
	query := `UPDATE users SET reputation_score=$1, updated_at=NOW() WHERE id=$2`
	_, err := r.db.ExecContext(ctx, query, score, userID)
	return err
}

func (r *UserRepository) FindByOAuth(ctx context.Context, provider, oauthID string) (*domain.User, error) {
	var user domain.User
	var lastActiveAt, createdAt, updatedAt sql.NullTime
	query := `SELECT id, email, username, COALESCE(full_name,''), COALESCE(bio,''), COALESCE(avatar_url,''),
	          skills_teach, skills_learn, COALESCE(skill_level,'beginner'), COALESCE(reputation_score,0),
	          COALESCE(is_online,false), last_active_at, created_at, updated_at
	          FROM users WHERE oauth_provider = $1 AND oauth_id = $2`

	err := r.db.QueryRowContext(ctx, query, provider, oauthID).Scan(
		&user.ID, &user.Email, &user.Username, &user.FullName, &user.Bio, &user.AvatarURL,
		pq.Array(&user.SkillsTeach), pq.Array(&user.SkillsLearn), &user.SkillLevel, &user.ReputationScore,
		&user.IsOnline, &lastActiveAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if lastActiveAt.Valid {
		user.LastActiveAt = lastActiveAt.Time
	}
	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		user.UpdatedAt = updatedAt.Time
	}
	return &user, nil
}

func (r *UserRepository) CreateOAuthUser(ctx context.Context, user *domain.User, provider, oauthID string) error {
	query := `
		INSERT INTO users (email, username, password_hash, full_name, avatar_url, skills_teach, skills_learn, skill_level, oauth_provider, oauth_id, created_at, updated_at)
		VALUES ($1, $2, '', $3, $4, $5, $6, 'beginner', $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Email, user.Username, user.FullName, user.AvatarURL,
		pq.Array(user.SkillsTeach), pq.Array(user.SkillsLearn), provider, oauthID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
