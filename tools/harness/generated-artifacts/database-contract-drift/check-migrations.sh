#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
DEV_SERVICES_SCRIPT="${CARTULARY_DEV_SERVICES_SCRIPT:-$ROOT_DIR/tools/harness/readiness/dev-services.sh}"
MIGRATIONS_DIR="${CARTULARY_MIGRATIONS_DIR:-$ROOT_DIR/db/migrations}"
GO_BIN="${GO:-go}"
NODE_BIN="${NODE:-node}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
CONFIG_FILE="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
EXPECTED_LINEAGE_ID="${CARTULARY_EXPECTED_MIGRATION_LINEAGE_ID:-cartulary.prod_ddl_rebaseline.v1}"
export GOCACHE="${GOCACHE:-/tmp/cartulary-go-build}"
export GOMODCACHE="${GOMODCACHE:-/tmp/cartulary-go-mod}"
EMPTY_DB="cartulary_migration_empty_$$"
PENULTIMATE_DB="cartulary_migration_penultimate_$$"
MODE="all"
MIGRATE_WORK_DIR=""

fail() {
  echo "migration verification failed: $*" >&2
  exit 1
}

usage() {
  cat >&2 <<'USAGE'
usage: check-migrations.sh [--mode input|scratch|all]

Modes:
  input    Validate migration filenames and migration history manifest shape.
  scratch  Apply migrations against scratch Postgres databases and validate lineage.
  all      Run both input validation and scratch apply evidence.
USAGE
  exit 2
}

while (($# > 0)); do
  case "$1" in
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --mode=*)
      MODE="${1#--mode=}"
      shift
      ;;
    input|scratch|all)
      MODE="$1"
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage
      ;;
  esac
done

case "$MODE" in
  input|scratch|all) ;;
  *) usage ;;
esac

migration_version_from_file() {
  local file="$1"
  local name
  local prefix

  name="$(basename "$file")"
  if [[ ! "$name" =~ ^([0-9]+)_.+\.sql$ ]]; then
    fail "invalid migration filename \"$name\": expected numeric goose prefix like 00009_name.sql"
  fi

  prefix="${BASH_REMATCH[1]}"
  printf "%d" "$((10#$prefix))"
}

migration_version_at_index() {
  local index="$1"

  if ((index < 0 || index >= MIGRATION_COUNT)); then
    fail "migration index ${index} is outside the available migration range 0..$((MIGRATION_COUNT - 1))"
  fi

  migration_version_from_file "${MIGRATION_FILES[$index]}"
}

cleanup() {
  local db_name
  for db_name in "$EMPTY_DB" "$PENULTIMATE_DB"; do
    docker compose -f "$COMPOSE_FILE" exec -T postgres \
      psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$db_name\";" >/dev/null 2>&1 || true
  done
  if [[ -n "$MIGRATE_WORK_DIR" ]]; then
    rm -rf "$MIGRATE_WORK_DIR"
  fi
}

load_and_validate_migration_inputs() {
  local migration_file

  mapfile -t MIGRATION_FILES < <(find "$MIGRATIONS_DIR" -maxdepth 1 -type f -name '*.sql' | LC_ALL=C sort)
  MIGRATION_COUNT="${#MIGRATION_FILES[@]}"
  if [ "$MIGRATION_COUNT" -eq 0 ]; then
    fail "no migration files present"
  fi
  for migration_file in "${MIGRATION_FILES[@]}"; do
    migration_version_from_file "$migration_file" >/dev/null
  done
}

run_input_validation() {
  "$NODE_BIN" "$ROOT_DIR/tools/harness/generated-artifacts/database-contract-drift/migration-history-cli.mjs"
  load_and_validate_migration_inputs
  echo "migration input drift: validated ${MIGRATION_COUNT} migration files and current-line manifest shape"
}

run_scratch_apply() {
  load_and_validate_migration_inputs
  MIGRATE_WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-migrate-cwd.XXXXXX")"
  trap cleanup EXIT

  docker compose -f "$COMPOSE_FILE" up -d postgres >/dev/null
  "$DEV_SERVICES_SCRIPT" wait-postgres

  echo "migration verification: empty database apply to head"
  create_database "$EMPTY_DB"
  run_migrate "$EMPTY_DB" up
  verify_lineage "$EMPTY_DB"

  echo "migration verification: upgrade path from penultimate boundary"
  create_database "$PENULTIMATE_DB"
  if [ "$MIGRATION_COUNT" -ge 2 ]; then
    PENULTIMATE_VERSION="$(migration_version_at_index "$((MIGRATION_COUNT - 2))")"
    run_migrate "$PENULTIMATE_DB" up-to "$PENULTIMATE_VERSION"
    run_migrate "$PENULTIMATE_DB" up
  else
    echo "upgrade-path coverage limited: only one migration exists; running best-available boundary" >&2
    run_migrate "$PENULTIMATE_DB" up
  fi
  verify_lineage "$PENULTIMATE_DB"
}

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
  shift 2
  (
    export CARTULARY_CONFIG_FILE="$CONFIG_FILE"
    export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="postgres://cartulary:cartulary@localhost:5432/$db_name?sslmode=disable"
    if [[ -n "$MIGRATE_BIN" && -x "$MIGRATE_BIN" ]]; then
      cd "$MIGRATE_WORK_DIR"
      "$MIGRATE_BIN" "$command" "$@"
    else
      cd "$ROOT_DIR"
      "$GO_BIN" run ./cmd/migrate "$command" "$@"
    fi
  )
}

verify_lineage() {
  local db_name="$1"
  local lineage

  lineage="$(
    docker compose -f "$COMPOSE_FILE" exec -T postgres \
      psql -U cartulary -d "$db_name" -Atc "SELECT COALESCE((SELECT lineage_id FROM schema_migration_lineage WHERE lineage_id = '${EXPECTED_LINEAGE_ID}' LIMIT 1), '');" |
      tr -d '[:space:]'
  )"
  if [[ "$lineage" != "$EXPECTED_LINEAGE_ID" ]]; then
    fail "database ${db_name} is missing expected migration lineage ${EXPECTED_LINEAGE_ID}"
  fi
  echo "migration verification: lineage marker present for ${db_name}"
}

case "$MODE" in
  input)
    run_input_validation
    ;;
  scratch)
    run_scratch_apply
    ;;
  all)
    run_input_validation
    run_scratch_apply
    ;;
esac
