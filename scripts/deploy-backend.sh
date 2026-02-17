#!/usr/bin/env bash
set -euo pipefail

echo "==> Building backend..."
docker build -f Dockerfile.prod -t skillsync-api:latest .

echo "==> Pushing image..."
docker tag skillsync-api:latest "${REGISTRY:-registry.example.com}/skillsync-api:latest"
docker push "${REGISTRY:-registry.example.com}/skillsync-api:latest"

echo "==> Backend deployment complete!"
