.PHONY: dev dev-up dev-down build run test migrate seed lint clean

# Development
dev: dev-up
	@echo "SkillSync is running at http://localhost:3000"

dev-up:
	docker compose up -d --build

dev-down:
	docker compose down

# Backend
build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run

# Database
migrate:
	go run ./cmd/api migrate

seed:
	go run ./scripts/seed.go

setup-db:
	bash scripts/setup-database.sh

# Production
prod-up:
	docker compose -f docker-compose.prod.yml up -d --build

prod-down:
	docker compose -f docker-compose.prod.yml down

deploy-backend:
	bash scripts/deploy-backend.sh

deploy-frontend:
	bash scripts/deploy-frontend.sh

# Cleanup
clean:
	rm -rf bin/
	docker compose down -v
