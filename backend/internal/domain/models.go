package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Custom types
// ---------------------------------------------------------------------------

// JSONB is a generic wrapper so any map/slice can live in a PostgreSQL jsonb column.
type JSONB json.RawMessage

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB("null")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("JSONB.Scan: type assertion to []byte failed")
	}
	*j = bytes
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB.UnmarshalJSON: on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// SkillCategory constrains the category column on skills.
type SkillCategory string

const (
	CategoryLanguage  SkillCategory = "language"
	CategoryFramework SkillCategory = "framework"
	CategoryTool      SkillCategory = "tool"
	CategoryConcept   SkillCategory = "concept"
	CategoryDatabase  SkillCategory = "database"
	CategoryDevOps    SkillCategory = "devops"
	CategoryOther     SkillCategory = "other"
)

// ProficiencyLevel constrains proficiency on user skills.
type ProficiencyLevel string

const (
	Beginner     ProficiencyLevel = "beginner"
	Intermediate ProficiencyLevel = "intermediate"
	Advanced     ProficiencyLevel = "advanced"
)

// MatchStatus constrains the status column on matches.
type MatchStatus string

const (
	MatchActive   MatchStatus = "active"
	MatchInactive MatchStatus = "inactive"
)

// RequestStatus constrains match-request status.
type RequestStatus string

const (
	RequestPending  RequestStatus = "pending"
	RequestAccepted RequestStatus = "accepted"
	RequestRejected RequestStatus = "rejected"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type User struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Email           string         `gorm:"uniqueIndex;type:varchar(255);not null" json:"email" validate:"required,email"`
	Username        string         `gorm:"uniqueIndex;type:varchar(100);not null" json:"username" validate:"required,min=3,max=100"`
	PasswordHash    string         `gorm:"type:varchar(255)" json:"-"`
	FullName        string         `gorm:"type:varchar(255)" json:"full_name"`
	Bio             string         `gorm:"type:text" json:"bio"`
	AvatarURL       string         `gorm:"type:varchar(512)" json:"avatar_url"`
	GithubURL       string         `gorm:"type:varchar(512)" json:"github_url"`
	LinkedinURL     string         `gorm:"type:varchar(512)" json:"linkedin_url"`
	GoogleID        string         `gorm:"type:varchar(255);index" json:"-"`
	GitHubID        string         `gorm:"type:varchar(255);index" json:"-"`
	ReputationScore float64        `gorm:"type:decimal(10,2);default:0" json:"reputation_score"`
	TotalSessions   int            `gorm:"default:0" json:"total_sessions"`
	Badges          JSONB          `gorm:"type:jsonb;default:'[]'" json:"badges"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Skills     []UserSkill    `gorm:"foreignKey:UserID" json:"skills,omitempty"`
	Reputation *UserReputation `gorm:"foreignKey:UserID" json:"reputation,omitempty"`
}

type Skill struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	Name        string        `gorm:"uniqueIndex;type:varchar(100);not null" json:"name" validate:"required"`
	Category    SkillCategory `gorm:"type:varchar(50);not null;index" json:"category" validate:"required"`
	Description string        `gorm:"type:text" json:"description"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
}

type UserSkill struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	UserID          string           `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_skill" json:"user_id"`
	SkillID         uint             `gorm:"not null;index;uniqueIndex:idx_user_skill" json:"skill_id"`
	ProficiencyLevel ProficiencyLevel `gorm:"type:varchar(20);not null" json:"proficiency_level" validate:"required"`
	YearsExperience float64          `gorm:"type:decimal(4,1)" json:"years_experience"`
	CredibilityScore float64         `gorm:"type:decimal(10,2);default:0" json:"credibility_score"`
	VerifiedByPeers int              `gorm:"default:0" json:"verified_by_peers"`
	CreatedAt       time.Time        `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Skill Skill `gorm:"foreignKey:SkillID;constraint:OnDelete:CASCADE" json:"skill,omitempty"`
}

type Match struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	User1ID    string      `gorm:"type:uuid;not null;index" json:"user1_id"`
	User2ID    string      `gorm:"type:uuid;not null;index" json:"user2_id"`
	MatchScore float64     `gorm:"type:decimal(5,2)" json:"match_score"`
	AIInsights JSONB       `gorm:"type:jsonb;default:'{}'" json:"ai_insights"`
	Status     MatchStatus `gorm:"type:varchar(20);default:'active';index" json:"status"`
	CreatedAt  time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time   `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User1    User      `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE" json:"user1,omitempty"`
	User2    User      `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE" json:"user2,omitempty"`
	Messages []Message `gorm:"foreignKey:MatchID" json:"messages,omitempty"`
}

type MatchRequest struct {
	ID                uint          `gorm:"primaryKey" json:"id"`
	SenderID          string        `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReceiverID        string        `gorm:"type:uuid;not null;index" json:"receiver_id"`
	Status            RequestStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	Message           string        `gorm:"type:text" json:"message"`
	AIPreviewInsights JSONB         `gorm:"type:jsonb;default:'{}'" json:"ai_preview_insights"`
	CreatedAt         time.Time     `gorm:"autoCreateTime" json:"created_at"`
	RespondedAt       *time.Time    `json:"responded_at"`

	// Relations
	Sender   User `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sender,omitempty"`
	Receiver User `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE" json:"receiver,omitempty"`
}

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   string    `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReceiverID string    `gorm:"type:uuid;not null;index" json:"receiver_id"`
	MatchID    uint      `gorm:"not null;index" json:"match_id"`
	Content    string    `gorm:"type:text;not null" json:"content" validate:"required"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Relations
	Sender   User  `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sender,omitempty"`
	Receiver User  `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE" json:"receiver,omitempty"`
	Match    Match `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"match,omitempty"`
}

