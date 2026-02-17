package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/yourusername/skillsync/internal/domain"
)

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
	ErrUserNotFound  = errors.New("user not found")
	ErrSkillNotFound = errors.New("skill not found")
	ErrSkillExists   = errors.New("user already has this skill")
	ErrInvalidLevel  = errors.New("invalid proficiency level; use beginner, intermediate, or advanced")
)

// UserWithReputation bundles a user with their reputation data for API
// responses that need both.
type UserWithReputation struct {
	domain.User
	Reputation *domain.UserReputation `json:"reputation"`
}

// UserService handles all user-related business logic.
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a UserService backed by the given database handle.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// ---------------------------------------------------------------------------
// CreateUser
// ---------------------------------------------------------------------------

func (s *UserService) CreateUser(email, username, password, fullName string) (*domain.User, error) {
	// Check email uniqueness.
	var count int64
	s.db.Model(&domain.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return nil, ErrEmailTaken
	}

	// Check username uniqueness.
	s.db.Model(&domain.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		FullName:     fullName,
		Badges:       domain.JSONB("[]"),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Bootstrap a reputation row for the new user.
	rep := domain.UserReputation{
		UserID:                 user.ID,
		SkillCredibilityScores: domain.JSONB("{}"),
	}
	s.db.Create(&rep)

	return &user, nil
}

// ---------------------------------------------------------------------------
// GetUser
// ---------------------------------------------------------------------------

func (s *UserService) GetUser(id string) (*domain.User, error) {
	var user domain.User
	err := s.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return &user, nil
}

// ---------------------------------------------------------------------------
// UpdateProfile
// ---------------------------------------------------------------------------

func (s *UserService) UpdateProfile(id string, updates map[string]interface{}) error {
	// Whitelist the columns that callers are allowed to touch.
	allowed := map[string]bool{
		"full_name":   true,
		"bio":         true,
		"avatar_url":  true,
		"github_url":  true,
		"linkedin_url": true,
	}

	clean := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		return nil
	}

	res := s.db.Model(&domain.User{}).Where("id = ?", id).Updates(clean)
	if res.Error != nil {
		return fmt.Errorf("failed to update profile: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// SearchUsers
// ---------------------------------------------------------------------------

func (s *UserService) SearchUsers(skills []string, skillLevel string, limit, offset int) ([]*domain.User, int64, error) {
	query := s.db.Model(&domain.User{})

	if len(skills) > 0 {
		query = query.Where(
			"id IN (?)",
			s.db.Model(&domain.UserSkill{}).
				Select("user_id").
				Joins("JOIN skills ON skills.id = user_skills.skill_id").
				Where("skills.name IN ?", skills),
		)
	}

	if skillLevel != "" {
		query = query.Where(
			"id IN (?)",
			s.db.Model(&domain.UserSkill{}).
				Select("user_id").
				Where("proficiency_level = ?", skillLevel),
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	var users []*domain.User
	err := query.
		Preload("Skills.Skill").
		Order("reputation_score DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	return users, total, nil
}

// ---------------------------------------------------------------------------
// AddSkill
// ---------------------------------------------------------------------------

func (s *UserService) AddSkill(userID string, skillName, proficiency string, years float64) error {
	level := domain.ProficiencyLevel(proficiency)
	switch level {
	case domain.Beginner, domain.Intermediate, domain.Advanced:
	default:
		return ErrInvalidLevel
	}

	// Find or create the skill.
	var skill domain.Skill
	err := s.db.Where("name = ?", skillName).First(&skill).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		skill = domain.Skill{
			Name:     skillName,
			Category: domain.CategoryOther,
		}
		if err := s.db.Create(&skill).Error; err != nil {
			return fmt.Errorf("failed to create skill: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to look up skill: %w", err)
	}

	// Guard against duplicates.
	var exists int64
	s.db.Model(&domain.UserSkill{}).
		Where("user_id = ? AND skill_id = ?", userID, skill.ID).
		Count(&exists)
	if exists > 0 {
		return ErrSkillExists
	}

	us := domain.UserSkill{
		UserID:           userID,
		SkillID:          skill.ID,
		ProficiencyLevel: level,
		YearsExperience:  years,
	}
	if err := s.db.Create(&us).Error; err != nil {
		return fmt.Errorf("failed to add skill: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GetUserWithReputation
// ---------------------------------------------------------------------------

func (s *UserService) GetUserWithReputation(id string) (*UserWithReputation, error) {
	var user domain.User
	err := s.db.First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	result := &UserWithReputation{
		User: user,
	}

	// Reputation table may not exist yet; ignore errors.
	var rep domain.UserReputation
	if err := s.db.Where("user_id = ?", id).First(&rep).Error; err == nil {
		result.Reputation = &rep
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// FindOrCreateOAuthUser
// ---------------------------------------------------------------------------

func (s *UserService) FindOrCreateOAuthUser(provider, providerID, email, fullName, avatarURL string) (*domain.User, error) {
	var user domain.User

	// 1. Look up by provider ID.
	providerCol := provider + "_id" // "google_id" or "github_id"
	err := s.db.Where(providerCol+" = ?", providerID).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// 2. Look up by email to link existing account.
	if email != "" {
		err = s.db.Where("email = ?", email).First(&user).Error
		if err == nil {
			// Link the provider ID to the existing account.
			s.db.Model(&user).Update(providerCol, providerID)
			if avatarURL != "" && user.AvatarURL == "" {
				s.db.Model(&user).Update("avatar_url", avatarURL)
			}
			return &user, nil
		}
	}

	// 3. Create new user.
	username := s.generateUniqueUsername(fullName, provider)

	user = domain.User{
		Email:     email,
		Username:  username,
		FullName:  fullName,
		AvatarURL: avatarURL,
		Badges:    domain.JSONB("[]"),
	}

	// Set the provider ID.
	switch provider {
	case "google":
		user.GoogleID = providerID
	case "github":
		user.GitHubID = providerID
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create oauth user: %w", err)
	}

	// Bootstrap reputation row.
	rep := domain.UserReputation{
		UserID:                 user.ID,
		SkillCredibilityScores: domain.JSONB("{}"),
	}
	s.db.Create(&rep)

	return &user, nil
}

// generateUniqueUsername creates a username from the user's name, appending a
// number if the base name is already taken.
func (s *UserService) generateUniqueUsername(fullName, provider string) string {
	base := strings.ToLower(strings.ReplaceAll(fullName, " ", ""))
	if base == "" {
		base = provider + "user"
	}
	// Keep only alphanumeric chars.
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, base)
	if len(clean) < 3 {
		clean = provider + "user"
	}

	username := clean
	var count int64
	for i := 1; ; i++ {
		s.db.Model(&domain.User{}).Where("username = ?", username).Count(&count)
		if count == 0 {
			return username
		}
		username = fmt.Sprintf("%s%d", clean, i)
	}
}

// ---------------------------------------------------------------------------
// Authenticate (used by login handler)
// ---------------------------------------------------------------------------

func (s *UserService) Authenticate(email, password string) (*domain.User, error) {
	var user domain.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}
