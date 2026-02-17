package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/yourusername/skillsync/config"
	"github.com/yourusername/skillsync/internal/handler"
	"github.com/yourusername/skillsync/internal/middleware"
	"github.com/yourusername/skillsync/internal/repository"
	"github.com/yourusername/skillsync/internal/service"
	ws "github.com/yourusername/skillsync/internal/websocket"
	"github.com/yourusername/skillsync/pkg/auth"
	"github.com/yourusername/skillsync/pkg/database"
	"github.com/yourusername/skillsync/pkg/logger"
)

func main() {
	// 🔹 Load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()
	appLogger := logger.New(cfg.LogLevel)

	// 🔹 Database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		appLogger.Fatal("Failed to run migrations", "error", err)
	}

	// 🔹 JWT
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	// 🔹 Repositories
	userRepo := repository.NewUserRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	ratingRepo := repository.NewRatingRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// 🔹 Services
	userService := service.NewUserService(userRepo)
	claudeService := service.NewClaudeService(cfg.ClaudeAPIKey)
	reputationService := service.NewReputationService(ratingRepo, userRepo)
	matchService := service.NewMatchService(matchRepo, userRepo, claudeService)
	pairingInsightsService := service.NewPairingInsightsService(claudeService, sessionRepo, matchRepo)

	// 🔹 WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// 🔹 Echo setup
	e := echo.New()
	e.HideBanner = true

	// 🔹 Logger
	e.Use(middleware.Logger(appLogger))

	// 🔥 CORS (uses ALLOWED_ORIGINS env var, falls back to localhost for dev)
	corsOrigins := []string{cfg.AllowedOrigins[0]}
	if cfg.Environment == "development" {
		corsOrigins = append(corsOrigins, "http://localhost:5173", "http://localhost:3000")
	}
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: corsOrigins,
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
	}))

	// 🔹 Security
	e.Use(middleware.Security())

	// 🔹 OAuth service
	oauthService := service.NewOAuthService(userService)

	// 🔹 Handlers
	authHandler := handler.NewAuthHandler(userService, jwtManager)
	oauthHandler := handler.NewOAuthHandler(oauthService, jwtManager)
	userHandler := handler.NewUserHandler(userService, ratingRepo, matchRepo)
	matchHandler := handler.NewMatchHandler(matchService)
	assessmentHandler := handler.NewAssessmentHandler(claudeService, userService)
	reputationHandler := handler.NewReputationHandler(reputationService)
	insightsHandler := handler.NewInsightsHandler(pairingInsightsService)
	wsHandler := handler.NewWebSocketHandler(hub, messageRepo, jwtManager)

	// =========================
	// 🌐 ROUTES
	// =========================

	api := e.Group("/api/v1")

	// 🔓 Public routes
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.RefreshToken)

	// 🌐 OAuth routes
	api.GET("/auth/google/login", oauthHandler.GoogleLogin)
	api.GET("/auth/google/callback", oauthHandler.GoogleCallback)
	api.GET("/auth/github/login", oauthHandler.GitHubLogin)
	api.GET("/auth/github/callback", oauthHandler.GitHubCallback)

	// 🔒 Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))

	// ✅ USERS
	protected.GET("/users", userHandler.List)
	protected.GET("/users/me", userHandler.GetMe) // ⭐ FIX — required for frontend auth
	protected.GET("/users/:id", userHandler.GetByID)
	protected.PUT("/users/me", userHandler.UpdateProfile)
	protected.GET("/users/me/reputation", reputationHandler.GetMyReputation)

	// ✅ MATCHES
	protected.POST("/matches", matchHandler.Create)
	protected.GET("/matches", matchHandler.List)
	protected.GET("/matches/:id", matchHandler.GetByID)
	protected.PUT("/matches/:id/status", matchHandler.UpdateStatus)

	// ✅ ASSESSMENT
	protected.POST("/assessment", assessmentHandler.Evaluate)

	// ✅ RATINGS
	protected.POST("/ratings", reputationHandler.SubmitRating)
	protected.GET("/leaderboard", reputationHandler.Leaderboard)

	// ✅ AI INSIGHTS
	protected.GET("/insights/pairing/:matchId", insightsHandler.GetPairingInsights)

	// 🔌 WebSocket
	e.GET("/ws", wsHandler.HandleConnection)

	// 🔹 Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	appLogger.Info("Starting server", "port", port)

	if err := e.Start(":" + port); err != nil {
		appLogger.Fatal("Server failed", "error", err)
		os.Exit(1)
	}
}

