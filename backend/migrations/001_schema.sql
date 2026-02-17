-- ============================================================================
-- SkillSync schema – PostgreSQL
-- Run: psql -U postgres -d skillsync -f migrations/001_schema.sql
-- ============================================================================

BEGIN;

-- ---------- extensions ----------
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ==========================================================================
-- 1. users
-- ==========================================================================
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL       PRIMARY KEY,
    email           VARCHAR(255)    NOT NULL,
    username        VARCHAR(100)    NOT NULL,
    password_hash   VARCHAR(255)    NOT NULL,
    full_name       VARCHAR(255)    DEFAULT '',
    bio             TEXT            DEFAULT '',
    avatar_url      VARCHAR(512)    DEFAULT '',
    github_url      VARCHAR(512)    DEFAULT '',
    linkedin_url    VARCHAR(512)    DEFAULT '',
    reputation_score DECIMAL(10,2)  DEFAULT 0,
    total_sessions  INTEGER         DEFAULT 0,
    badges          JSONB           DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email    ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at      ON users (deleted_at);

-- ==========================================================================
-- 2. skills
-- ==========================================================================
CREATE TABLE IF NOT EXISTS skills (
    id          BIGSERIAL       PRIMARY KEY,
    name        VARCHAR(100)    NOT NULL,
    category    VARCHAR(50)     NOT NULL,
    description TEXT            DEFAULT '',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_skills_name     ON skills (name);
CREATE INDEX IF NOT EXISTS idx_skills_category        ON skills (category);

-- ==========================================================================
-- 3. user_skills
-- ==========================================================================
CREATE TABLE IF NOT EXISTS user_skills (
    id                BIGSERIAL       PRIMARY KEY,
    user_id           BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    skill_id          BIGINT          NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
    proficiency_level VARCHAR(20)     NOT NULL
                      CHECK (proficiency_level IN ('beginner','intermediate','advanced')),
    years_experience  DECIMAL(4,1)    DEFAULT 0,
    credibility_score DECIMAL(10,2)   DEFAULT 0,
    verified_by_peers INTEGER         DEFAULT 0,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_skill ON user_skills (user_id, skill_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_user  ON user_skills (user_id);
CREATE INDEX IF NOT EXISTS idx_user_skills_skill ON user_skills (skill_id);

-- ==========================================================================
-- 4. matches
-- ==========================================================================
CREATE TABLE IF NOT EXISTS matches (
    id          BIGSERIAL       PRIMARY KEY,
    user1_id    BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    user2_id    BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    match_score DECIMAL(5,2)    DEFAULT 0,
    ai_insights JSONB           DEFAULT '{}'::jsonb,
    status      VARCHAR(20)     DEFAULT 'active'
                CHECK (status IN ('active','inactive')),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_matches_user1  ON matches (user1_id);
CREATE INDEX IF NOT EXISTS idx_matches_user2  ON matches (user2_id);
CREATE INDEX IF NOT EXISTS idx_matches_status ON matches (status);

-- ==========================================================================
-- 5. match_requests
-- ==========================================================================
CREATE TABLE IF NOT EXISTS match_requests (
    id                  BIGSERIAL       PRIMARY KEY,
    sender_id           BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    receiver_id         BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status              VARCHAR(20)     DEFAULT 'pending'
                        CHECK (status IN ('pending','accepted','rejected')),
    message             TEXT            DEFAULT '',
    ai_preview_insights JSONB           DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    responded_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_match_requests_sender   ON match_requests (sender_id);
CREATE INDEX IF NOT EXISTS idx_match_requests_receiver ON match_requests (receiver_id);
CREATE INDEX IF NOT EXISTS idx_match_requests_status   ON match_requests (status);

-- ==========================================================================
-- 6. messages
-- ==========================================================================
CREATE TABLE IF NOT EXISTS messages (
    id          BIGSERIAL       PRIMARY KEY,
    sender_id   BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    receiver_id BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    match_id    BIGINT          NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    content     TEXT            NOT NULL,
    is_read     BOOLEAN         DEFAULT FALSE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_sender   ON messages (sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages (receiver_id);
CREATE INDEX IF NOT EXISTS idx_messages_match    ON messages (match_id);
CREATE INDEX IF NOT EXISTS idx_messages_created  ON messages (created_at);

-- ==========================================================================
-- 7. coding_sessions
-- ==========================================================================
CREATE TABLE IF NOT EXISTS coding_sessions (
    id               BIGSERIAL       PRIMARY KEY,
    match_id         BIGINT          NOT NULL REFERENCES matches (id) ON DELETE CASCADE,
    started_at       TIMESTAMPTZ     NOT NULL,
    ended_at         TIMESTAMPTZ,
    duration_minutes INTEGER         DEFAULT 0,
    code_snapshots   JSONB           DEFAULT '[]'::jsonb,
    session_notes    TEXT            DEFAULT '',
    success_rating   DECIMAL(3,2)    DEFAULT 0
                     CHECK (success_rating >= 0 AND success_rating <= 1),
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coding_sessions_match ON coding_sessions (match_id);

-- ==========================================================================
-- 8. assessments
-- ==========================================================================
CREATE TABLE IF NOT EXISTS assessments (
    id             BIGSERIAL       PRIMARY KEY,
    user_id        BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    challenge_id   VARCHAR(100)    NOT NULL,
    code_submitted TEXT            NOT NULL,
    language       VARCHAR(50)     NOT NULL,
    ai_score       DECIMAL(5,2)    DEFAULT 0
                   CHECK (ai_score >= 0 AND ai_score <= 100),
    skill_level    VARCHAR(20)     DEFAULT '',
    ai_feedback    JSONB           DEFAULT '{}'::jsonb,
    completed_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assessments_user      ON assessments (user_id);
CREATE INDEX IF NOT EXISTS idx_assessments_challenge ON assessments (challenge_id);

-- ==========================================================================
-- 9. ratings
-- ==========================================================================
CREATE TABLE IF NOT EXISTS ratings (
    id                   BIGSERIAL   PRIMARY KEY,
    rater_id             BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rated_id             BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    session_id           BIGINT      NOT NULL REFERENCES coding_sessions (id) ON DELETE CASCADE,
    overall_rating       SMALLINT    NOT NULL CHECK (overall_rating       >= 1 AND overall_rating       <= 5),
    code_quality_rating  SMALLINT    NOT NULL CHECK (code_quality_rating  >= 1 AND code_quality_rating  <= 5),
    communication_rating SMALLINT    NOT NULL CHECK (communication_rating >= 1 AND communication_rating <= 5),
    helpfulness_rating   SMALLINT    NOT NULL CHECK (helpfulness_rating   >= 1 AND helpfulness_rating   <= 5),
    reliability_rating   SMALLINT    NOT NULL CHECK (reliability_rating   >= 1 AND reliability_rating   <= 5),
    comment              TEXT        DEFAULT '',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ratings_rater   ON ratings (rater_id);
CREATE INDEX IF NOT EXISTS idx_ratings_rated   ON ratings (rated_id);
CREATE INDEX IF NOT EXISTS idx_ratings_session ON ratings (session_id);

-- ==========================================================================
-- 10. session_feedbacks
-- ==========================================================================
CREATE TABLE IF NOT EXISTS session_feedbacks (
    id                BIGSERIAL       PRIMARY KEY,
    session_id        BIGINT          NOT NULL REFERENCES coding_sessions (id) ON DELETE CASCADE,
    user_id           BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    enjoyed           BOOLEAN         NOT NULL DEFAULT FALSE,
    learned_something BOOLEAN         NOT NULL DEFAULT FALSE,
    would_pair_again  BOOLEAN         NOT NULL DEFAULT FALSE,
    strengths         JSONB           DEFAULT '[]'::jsonb,
    improvements      JSONB           DEFAULT '[]'::jsonb,
    rating            SMALLINT        NOT NULL CHECK (rating >= 1 AND rating <= 5),
    feedback_text     TEXT            DEFAULT '',
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_feedbacks_session ON session_feedbacks (session_id);
CREATE INDEX IF NOT EXISTS idx_session_feedbacks_user    ON session_feedbacks (user_id);

-- ==========================================================================
-- 11. user_reputations
-- ==========================================================================
CREATE TABLE IF NOT EXISTS user_reputations (
    id                       BIGSERIAL       PRIMARY KEY,
    user_id                  BIGINT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    overall_score            DECIMAL(5,2)    DEFAULT 0
                             CHECK (overall_score >= 0 AND overall_score <= 100),
    code_quality_score       DECIMAL(5,2)    DEFAULT 0,
    communication_score      DECIMAL(5,2)    DEFAULT 0,
    helpfulness_score        DECIMAL(5,2)    DEFAULT 0,
    reliability_score        DECIMAL(5,2)    DEFAULT 0,
    total_ratings            INTEGER         DEFAULT 0,
    average_rating           DECIMAL(3,2)    DEFAULT 0,
    completed_sessions       INTEGER         DEFAULT 0,
    successful_matches       INTEGER         DEFAULT 0,
    skill_credibility_scores JSONB           DEFAULT '{}'::jsonb,
    updated_at               TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_reputations_user ON user_reputations (user_id);

-- ==========================================================================
-- Seed: 25 skills
-- ==========================================================================
INSERT INTO skills (name, category, description) VALUES
    ('JavaScript',  'language',  'Dynamic scripting language for web development'),
    ('TypeScript',  'language',  'Typed superset of JavaScript'),
    ('Python',      'language',  'General-purpose language popular in data science and backends'),
    ('Go',          'language',  'Statically typed language built for concurrency'),
    ('Rust',        'language',  'Systems language focused on safety and performance'),
    ('Java',        'language',  'Enterprise-grade object-oriented language'),
    ('C++',         'language',  'High-performance systems programming language'),
    ('Ruby',        'language',  'Dynamic language optimized for developer happiness'),
    ('React',       'framework', 'Component-based UI library for JavaScript'),
    ('Vue.js',      'framework', 'Progressive JavaScript framework for UIs'),
    ('Angular',     'framework', 'Full-featured frontend framework by Google'),
    ('Next.js',     'framework', 'React meta-framework with SSR and routing'),
    ('Django',      'framework', 'Batteries-included Python web framework'),
    ('Express.js',  'framework', 'Minimal Node.js web framework'),
    ('Spring Boot', 'framework', 'Java framework for production-grade applications'),
    ('PostgreSQL',  'database',  'Advanced open-source relational database'),
    ('MongoDB',     'database',  'Document-oriented NoSQL database'),
    ('Redis',       'database',  'In-memory data structure store'),
    ('Docker',      'devops',    'Container platform for packaging applications'),
    ('Kubernetes',  'devops',    'Container orchestration system'),
    ('AWS',         'devops',    'Amazon Web Services cloud platform'),
    ('Git',         'tool',      'Distributed version control system'),
    ('GraphQL',     'tool',      'Query language for APIs'),
    ('REST API',    'concept',   'Architectural style for networked applications'),
    ('System Design','concept',  'Designing large-scale distributed systems')
ON CONFLICT (name) DO NOTHING;

COMMIT;
