package database

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Connect initialises the PostgreSQL connection with connection pooling.
// It reads configuration from environment variables:
//
//	DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
//	DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME (minutes)
func Connect() (*gorm.DB, error) {
	var connectErr error

	once.Do(func() {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "postgres")
		dbname := getEnv("DB_NAME", "skillsync")
		sslmode := getEnv("DB_SSLMODE", "disable")

		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)

		// Choose GORM log level based on environment.
		logLevel := logger.Warn
		if getEnv("APP_ENV", "development") == "development" {
			logLevel = logger.Info
		}

		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                 logger.Default.LogMode(logLevel),
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
		})
		if err != nil {
			connectErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Connection pool settings.
		sqlDB, err := db.DB()
		if err != nil {
			connectErr = fmt.Errorf("failed to get underlying sql.DB: %w", err)
			return
		}

		maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 25)
		maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 10)
		maxLifetime := getEnvInt("DB_CONN_MAX_LIFETIME", 30) // minutes

		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Minute)

		log.Info().
			Str("host", host).
			Str("port", port).
			Str("database", dbname).
			Int("max_open_conns", maxOpen).
			Int("max_idle_conns", maxIdle).
			Msg("database connected")
	})

	return db, connectErr
}

// Migrate runs GORM AutoMigrate for every domain model.
func Migrate() error {
	if db == nil {
		return fmt.Errorf("database not connected; call Connect() first")
	}

	log.Info().Msg("running database migrations")

	// Run incremental SQL migrations rather than AutoMigrate, because the
	// existing DB schema (UUID PKs, array columns) diverges from GORM's
	// model-based expectations.
	migrations := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS git_hub_id VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS github_url VARCHAR(512)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS linkedin_url VARCHAR(512)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS total_sessions BIGINT DEFAULT 0",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS badges JSONB DEFAULT '[]'",
		`DO $$ BEGIN
			ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
		EXCEPTION WHEN others THEN NULL;
		END $$`,
	}
	for _, stmt := range migrations {
		if err := db.Exec(stmt).Error; err != nil {
			log.Warn().Err(err).Str("stmt", stmt).Msg("migration statement failed (may be safe to ignore)")
		}
	}

	log.Info().Msg("database migrations completed")
	return nil
}

// Close gracefully shuts down the database connection pool.
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB for close: %w", err)
	}

	log.Info().Msg("closing database connection")
	return sqlDB.Close()
}

// GetDB returns the singleton *gorm.DB instance.
// Panics if Connect() has not been called.
func GetDB() *gorm.DB {
	if db == nil {
		panic("database not initialised: call database.Connect() first")
	}
	return db
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
