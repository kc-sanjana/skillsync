package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/rs/zerolog/log"

	"github.com/yourusername/skillsync/internal/domain"
)

// ---------------------------------------------------------------------------
// Result structs
// ---------------------------------------------------------------------------

type CodeAnalysisResult struct {
	Score          int      `json:"score"`
	SkillLevel     string   `json:"skill_level"`
	Strengths      []string `json:"strengths"`
	Improvements   []string `json:"improvements"`
	CodeQuality    string   `json:"code_quality"`
	Readability    int      `json:"readability"`
	Efficiency     int      `json:"efficiency"`
	ErrorHandling  bool     `json:"error_handling"`
	Recommendation string   `json:"recommendation"`
}

type ProjectSuggestion struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	SkillsUsed       []string `json:"skills_used"`
	Difficulty       string   `json:"difficulty"`
	EstimatedHours   int      `json:"estimated_hours"`
	LearningOutcomes []string `json:"learning_outcomes"`
}

type PairingInsights struct {
	OverallReasoning      string   `json:"overall_reasoning"`
	SkillComplement       string   `json:"skill_complement"`
	LearningOpportunities []string `json:"learning_opportunities"`
	CollaborationIdeas    []string `json:"collaboration_ideas"`
	Recommendation        string   `json:"recommendation"`
}

type SuccessPrediction struct {
	SuccessProbability float64  `json:"success_probability"`
	Confidence         string   `json:"confidence"`
	SuccessFactors     []string `json:"success_factors"`
	Challenges         []string `json:"challenges"`
	Tips               []string `json:"tips"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ClaudeService struct {
	client *anthropic.Client
}

func NewClaudeService() *ClaudeService {
	client := anthropic.NewClient() // reads ANTHROPIC_API_KEY from env
	return &ClaudeService{client: &client}
}

// ---------------------------------------------------------------------------
// AnalyzeCode
// ---------------------------------------------------------------------------

func (s *ClaudeService) AnalyzeCode(code, language string) (*CodeAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following %s code and return a JSON object with exactly these fields:
{
  "score": <int 0-100>,
  "skill_level": "<beginner|intermediate|advanced>",
  "strengths": ["<strength1>", "<strength2>", ...],
  "improvements": ["<improvement1>", "<improvement2>", ...],
  "code_quality": "<brief assessment>",
  "readability": <int 1-10>,
  "efficiency": <int 1-10>,
  "error_handling": <bool whether code handles errors properly>,
  "recommendation": "<one paragraph recommendation>"
}

Return ONLY the JSON object, no other text.

Code:
%s`, language, code)

	raw, err := s.call(anthropic.ModelClaudeSonnet4_5, prompt, "You are an expert code reviewer. Respond only with valid JSON.", 1024)
	if err != nil {
		return nil, fmt.Errorf("AnalyzeCode: %w", err)
	}

	var result CodeAnalysisResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("AnalyzeCode: failed to parse response: %w", err)
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// GenerateHint
// ---------------------------------------------------------------------------

func (s *ClaudeService) GenerateHint(code, language, problem string) (string, error) {
	prompt := fmt.Sprintf(`A developer is working on the following problem in %s:

Problem: %s

Their current code:
%s

Give a helpful hint that guides them toward the solution WITHOUT giving the answer directly.
Be encouraging and educational. Keep your hint to 2-3 sentences.`, language, problem, code)

	hint, err := s.call(anthropic.ModelClaudeHaiku4_5, prompt, "You are a supportive coding mentor. Give hints, never full solutions.", 256)
	if err != nil {
		return "", fmt.Errorf("GenerateHint: %w", err)
	}
	return hint, nil
}

// ---------------------------------------------------------------------------
// CalculateMatchScore
// ---------------------------------------------------------------------------

func (s *ClaudeService) CalculateMatchScore(user1Skills, user2Skills []string, user1Goals, user2Goals string) (float64, string, error) {
	prompt := fmt.Sprintf(`Given two developers, calculate how well they would pair for collaborative learning.

User 1 skills: %s
User 1 goals: %s

User 2 skills: %s
User 2 goals: %s

Return ONLY a JSON object:
{
  "score": <float 0-100>,
  "reasoning": "<2-3 sentence explanation>"
}

Consider: skill complementarity, goal alignment, potential for mutual teaching, and learning synergy.
A high score means they can teach each other effectively.`,
		strings.Join(user1Skills, ", "), user1Goals,
		strings.Join(user2Skills, ", "), user2Goals)

	raw, err := s.call(anthropic.ModelClaudeHaiku4_5, prompt, "You are a matching algorithm expert. Respond only with valid JSON.", 256)
	if err != nil {
		return 0, "", fmt.Errorf("CalculateMatchScore: %w", err)
	}

	var result struct {
		Score     float64 `json:"score"`
		Reasoning string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return 0, "", fmt.Errorf("CalculateMatchScore: failed to parse response: %w", err)
	}
	return result.Score, result.Reasoning, nil
}

// ---------------------------------------------------------------------------
// SuggestProjects
// ---------------------------------------------------------------------------

