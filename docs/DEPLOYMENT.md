# SkillSync Deployment Guide

## Prerequisites

- Docker & Docker Compose
- PostgreSQL 16+
- Go 1.23+
- Node.js 20+
- Claude API key from Anthropic

## Local Development

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your Claude API key and other settings.

3. Start all services:
   ```bash
   make dev
   ```

4. Seed the database (optional):
   ```bash
   make seed
   ```

5. Access the app at http://localhost:3000

## Production Deployment

1. Set production environment variables:
   ```bash
   export DB_USER=skillsync
   export DB_PASSWORD=<strong-password>
   export DB_NAME=skillsync
   export JWT_SECRET=<strong-secret>
   export CLAUDE_API_KEY=<your-key>
   export ALLOWED_ORIGINS=https://yourdomain.com
   ```

2. Deploy with Docker Compose:
   ```bash
   make prod-up
   ```

3. Or deploy individual services:
   ```bash
   make deploy-backend
   make deploy-frontend
   ```

## Database

- Setup: `make setup-db`
- Migrations run automatically on API startup
- Manual migration: `make migrate`

## CI/CD

GitHub Actions workflow runs on push/PR to `main`:
1. Runs Go tests with PostgreSQL service
2. Builds frontend
3. Deploys on merge to main (requires `REGISTRY` secret)
