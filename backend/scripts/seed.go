// Usage:
//   go run scripts/seed.go                  # seed default data
//   go run scripts/seed.go --reset          # drop + recreate tables, then seed
//   go run scripts/seed.go --users=30       # create 30 users instead of 15
//
// Requires the same DB env vars as the main API (DB_HOST, DB_USER, …).
// Reads .env from the project root automatically.

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"os"

	"github.com/yourusername/skillsync/internal/domain"
	"github.com/yourusername/skillsync/pkg/database"
)

func main() {
	reset := flag.Bool("reset", false, "drop all tables before seeding")
	userCount := flag.Int("users", 15, "number of sample users to create")
	flag.Parse()

	godotenv.Load()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	db, err := database.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	defer database.Close()

	if *reset {
		log.Warn().Msg("dropping all tables")
		db.Exec("DROP TABLE IF EXISTS session_feedbacks CASCADE")
		db.Exec("DROP TABLE IF EXISTS ratings CASCADE")
		db.Exec("DROP TABLE IF EXISTS assessments CASCADE")
		db.Exec("DROP TABLE IF EXISTS coding_sessions CASCADE")
		db.Exec("DROP TABLE IF EXISTS messages CASCADE")
		db.Exec("DROP TABLE IF EXISTS match_requests CASCADE")
		db.Exec("DROP TABLE IF EXISTS matches CASCADE")
		db.Exec("DROP TABLE IF EXISTS user_skills CASCADE")
		db.Exec("DROP TABLE IF EXISTS user_reputations CASCADE")
		db.Exec("DROP TABLE IF EXISTS skills CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		log.Info().Msg("tables dropped")
	}

	if err := database.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	// ----- password hash (shared) -----
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	passwordHash := string(hash)

	// ===================================================================
	// 1. Skills (25)
	// ===================================================================
	skills := []domain.Skill{
		{Name: "JavaScript", Category: "language", Description: "Dynamic scripting language for web development"},
		{Name: "TypeScript", Category: "language", Description: "Typed superset of JavaScript"},
		{Name: "Python", Category: "language", Description: "General-purpose language popular in data science and backends"},
		{Name: "Go", Category: "language", Description: "Statically typed language built for concurrency"},
		{Name: "Rust", Category: "language", Description: "Systems language focused on safety and performance"},
		{Name: "Java", Category: "language", Description: "Enterprise-grade object-oriented language"},
		{Name: "C++", Category: "language", Description: "High-performance systems programming language"},
		{Name: "Ruby", Category: "language", Description: "Dynamic language optimized for developer happiness"},
		{Name: "React", Category: "framework", Description: "Component-based UI library for JavaScript"},
		{Name: "Vue.js", Category: "framework", Description: "Progressive JavaScript framework for UIs"},
		{Name: "Angular", Category: "framework", Description: "Full-featured frontend framework by Google"},
		{Name: "Next.js", Category: "framework", Description: "React meta-framework with SSR and routing"},
		{Name: "Django", Category: "framework", Description: "Batteries-included Python web framework"},
		{Name: "Express.js", Category: "framework", Description: "Minimal Node.js web framework"},
		{Name: "Spring Boot", Category: "framework", Description: "Java framework for production-grade applications"},
		{Name: "PostgreSQL", Category: "database", Description: "Advanced open-source relational database"},
		{Name: "MongoDB", Category: "database", Description: "Document-oriented NoSQL database"},
		{Name: "Redis", Category: "database", Description: "In-memory data structure store"},
		{Name: "Docker", Category: "devops", Description: "Container platform for packaging applications"},
		{Name: "Kubernetes", Category: "devops", Description: "Container orchestration system"},
		{Name: "AWS", Category: "devops", Description: "Amazon Web Services cloud platform"},
		{Name: "Git", Category: "tool", Description: "Distributed version control system"},
		{Name: "GraphQL", Category: "tool", Description: "Query language for APIs"},
		{Name: "REST API", Category: "concept", Description: "Architectural style for networked applications"},
		{Name: "System Design", Category: "concept", Description: "Designing large-scale distributed systems"},
	}

	for i := range skills {
		db.Where("name = ?", skills[i].Name).FirstOrCreate(&skills[i])
	}
	log.Info().Int("count", len(skills)).Msg("skills seeded")

	// Build a skill-name → ID lookup.
	skillMap := make(map[string]uint)
	for _, s := range skills {
		skillMap[s.Name] = s.ID
	}

	// ===================================================================
	// 2. Users
	// ===================================================================
	type userSeed struct {
		Email    string
		Username string
		FullName string
		Bio      string
		Github   string
		Skills   []string // skill names
		Levels   []string // proficiency per skill
		Years    []float64
	}

	templates := []userSeed{
		{"alex@example.com", "alexdev", "Alex Johnson", "Full-stack developer passionate about React and Go", "https://github.com/alexj", []string{"JavaScript", "React", "Go", "PostgreSQL", "Docker"}, []string{"advanced", "advanced", "intermediate", "intermediate", "beginner"}, []float64{6, 4, 2, 3, 1}},
		{"sarah@example.com", "sarahcodes", "Sarah Chen", "Backend engineer, Python and distributed systems enthusiast", "https://github.com/sarahc", []string{"Python", "Django", "PostgreSQL", "Docker", "AWS", "System Design"}, []string{"advanced", "advanced", "advanced", "intermediate", "intermediate", "intermediate"}, []float64{7, 5, 5, 3, 3, 4}},
		{"mike@example.com", "mikerust", "Mike Rodriguez", "Systems programmer exploring Rust and WebAssembly", "https://github.com/miker", []string{"Rust", "C++", "Go", "Docker", "Git"}, []string{"intermediate", "advanced", "beginner", "intermediate", "advanced"}, []float64{2, 6, 1, 2, 5}},
		{"emma@example.com", "emmaweb", "Emma Wilson", "Frontend specialist, React and TypeScript advocate", "https://github.com/emmaw", []string{"TypeScript", "React", "Next.js", "GraphQL", "Git"}, []string{"advanced", "advanced", "advanced", "intermediate", "advanced"}, []float64{4, 5, 3, 2, 6}},
		{"raj@example.com", "rajcloud", "Raj Patel", "DevOps engineer automating everything", "https://github.com/rajp", []string{"Docker", "Kubernetes", "AWS", "Python", "Go"}, []string{"advanced", "advanced", "advanced", "intermediate", "intermediate"}, []float64{5, 4, 5, 3, 2}},
		{"lisa@example.com", "lisadata", "Lisa Thompson", "Data engineer bridging Python and Java ecosystems", "https://github.com/lisat", []string{"Python", "Java", "PostgreSQL", "MongoDB", "Docker"}, []string{"advanced", "intermediate", "advanced", "intermediate", "beginner"}, []float64{5, 3, 4, 2, 1}},
		{"david@example.com", "davidjs", "David Kim", "JavaScript polyglot: Node, React, Vue", "https://github.com/davidk", []string{"JavaScript", "TypeScript", "React", "Vue.js", "Express.js", "MongoDB"}, []string{"advanced", "advanced", "advanced", "intermediate", "advanced", "intermediate"}, []float64{8, 5, 6, 2, 5, 3}},
		{"nina@example.com", "ninaml", "Nina Petrova", "ML engineer with strong Python fundamentals", "https://github.com/ninap", []string{"Python", "Java", "PostgreSQL", "Docker", "AWS"}, []string{"advanced", "beginner", "intermediate", "intermediate", "beginner"}, []float64{6, 1, 3, 2, 1}},
		{"carlos@example.com", "carlosgo", "Carlos Mendez", "Go developer building microservices", "https://github.com/carlosm", []string{"Go", "Docker", "Kubernetes", "PostgreSQL", "Redis", "REST API"}, []string{"advanced", "advanced", "intermediate", "advanced", "intermediate", "advanced"}, []float64{4, 4, 2, 5, 2, 5}},
		{"yuki@example.com", "yukidev", "Yuki Tanaka", "Full-stack TypeScript developer", "https://github.com/yukit", []string{"TypeScript", "React", "Next.js", "Express.js", "PostgreSQL"}, []string{"advanced", "advanced", "intermediate", "advanced", "intermediate"}, []float64{4, 4, 2, 4, 2}},
		{"omar@example.com", "omarjava", "Omar Hassan", "Enterprise Java and Spring Boot specialist", "https://github.com/omarh", []string{"Java", "Spring Boot", "PostgreSQL", "Docker", "Kubernetes"}, []string{"advanced", "advanced", "advanced", "intermediate", "beginner"}, []float64{8, 6, 5, 2, 1}},
		{"anna@example.com", "annapy", "Anna Kowalski", "Python web developer with Django expertise", "https://github.com/annak", []string{"Python", "Django", "JavaScript", "PostgreSQL", "Redis"}, []string{"advanced", "advanced", "intermediate", "intermediate", "beginner"}, []float64{5, 4, 3, 3, 1}},
		{"james@example.com", "jamesops", "James O'Brien", "Platform engineer: Kubernetes, Terraform, Go", "https://github.com/jamesob", []string{"Go", "Docker", "Kubernetes", "AWS", "Git"}, []string{"intermediate", "advanced", "advanced", "advanced", "advanced"}, []float64{2, 5, 4, 6, 7}},
		{"priya@example.com", "priyaui", "Priya Sharma", "UI/UX focused frontend developer", "https://github.com/priyas", []string{"JavaScript", "TypeScript", "React", "Angular", "Git"}, []string{"advanced", "intermediate", "advanced", "intermediate", "advanced"}, []float64{5, 2, 4, 2, 4}},
		{"tom@example.com", "tomrust", "Tom Baker", "Low-level programming: Rust, C++, systems design", "https://github.com/tomb", []string{"Rust", "C++", "Go", "System Design", "Git"}, []string{"advanced", "advanced", "intermediate", "advanced", "advanced"}, []float64{4, 7, 2, 5, 6}},
	}

	// Generate extra users if --users exceeds the template count.
	extraLangs := []string{"JavaScript", "Python", "Go", "TypeScript", "Java", "Rust"}
	for i := len(templates); i < *userCount; i++ {
		idx := i + 1
		sk := []string{extraLangs[rand.Intn(len(extraLangs))], "Git", "Docker"}
		templates = append(templates, userSeed{
			Email:    fmt.Sprintf("user%d@example.com", idx),
			Username: fmt.Sprintf("user%d", idx),
			FullName: fmt.Sprintf("Sample User %d", idx),
			Bio:      "Eager developer looking to learn and collaborate",
			Github:   fmt.Sprintf("https://github.com/user%d", idx),
			Skills:   sk,
			Levels:   []string{"beginner", "intermediate", "beginner"},
			Years:    []float64{1, 2, 0.5},
		})
	}

	users := make([]domain.User, 0, len(templates))
	for _, t := range templates {
		u := domain.User{
			Email:        t.Email,
			Username:     t.Username,
			PasswordHash: passwordHash,
			FullName:     t.FullName,
			Bio:          t.Bio,
			GithubURL:    t.Github,
			Badges:       domain.JSONB("[]"),
		}
		res := db.Where("email = ?", u.Email).FirstOrCreate(&u)
		if res.Error != nil {
			log.Warn().Err(res.Error).Str("email", u.Email).Msg("skip user")
			continue
		}
		users = append(users, u)

		// Assign skills.
		for j, skillName := range t.Skills {
			sid, ok := skillMap[skillName]
			if !ok {
				continue
			}
			us := domain.UserSkill{
				UserID:           u.ID,
				SkillID:          sid,
				ProficiencyLevel: domain.ProficiencyLevel(t.Levels[j]),
				YearsExperience:  t.Years[j],
			}
			db.Where("user_id = ? AND skill_id = ?", u.ID, sid).FirstOrCreate(&us)
		}

		// Bootstrap reputation row.
		rep := domain.UserReputation{
			UserID:                 u.ID,
			SkillCredibilityScores: domain.JSONB("{}"),
		}
		db.Where("user_id = ?", u.ID).FirstOrCreate(&rep)
	}
	log.Info().Int("count", len(users)).Msg("users seeded")

	if len(users) < 4 {
		log.Warn().Msg("not enough users for match/message seed, stopping")
		return
	}

	// ===================================================================
	// 3. Matches (5)
	// ===================================================================
	matchPairs := [][2]int{{0, 1}, {2, 3}, {0, 4}, {1, 5}, {6, 7}}
	matchInsights := []string{
		`{"overall_reasoning":"Alex and Sarah complement each other with frontend and backend strengths","skill_complement":"React+Go meets Python+Django","learning_opportunities":["Sarah can learn React","Alex can learn distributed systems"],"collaboration_ideas":["Build a full-stack dashboard","Microservices project"],"recommendation":"pair"}`,
		`{"overall_reasoning":"Mike and Emma bridge systems and frontend worlds","skill_complement":"Rust+C++ meets TypeScript+React","learning_opportunities":["Mike can learn modern frontend","Emma can learn systems programming"],"collaboration_ideas":["WebAssembly project","Performance-critical UI"],"recommendation":"pair"}`,
		`{"overall_reasoning":"Alex and Raj cover full-stack and DevOps","skill_complement":"React+Go meets Kubernetes+AWS","learning_opportunities":["Alex can learn cloud architecture","Raj can learn frontend"],"collaboration_ideas":["Deploy a Go service to EKS","CI/CD pipeline project"],"recommendation":"pair"}`,
		`{"overall_reasoning":"Sarah and Lisa share Python but differ in DB expertise","skill_complement":"Django+AWS meets Java+MongoDB","learning_opportunities":["Sarah can learn Java","Lisa can learn cloud deployment"],"collaboration_ideas":["Data pipeline project","Multi-DB application"],"recommendation":"consider"}`,
		`{"overall_reasoning":"David and Nina combine JavaScript breadth with ML depth","skill_complement":"Node+Vue meets Python+ML","learning_opportunities":["David can learn ML basics","Nina can learn full-stack JS"],"collaboration_ideas":["ML-powered web app","Real-time data dashboard"],"recommendation":"pair"}`,
	}

	matchScores := []float64{87.5, 72.3, 81.0, 65.4, 78.9}
	var createdMatches []domain.Match

	for i, pair := range matchPairs {
		if pair[0] >= len(users) || pair[1] >= len(users) {
			continue
		}
		m := domain.Match{
			User1ID:    users[pair[0]].ID,
			User2ID:    users[pair[1]].ID,
			MatchScore: matchScores[i],
			AIInsights: domain.JSONB(matchInsights[i]),
			Status:     domain.MatchActive,
		}
		db.Where("user1_id = ? AND user2_id = ?", m.User1ID, m.User2ID).FirstOrCreate(&m)
		createdMatches = append(createdMatches, m)
	}
	log.Info().Int("count", len(createdMatches)).Msg("matches seeded")

	// ===================================================================
	// 4. Messages (10)
	// ===================================================================
	if len(createdMatches) > 0 {
		m := createdMatches[0] // Alex ↔ Sarah
		msgs := []struct {
			senderIdx int
			content   string
		}{
			{0, "Hey Sarah! I saw we matched. I'd love to learn more about your distributed systems experience."},
			{1, "Hi Alex! Absolutely. I've been meaning to pick up React too, so this could work great for both of us."},
			{0, "Perfect. I was thinking we could start with a project — maybe a real-time dashboard?"},
			{1, "Love that idea. We could use Go for the backend and React for the frontend."},
			{0, "Exactly what I was thinking. Should we set up a coding session this week?"},
			{1, "How about Thursday evening? I usually have a couple of hours free."},
			{0, "Thursday works. I'll set up a repo beforehand with the basic Go structure."},
			{1, "Sounds good. I'll sketch out the Python data pipeline we can integrate later."},
			{0, "Looking forward to it! This is going to be a great collaboration."},
			{1, "Same here! See you Thursday."},
		}

		for _, msg := range msgs {
			var senderID, receiverID string
			if msg.senderIdx == 0 {
				senderID, receiverID = m.User1ID, m.User2ID
			} else {
				senderID, receiverID = m.User2ID, m.User1ID
			}
			dm := domain.Message{
				SenderID:   senderID,
				ReceiverID: receiverID,
				MatchID:    m.ID,
				Content:    msg.content,
			}
			db.Where("match_id = ? AND sender_id = ? AND content = ?", m.ID, senderID, msg.content).
				FirstOrCreate(&dm)
		}
		log.Info().Msg("messages seeded (10)")
	}

	// ===================================================================
	// 5. Coding sessions + ratings (5 each)
	// ===================================================================
	if len(createdMatches) >= 3 {
		now := time.Now()

		type sessionSeed struct {
			matchIdx int
			started  time.Time
			duration int
			success  float64
			notes    string
		}
		sessions := []sessionSeed{
			{0, now.Add(-7 * 24 * time.Hour), 90, 0.85, "Built the initial Go backend structure together. Great pair programming session."},
			{0, now.Add(-5 * 24 * time.Hour), 120, 0.92, "Implemented the React dashboard. Sarah learned JSX quickly."},
			{1, now.Add(-4 * 24 * time.Hour), 60, 0.75, "Explored Rust concepts and set up a small CLI tool."},
			{2, now.Add(-3 * 24 * time.Hour), 75, 0.88, "Containerized Alex's Go service. Raj walked through Dockerfile best practices."},
			{0, now.Add(-1 * 24 * time.Hour), 45, 0.90, "Quick session to add WebSocket support. Both learned gorilla/websocket."},
		}

		var createdSessions []domain.CodingSession
		for _, s := range sessions {
			if s.matchIdx >= len(createdMatches) {
				continue
			}
			ended := s.started.Add(time.Duration(s.duration) * time.Minute)
			cs := domain.CodingSession{
				MatchID:         createdMatches[s.matchIdx].ID,
				StartedAt:       s.started,
				EndedAt:         &ended,
				DurationMinutes: s.duration,
				CodeSnapshots:   domain.JSONB("[]"),
				SessionNotes:    s.notes,
				SuccessRating:   s.success,
			}
			db.Where("match_id = ? AND started_at = ?", cs.MatchID, cs.StartedAt).FirstOrCreate(&cs)
			createdSessions = append(createdSessions, cs)
		}
		log.Info().Int("count", len(createdSessions)).Msg("coding sessions seeded")

		// Ratings — one per session, rated by the other participant.
		ratingData := []struct {
			sessionIdx int
			overall    int
			code       int
			comm       int
			help       int
			reli       int
			comment    string
		}{
			{0, 5, 4, 5, 5, 5, "Alex is a fantastic pair programmer. Very patient explaining Go concepts."},
			{1, 5, 5, 5, 4, 5, "Sarah picked up React amazingly fast. Great communicator."},
			{2, 4, 4, 3, 4, 4, "Mike knows Rust deeply. Could improve on explaining concepts to beginners."},
			{3, 5, 5, 5, 5, 5, "Raj made Docker click for me. Best DevOps mentor I've worked with."},
			{4, 4, 5, 4, 4, 5, "Efficient session. We got WebSocket support done in under an hour."},
		}

		for _, rd := range ratingData {
			if rd.sessionIdx >= len(createdSessions) {
				continue
			}
			cs := createdSessions[rd.sessionIdx]
			m := createdMatches[sessions[rd.sessionIdx].matchIdx]

			// Rater = user2, rated = user1 (simple convention for seed).
			r := domain.Rating{
				RaterID:             m.User2ID,
				RatedID:             m.User1ID,
				SessionID:           cs.ID,
				OverallRating:       rd.overall,
				CodeQualityRating:   rd.code,
				CommunicationRating: rd.comm,
				HelpfulnessRating:   rd.help,
				ReliabilityRating:   rd.reli,
				Comment:             rd.comment,
			}
			db.Where("rater_id = ? AND rated_id = ? AND session_id = ?", r.RaterID, r.RatedID, r.SessionID).
				FirstOrCreate(&r)
		}
		log.Info().Int("count", len(ratingData)).Msg("ratings seeded")
	}

	log.Info().Msg("seed complete")
}