type CodingSession struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	MatchID         uint       `gorm:"not null;index" json:"match_id"`
	StartedAt       time.Time  `gorm:"not null" json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	DurationMinutes int        `json:"duration_minutes"`
	CodeSnapshots   JSONB      `gorm:"type:jsonb;default:'[]'" json:"code_snapshots"`
	SessionNotes    string     `gorm:"type:text" json:"session_notes"`
	SuccessRating   float64    `gorm:"type:decimal(3,2)" json:"success_rating"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Match    Match             `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"match,omitempty"`
	Feedback []SessionFeedback `gorm:"foreignKey:SessionID" json:"feedback,omitempty"`
}

type Assessment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ChallengeID   string    `gorm:"type:varchar(100);not null;index" json:"challenge_id"`
	CodeSubmitted string    `gorm:"type:text;not null" json:"code_submitted"`
	Language      string    `gorm:"type:varchar(50);not null" json:"language"`
	AIScore       float64   `gorm:"type:decimal(5,2)" json:"ai_score"`
	SkillLevel    string    `gorm:"type:varchar(20)" json:"skill_level"`
	AIFeedback    JSONB     `gorm:"type:jsonb;default:'{}'" json:"ai_feedback"`
	CompletedAt   time.Time `json:"completed_at"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type Rating struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	RaterID             string    `gorm:"type:uuid;not null;index" json:"rater_id"`
	RatedID             string    `gorm:"type:uuid;not null;index" json:"rated_id"`
	SessionID           uint      `gorm:"not null;index" json:"session_id"`
	OverallRating       int       `gorm:"type:smallint;not null;check:overall_rating >= 1 AND overall_rating <= 5" json:"overall_rating" validate:"required,min=1,max=5"`
	CodeQualityRating   int       `gorm:"type:smallint;not null;check:code_quality_rating >= 1 AND code_quality_rating <= 5" json:"code_quality_rating" validate:"required,min=1,max=5"`
	CommunicationRating int       `gorm:"type:smallint;not null;check:communication_rating >= 1 AND communication_rating <= 5" json:"communication_rating" validate:"required,min=1,max=5"`
	HelpfulnessRating   int       `gorm:"type:smallint;not null;check:helpfulness_rating >= 1 AND helpfulness_rating <= 5" json:"helpfulness_rating" validate:"required,min=1,max=5"`
	ReliabilityRating   int       `gorm:"type:smallint;not null;check:reliability_rating >= 1 AND reliability_rating <= 5" json:"reliability_rating" validate:"required,min=1,max=5"`
	Comment             string    `gorm:"type:text" json:"comment"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Rater   User          `gorm:"foreignKey:RaterID;constraint:OnDelete:CASCADE" json:"rater,omitempty"`
	Rated   User          `gorm:"foreignKey:RatedID;constraint:OnDelete:CASCADE" json:"rated,omitempty"`
	Session CodingSession `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"session,omitempty"`
}

type SessionFeedback struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SessionID      uint      `gorm:"not null;index" json:"session_id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Enjoyed        bool      `gorm:"not null" json:"enjoyed"`
	LearnedSomething bool   `gorm:"not null" json:"learned_something"`
	WouldPairAgain bool      `gorm:"not null" json:"would_pair_again"`
	Strengths      JSONB     `gorm:"type:jsonb;default:'[]'" json:"strengths"`
	Improvements   JSONB     `gorm:"type:jsonb;default:'[]'" json:"improvements"`
	Rating         int       `gorm:"type:smallint;not null;check:rating >= 1 AND rating <= 5" json:"rating" validate:"required,min=1,max=5"`
	FeedbackText   string    `gorm:"type:text" json:"feedback_text"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Session CodingSession `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"session,omitempty"`
	User    User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type UserReputation struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	UserID                string    `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	OverallScore          float64   `gorm:"type:decimal(5,2);default:0;check:overall_score >= 0 AND overall_score <= 100" json:"overall_score"`
	CodeQualityScore      float64   `gorm:"type:decimal(5,2);default:0" json:"code_quality_score"`
	CommunicationScore    float64   `gorm:"type:decimal(5,2);default:0" json:"communication_score"`
	HelpfulnessScore      float64   `gorm:"type:decimal(5,2);default:0" json:"helpfulness_score"`
	ReliabilityScore      float64   `gorm:"type:decimal(5,2);default:0" json:"reliability_score"`
	TotalRatings          int       `gorm:"default:0" json:"total_ratings"`
	AverageRating         float64   `gorm:"type:decimal(3,2);default:0" json:"average_rating"`
	CompletedSessions     int       `gorm:"default:0" json:"completed_sessions"`
	SuccessfulMatches     int       `gorm:"default:0" json:"successful_matches"`
	SkillCredibilityScores JSONB    `gorm:"type:jsonb;default:'{}'" json:"skill_credibility_scores"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// ---------------------------------------------------------------------------
// AllModels returns every model for auto-migration.
// ---------------------------------------------------------------------------

func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Skill{},
		&UserSkill{},
		&Match{},
		&MatchRequest{},
		&Message{},
		&CodingSession{},
		&Assessment{},
		&Rating{},
		&SessionFeedback{},
		&UserReputation{},
	}
}