func (s *ClaudeService) SuggestProjects(skills []string, skillLevel string) ([]*ProjectSuggestion, error) {
	prompt := fmt.Sprintf(`Suggest exactly 3 collaborative coding projects for a developer with these skills: %s
Skill level: %s

Return ONLY a JSON array with exactly 3 objects, each having:
{
  "title": "<project title>",
  "description": "<2-3 sentence description>",
  "skills_used": ["<skill1>", "<skill2>", ...],
  "difficulty": "<beginner|intermediate|advanced>",
  "estimated_hours": <int>,
  "learning_outcomes": ["<outcome1>", "<outcome2>", ...]
}

Projects should be practical, interesting, and appropriate for the skill level.`,
		strings.Join(skills, ", "), skillLevel)

	raw, err := s.call(anthropic.ModelClaudeSonnet4_5, prompt, "You are a senior developer who suggests engaging projects. Respond only with valid JSON.", 1024)
	if err != nil {
		return nil, fmt.Errorf("SuggestProjects: %w", err)
	}

	var results []*ProjectSuggestion
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, fmt.Errorf("SuggestProjects: failed to parse response: %w", err)
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// GeneratePairingInsights
// ---------------------------------------------------------------------------

func (s *ClaudeService) GeneratePairingInsights(
	user1, user2 domain.User,
	user1Skills, user2Skills []domain.UserSkill,
) (*PairingInsights, error) {
	u1s := formatSkills(user1Skills)
	u2s := formatSkills(user2Skills)

	prompt := fmt.Sprintf(`Analyze the potential pairing between two developers:

Developer 1: %s
  Skills: %s
  Reputation score: %.1f
  Total sessions: %d

Developer 2: %s
  Skills: %s
  Reputation score: %.1f
  Total sessions: %d

Return ONLY a JSON object:
{
  "overall_reasoning": "<paragraph explaining the match>",
  "skill_complement": "<how their skills complement each other>",
  "learning_opportunities": ["<opportunity1>", "<opportunity2>", ...],
  "collaboration_ideas": ["<idea1>", "<idea2>", ...],
  "recommendation": "<pair / consider / skip>"
}`,
		user1.FullName, u1s, user1.ReputationScore, user1.TotalSessions,
		user2.FullName, u2s, user2.ReputationScore, user2.TotalSessions)

	raw, err := s.call(anthropic.ModelClaudeSonnet4_5, prompt, "You are an expert at building effective developer teams. Respond only with valid JSON.", 1024)
	if err != nil {
		return nil, fmt.Errorf("GeneratePairingInsights: %w", err)
	}

	var result PairingInsights
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("GeneratePairingInsights: failed to parse response: %w", err)
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// PredictSessionSuccess
// ---------------------------------------------------------------------------

func (s *ClaudeService) PredictSessionSuccess(user1Rep, user2Rep domain.UserReputation) (*SuccessPrediction, error) {
	prompt := fmt.Sprintf(`Predict the success of a pair-programming session between two developers based on their reputation data.

Developer 1 reputation:
  Overall score: %.1f/100
  Code quality: %.1f, Communication: %.1f, Helpfulness: %.1f, Reliability: %.1f
  Average rating: %.1f/5
  Completed sessions: %d, Successful matches: %d

Developer 2 reputation:
  Overall score: %.1f/100
  Code quality: %.1f, Communication: %.1f, Helpfulness: %.1f, Reliability: %.1f
  Average rating: %.1f/5
  Completed sessions: %d, Successful matches: %d

Return ONLY a JSON object:
{
  "success_probability": <float 0-100>,
  "confidence": "<low|medium|high>",
  "success_factors": ["<factor1>", "<factor2>", ...],
  "challenges": ["<challenge1>", "<challenge2>", ...],
  "tips": ["<tip1>", "<tip2>", ...]
}`,
		user1Rep.OverallScore,
		user1Rep.CodeQualityScore, user1Rep.CommunicationScore,
		user1Rep.HelpfulnessScore, user1Rep.ReliabilityScore,
		user1Rep.AverageRating, user1Rep.CompletedSessions, user1Rep.SuccessfulMatches,
		user2Rep.OverallScore,
		user2Rep.CodeQualityScore, user2Rep.CommunicationScore,
		user2Rep.HelpfulnessScore, user2Rep.ReliabilityScore,
		user2Rep.AverageRating, user2Rep.CompletedSessions, user2Rep.SuccessfulMatches)

	raw, err := s.call(anthropic.ModelClaudeHaiku4_5, prompt, "You are a data-driven session-success predictor. Respond only with valid JSON.", 512)
	if err != nil {
		return nil, fmt.Errorf("PredictSessionSuccess: %w", err)
	}

	var result SuccessPrediction
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("PredictSessionSuccess: failed to parse response: %w", err)
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// call makes a single Messages API request and returns the text content.
func (s *ClaudeService) call(model anthropic.Model, userPrompt, systemPrompt string, maxTokens int64) (string, error) {
	resp, err := s.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
		Temperature: anthropic.Float(0.3),
	})
	if err != nil {
		log.Error().Err(err).Str("model", string(model)).Msg("claude api call failed")
		return "", fmt.Errorf("claude api error: %w", err)
	}

	return extractText(resp), nil
}

// extractText pulls the concatenated text out of a Message response.
func extractText(msg *anthropic.Message) string {
	var b strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			b.WriteString(block.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

// formatSkills turns a slice of UserSkill into a readable string.
func formatSkills(skills []domain.UserSkill) string {
	if len(skills) == 0 {
		return "none listed"
	}
	parts := make([]string, len(skills))
	for i, s := range skills {
		parts[i] = fmt.Sprintf("%s (%s, %.1f yrs)",
			s.Skill.Name, s.ProficiencyLevel, s.YearsExperience)
	}
	return strings.Join(parts, ", ")
}
