#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.dev.yml"
DEV_SERVICES_SCRIPT="$ROOT_DIR/scripts/dev-services.sh"
GO_BIN="${GO:-go}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
CONFIG_FILE="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
export GOCACHE="${GOCACHE:-/tmp/cartulary-go-build}"
export GOMODCACHE="${GOMODCACHE:-/tmp/cartulary-go-mod}"
EMPTY_DB="cartulary_migration_empty_$$"
UPGRADE_DB="cartulary_migration_upgrade_$$"

cleanup() {
  local db_name
  for db_name in "$EMPTY_DB" "$UPGRADE_DB"; do
    docker compose -f "$COMPOSE_FILE" exec -T postgres \
      psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$db_name\";" >/dev/null 2>&1 || true
  done
}

trap cleanup EXIT

docker compose -f "$COMPOSE_FILE" up -d postgres >/dev/null
"$DEV_SERVICES_SCRIPT" wait-postgres

mapfile -t MIGRATION_FILES < <(find "$ROOT_DIR/db/migrations" -maxdepth 1 -type f -name '*.sql' | LC_ALL=C sort)
MIGRATION_COUNT="${#MIGRATION_FILES[@]}"
if [ "$MIGRATION_COUNT" -eq 0 ]; then
  echo "migration verification failed: no migration files present" >&2
  exit 1
fi

create_database() {
  local db_name="$1"
  docker compose -f "$COMPOSE_FILE" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$db_name\";" >/dev/null
  docker compose -f "$COMPOSE_FILE" exec -T postgres \
    psql -U cartulary -d postgres -c "CREATE DATABASE \"$db_name\";" >/dev/null
}

run_migrate() {
  local db_name="$1"
  local command="$2"
  (
    cd "$ROOT_DIR"
    export CARTULARY_CONFIG_FILE="$CONFIG_FILE"
    export CARTULARY_POSTGRES_DSN="postgres://cartulary:cartulary@localhost:5432/$db_name?sslmode=disable"
    if [[ -n "$MIGRATE_BIN" && -x "$MIGRATE_BIN" ]]; then
      "$MIGRATE_BIN" "$command"
    else
      "$GO_BIN" run ./cmd/migrate "$command"
    fi
  )
}

echo "migration verification: empty database apply to head"
create_database "$EMPTY_DB"
run_migrate "$EMPTY_DB" up

echo "migration verification: upgrade path from non-head boundary"
create_database "$UPGRADE_DB"
if [ "$MIGRATION_COUNT" -ge 2 ]; then
  PENULTIMATE_STEPS=$((MIGRATION_COUNT - 1))
  for _ in $(seq 1 "$PENULTIMATE_STEPS"); do
    run_migrate "$UPGRADE_DB" up-by-one
  done
  run_migrate "$UPGRADE_DB" up
else
  echo "upgrade-path coverage limited: only one migration exists; running best-available boundary" >&2
  run_migrate "$UPGRADE_DB" up
fi
