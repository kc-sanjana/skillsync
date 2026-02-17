package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yourusername/skillsync/internal/domain"
)

type ClaudeService struct {
	client anthropic.Client
}

func NewClaudeService(apiKey string) *ClaudeService {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeService{client: client}
}

func (s *ClaudeService) EvaluateSkill(ctx context.Context, userID, skill string, answers []string) (*domain.Assessment, error) {
	prompt := fmt.Sprintf(
		`Evaluate skill "%s" based on answers: %v.
Respond in JSON: {"level":"beginner|intermediate|advanced","score":0-100,"feedback":"..."}`,
		skill, answers,
	)

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 500,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(prompt),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude error: %w", err)
	}

	text := resp.Content[0].Text

	var result struct {
		Level    string  `json:"level"`
		Score    float64 `json:"score"`
		Feedback string  `json:"feedback"`
	}

	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &domain.Assessment{
		UserID:   userID,
		Skill:    skill,
		Level:    result.Level,
		Score:    result.Score,
		Feedback: result.Feedback,
		Answers:  answers,
	}, nil
}

func (s *ClaudeService) GeneratePairingInsights(ctx context.Context, userA, userB *domain.User, match *domain.Match) (*domain.PairingInsight, error) {
	prompt := fmt.Sprintf(
		`Analyze compatibility between %s and %s. Respond in JSON.`,
		userA.Username, userB.Username,
	)

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 500,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(prompt),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude error: %w", err)
	}

	text := resp.Content[0].Text

	var insight domain.PairingInsight
	if err := json.Unmarshal([]byte(text), &insight); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	insight.MatchID = match.ID
	return &insight, nil
}

