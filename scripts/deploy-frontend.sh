#!/usr/bin/env bash
set -euo pipefail

echo "==> Building frontend..."
docker build -f frontend/Dockerfile.prod -t skillsync-frontend:latest frontend/

echo "==> Pushing image..."
docker tag skillsync-frontend:latest "${REGISTRY:-registry.example.com}/skillsync-frontend:latest"
docker push "${REGISTRY:-registry.example.com}/skillsync-frontend:latest"

echo "==> Frontend deployment complete!"
