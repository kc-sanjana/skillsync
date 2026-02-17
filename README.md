# SkillSync

A peer-to-peer skill exchange platform where users teach what they know and learn what they want. Powered by Claude AI for skill assessments and pairing insights.

## Features

- **Skill Matching** — Find partners based on complementary skills with AI-scored compatibility
- **Real-time Chat** — WebSocket-powered messaging within matched pairs
- **AI Assessments** — Claude evaluates your skill level through interactive Q&A
- **Pairing Insights** — AI-generated analysis of match compatibility, learning plans, and suggested topics
- **Reputation System** — Multi-dimensional ratings (communication, knowledge, helpfulness) with badges and leaderboard

## Tech Stack

| Layer      | Technology                          |
|------------|-------------------------------------|
| Backend    | Go, Echo, PostgreSQL                |
| Frontend   | React, TypeScript, Vite             |
| AI         | Claude API (Anthropic)              |
| Realtime   | WebSocket (gorilla/websocket)       |
| Auth       | JWT (golang-jwt)                    |
| Deploy     | Docker, GitHub Actions              |

## Quick Start

```bash
cp .env.example .env          # Configure your environment
make dev                      # Start all services
make seed                     # Seed sample data
```

Open http://localhost:3000

## Project Structure

```
skillsync/
├── cmd/api/                  # Application entrypoint
├── internal/
│   ├── domain/               # Domain models
│   ├── handler/              # HTTP handlers
│   ├── service/              # Business logic
│   ├── repository/           # Database access
│   ├── middleware/            # Auth, CORS, logging, security
│   └── websocket/            # WebSocket hub and client
├── pkg/                      # Shared packages (database, auth, logger)
├── config/                   # Configuration loading
├── migrations/               # SQL migrations
├── scripts/                  # Seed data, deployment scripts
├── frontend/                 # React TypeScript app
│   ├── src/
│   │   ├── components/       # RatingModal, InsightsCard, ReputationDisplay
│   │   ├── pages/            # Login, Register, Dashboard, Users, Chat, etc.
│   │   ├── services/         # API client, auth helpers
│   │   ├── hooks/            # useWebSocket, useAuth
│   │   ├── contexts/         # AuthContext
│   │   └── types/            # TypeScript interfaces
│   └── nginx.conf
├── docs/                     # API and deployment docs
├── docker-compose.yml        # Local development
├── docker-compose.prod.yml   # Production
├── Makefile                  # Common commands
└── .github/workflows/        # CI/CD pipeline
```

## Documentation

- [API Reference](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
