#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"
GO_BIN="${GO:-go}"
CONFIG_FILE="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
export GOCACHE="${GOCACHE:-/tmp/cartulary-go-build}"
export GOMODCACHE="${GOMODCACHE:-/tmp/cartulary-go-mod}"
SCRATCH_DB="cartulary_migration_check_$$"
READY=0

cleanup() {
  docker compose -f "$COMPOSE_FILE" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$SCRATCH_DB\";" >/dev/null 2>&1 || true
}

trap cleanup EXIT

docker compose -f "$COMPOSE_FILE" up -d postgres >/dev/null

for _ in $(seq 1 30); do
  if docker compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U cartulary -d postgres >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 1
done

if [ "$READY" -ne 1 ]; then
  echo "postgres did not become ready for migration verification" >&2
  exit 1
fi

docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$SCRATCH_DB\";" >/dev/null
docker compose -f "$COMPOSE_FILE" exec -T postgres \
  psql -U cartulary -d postgres -c "CREATE DATABASE \"$SCRATCH_DB\";" >/dev/null

if ! find "$ROOT_DIR/db/migrations" -maxdepth 1 -type f ! -name '.gitkeep' | grep -q .; then
  echo "migration verification failed: no migration files present" >&2
  exit 1
fi

cd "$ROOT_DIR"
CARTULARY_CONFIG_FILE="$CONFIG_FILE" \
CARTULARY_POSTGRES_DSN="postgres://cartulary:cartulary@localhost:5432/$SCRATCH_DB?sslmode=disable" \
  "$GO_BIN" run ./cmd/migrate up
