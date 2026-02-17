package domain

import "time"

// User represents a registered user with their skills.
type User struct {
	ID             string    `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	Username       string    `json:"username" db:"username"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	FullName       string    `json:"full_name" db:"full_name"`
	Bio            string    `json:"bio" db:"bio"`
	AvatarURL      string    `json:"avatar_url" db:"avatar_url"`
	SkillsTeach    []string  `json:"skills_teach" db:"skills_teach"`
	SkillsLearn    []string  `json:"skills_learn" db:"skills_learn"`
	SkillLevel     string    `json:"skill_level" db:"skill_level"` // beginner, intermediate, advanced
	ReputationScore float64  `json:"reputation_score" db:"reputation_score"`
	IsOnline       bool      `json:"is_online" db:"is_online"`
	LastActiveAt   time.Time `json:"last_active_at" db:"last_active_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Match represents a skill-exchange pairing between two users.
type Match struct {
	ID          string    `json:"id" db:"id"`
	UserAID     string    `json:"user_a_id" db:"user_a_id"`
	UserBID     string    `json:"user_b_id" db:"user_b_id"`
	SkillOffered string   `json:"skill_offered" db:"skill_offered"`
	SkillWanted  string   `json:"skill_wanted" db:"skill_wanted"`
	Status      string    `json:"status" db:"status"` // pending, accepted, rejected, completed
	MatchScore  float64   `json:"match_score" db:"match_score"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Message represents a chat message within a match.
type Message struct {
	ID        string    `json:"id" db:"id"`
	MatchID   string    `json:"match_id" db:"match_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	Type      string    `json:"type" db:"type"` // text, code, file
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Rating represents feedback one user gives another after a session.
type Rating struct {
	ID             string    `json:"id" db:"id"`
	MatchID        string    `json:"match_id" db:"match_id"`
	RaterID        string    `json:"rater_id" db:"rater_id"`
	RatedUserID    string    `json:"rated_user_id" db:"rated_user_id"`
	Score          int       `json:"score" db:"score"` // 1-5
	Communication  int       `json:"communication" db:"communication"`
	Knowledge      int       `json:"knowledge" db:"knowledge"`
	Helpfulness    int       `json:"helpfulness" db:"helpfulness"`
	Comment        string    `json:"comment" db:"comment"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Reputation aggregates a user's rating history.
type Reputation struct {
	UserID            string  `json:"user_id" db:"user_id"`
	OverallScore      float64 `json:"overall_score" db:"overall_score"`
	TotalRatings      int     `json:"total_ratings" db:"total_ratings"`
	TotalSessions     int     `json:"total_sessions" db:"total_sessions"`
	AvgCommunication  float64 `json:"avg_communication" db:"avg_communication"`
	AvgKnowledge      float64 `json:"avg_knowledge" db:"avg_knowledge"`
	AvgHelpfulness    float64 `json:"avg_helpfulness" db:"avg_helpfulness"`
	Rank              int     `json:"rank" db:"rank"`
	Badge             string  `json:"badge" db:"badge"` // newcomer, rising_star, expert, mentor
}

// Session tracks a live skill-exchange session.
type Session struct {
	ID          string    `json:"id" db:"id"`
	MatchID     string    `json:"match_id" db:"match_id"`
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	EndedAt     *time.Time `json:"ended_at" db:"ended_at"`
	DurationMin int       `json:"duration_min" db:"duration_min"`
	Notes       string    `json:"notes" db:"notes"`
	Status      string    `json:"status" db:"status"` // active, completed, cancelled
}

// Assessment holds Claude's evaluation of a user's skill.
type Assessment struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Skill      string    `json:"skill" db:"skill"`
	Level      string    `json:"level" db:"level"` // beginner, intermediate, advanced
	Score      float64   `json:"score" db:"score"`
	Feedback   string    `json:"feedback" db:"feedback"`
	Questions  []string  `json:"questions" db:"questions"`
	Answers    []string  `json:"answers" db:"answers"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// PairingInsight contains Claude-generated analysis of a match.
type PairingInsight struct {
	MatchID          string   `json:"match_id"`
	CompatibilityScore float64 `json:"compatibility_score"`
	Strengths        []string `json:"strengths"`
	Challenges       []string `json:"challenges"`
	SuggestedTopics  []string `json:"suggested_topics"`
	LearningPlan     string   `json:"learning_plan"`
}

// LeaderboardEntry is a row in the reputation leaderboard.
type LeaderboardEntry struct {
	Rank           int     `json:"rank"`
	UserID         string  `json:"user_id"`
	Username       string  `json:"username"`
	AvatarURL      string  `json:"avatar_url"`
	OverallScore   float64 `json:"overall_score"`
	TotalSessions  int     `json:"total_sessions"`
	Badge          string  `json:"badge"`
}
