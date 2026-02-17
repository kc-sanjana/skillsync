package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/yourusername/skillsync/internal/handler"
	"github.com/yourusername/skillsync/internal/middleware"
	"github.com/yourusername/skillsync/internal/service"
	ws "github.com/yourusername/skillsync/internal/websocket"
	"github.com/yourusername/skillsync/pkg/database"
)

// ---------------------------------------------------------------------------
// Echo validator adapter
// ---------------------------------------------------------------------------

type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	// ---- env ----
	if err := godotenv.Load(); err != nil {
		// Not fatal: in production env vars are set by the orchestrator.
		fmt.Println("no .env file found, reading from environment")
	}

	// ---- logger ----
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	log.Info().Msg("starting skillsync api server")

	// ---- database ----
	db, err := database.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// ---- services ----
	claudeService := service.NewClaudeService()
	userService := service.NewUserService(db)
	matchService := service.NewMatchService(db, claudeService)
	repService := service.NewReputationService(db)

	// ---- websocket hub ----
	hub := ws.NewHub()
	go hub.Run()

	// ---- services (oauth) ----
	oauthService := service.NewOAuthService(db, userService)

	// ---- handlers ----
	authHandler := handler.NewAuthHandler(userService)
	oauthHandler := handler.NewOAuthHandler(oauthService)
	userHandler := handler.NewUserHandler(userService)
	assessmentHandler := handler.NewAssessmentHandler(claudeService, db)
	matchHandler := handler.NewMatchHandler(matchService, claudeService, db)
	repHandler := handler.NewReputationHandler(repService, db)
	wsHandler := handler.NewWebSocketHandler(hub, db)
	msgHandler := handler.NewMessageHandler(db, hub)

	// ---- echo ----
	e := echo.New()
	e.HideBanner = true
	e.Validator = &customValidator{v: validator.New()}

	// ---- global middleware ----
	e.Use(middleware.RequestLoggerMiddleware())
	e.Use(echomw.Recover())
	e.Use(middleware.CORSMiddleware())
	e.Use(middleware.SecurityHeadersMiddleware())
	e.Use(middleware.RateLimitMiddleware(100, time.Minute))
	e.Use(middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024)) // 10 MB

	// ---- health routes ----
	e.GET("/health", healthCheck)
	e.GET("/health/db", healthDB)

	// ---- public auth routes ----
	api := e.Group("/api")
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// OAuth routes
	authGroup.GET("/google/login", oauthHandler.GoogleLogin)
	authGroup.GET("/google/callback", oauthHandler.GoogleCallback)
	authGroup.GET("/github/login", oauthHandler.GitHubLogin)
	authGroup.GET("/github/callback", oauthHandler.GitHubCallback)

	// ---- protected routes ----
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware())

	// Auth
	protected.GET("/auth/me", authHandler.GetMe)

	// Users
	protected.GET("/users", userHandler.GetUsers)
	protected.GET("/users/:id", userHandler.GetUser)
	protected.PUT("/users/:id", userHandler.UpdateUser)
	protected.POST("/users/:id/skills", userHandler.AddUserSkill)
	protected.GET("/users/:id/reputation", userHandler.GetUserReputation)

	// Assessments
	protected.POST("/assessments", assessmentHandler.SubmitCode)
	protected.POST("/assessments/hint", assessmentHandler.GetHint)
	protected.GET("/assessments/history", assessmentHandler.GetAssessmentHistory)
	protected.GET("/projects/suggestions", assessmentHandler.GetProjectSuggestions)

	// Matches
	protected.GET("/matches/suggestions", matchHandler.GetMatchSuggestions)
	protected.POST("/matches/request", matchHandler.SendMatchRequest)
	protected.PUT("/matches/request/:id/accept", matchHandler.AcceptMatchRequest)
	protected.PUT("/matches/request/:id/reject", matchHandler.RejectMatchRequest)
	protected.GET("/matches", matchHandler.GetMyMatches)
	protected.GET("/matches/requests/pending", matchHandler.GetPendingRequests)
	protected.GET("/matches/:id/insights", matchHandler.GetMatchInsights)
	protected.GET("/matches/:id/suggestions", matchHandler.GetCollaborationSuggestions)

	// Messages
	protected.GET("/matches/:matchId/messages", msgHandler.GetMessages)
	protected.POST("/messages", msgHandler.SendMessage)
	protected.PUT("/messages/read", msgHandler.MarkMessagesRead)

	// Reputation & Ratings
	protected.POST("/ratings", repHandler.SubmitRating)
	protected.POST("/sessions/:id/feedback", repHandler.SubmitSessionFeedback)
	protected.GET("/ratings/received", repHandler.GetMyRatings)
	protected.GET("/leaderboard", repHandler.GetLeaderboard)

	// WebSocket (auth is done inside the handler via query param)
	e.GET("/ws", wsHandler.HandleWebSocket)

	// ---- start server ----
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	go func() {
		log.Info().Str("addr", addr).Msg("server listening")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// ---- graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}

// ---------------------------------------------------------------------------
// Health check handlers
// ---------------------------------------------------------------------------

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func healthDB(c echo.Context) error {
	sqlDB, err := database.GetDB().DB()
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "cannot get database handle",
		})
	}
	if err := sqlDB.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "database ping failed",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
