#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:?Usage: deploy.sh <commit-sha>}"
DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
ENV_FILE="${DEPLOY_DIR}/.env"

if [[ ! -f "${COMPOSE_FILE}" ]]; then
  echo "Missing Compose file: ${COMPOSE_FILE}" >&2
  exit 1
fi

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing environment file: ${ENV_FILE}" >&2
  exit 1
fi

export BACKEND_IMAGE="ghcr.io/askling-for-it/digrone-backend:${VERSION}"
export FRONTEND_IMAGE="ghcr.io/askling-for-it/digrone-frontend:${VERSION}"
export APP_VERSION="${VERSION}"

docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" pull
docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" up -d --remove-orphans
docker image prune -f
