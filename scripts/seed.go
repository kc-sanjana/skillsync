package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/skillsync?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	seedUsers := []struct {
		email, username, password, fullName, bio string
		teach, learn                             []string
	}{
		{"alice@example.com", "alice", "password123", "Alice Johnson", "Full-stack developer passionate about Go and React",
			[]string{"Go", "React", "PostgreSQL"}, []string{"Rust", "Machine Learning"}},
		{"bob@example.com", "bob", "password123", "Bob Smith", "Data scientist exploring web technologies",
			[]string{"Python", "Machine Learning", "Data Analysis"}, []string{"Go", "React"}},
		{"carol@example.com", "carol", "password123", "Carol Williams", "DevOps engineer and cloud enthusiast",
			[]string{"Docker", "Kubernetes", "AWS"}, []string{"Go", "Python"}},
		{"dave@example.com", "dave", "password123", "Dave Brown", "Mobile developer learning backend",
			[]string{"React Native", "TypeScript", "Swift"}, []string{"Go", "Docker"}},
		{"eve@example.com", "eve", "password123", "Eve Davis", "Systems programmer getting into web dev",
			[]string{"Rust", "C++", "Linux"}, []string{"React", "TypeScript"}},
	}

	for _, u := range seedUsers {
		hash, _ := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		_, err := db.Exec(`
			INSERT INTO users (email, username, password_hash, full_name, bio, skills_teach, skills_learn)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) DO NOTHING`,
			u.email, u.username, string(hash), u.fullName, u.bio, u.teach, u.learn,
		)
		if err != nil {
			log.Printf("Failed to seed user %s: %v", u.username, err)
		} else {
			fmt.Printf("Seeded user: %s\n", u.username)
		}
	}

	fmt.Println("Seeding complete!")
}
