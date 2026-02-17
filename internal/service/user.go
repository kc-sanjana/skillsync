package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	FullName    string
	SkillsTeach []string
	SkillsLearn []string
}

type UpdateProfileInput struct {
	FullName    string   `json:"full_name"`
	Bio         string   `json:"bio"`
	AvatarURL   string   `json:"avatar_url"`
	SkillsTeach []string `json:"skills_teach"`
	SkillsLearn []string `json:"skills_learn"`
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	existing, _ := s.repo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		SkillsTeach:  input.SkillsTeach,
		SkillsLearn:  input.SkillsLearn,
		SkillLevel:   "beginner",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, skill, level string) ([]domain.User, error) {
	return s.repo.List(ctx, skill, level)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.FullName != "" {
		user.FullName = input.FullName
	}
	if input.Bio != "" {
		user.Bio = input.Bio
	}
	if input.AvatarURL != "" {
		user.AvatarURL = input.AvatarURL
	}
	if input.SkillsTeach != nil {
		user.SkillsTeach = input.SkillsTeach
	}
	if input.SkillsLearn != nil {
		user.SkillsLearn = input.SkillsLearn
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateSkillLevel(ctx context.Context, userID, skill, level string) error {
	return s.repo.UpdateSkillLevel(ctx, userID, skill, level)
}

func (s *UserService) FindOrCreateOAuthUser(ctx context.Context, provider, oauthID, email, name, avatarURL string) (*domain.User, error) {
	// Check if OAuth user already exists
	user, err := s.repo.FindByOAuth(ctx, provider, oauthID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}

	// Generate a unique username from the name
	username := strings.ToLower(strings.ReplaceAll(name, " ", "")) + "_" + provider

	newUser := &domain.User{
		Email:       email,
		Username:    username,
		FullName:    name,
		AvatarURL:   avatarURL,
		SkillsTeach: []string{},
		SkillsLearn: []string{},
	}

	if err := s.repo.CreateOAuthUser(ctx, newUser, provider, oauthID); err != nil {
		return nil, err
	}

	return newUser, nil
}
