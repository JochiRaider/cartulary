#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
DEV_SERVICES_SCRIPT="${CARTULARY_DEV_SERVICES_SCRIPT:-$ROOT_DIR/tools/harness/readiness/dev-services.sh}"
MIGRATIONS_DIR="${CARTULARY_MIGRATIONS_DIR:-$ROOT_DIR/db/migrations}"
GO_BIN="${GO:-go}"
NODE_BIN="${NODE:-node}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
CONFIG_FILE="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
export GOCACHE="${GOCACHE:-/tmp/cartulary-go-build}"
export GOMODCACHE="${GOMODCACHE:-/tmp/cartulary-go-mod}"
EMPTY_DB="cartulary_migration_empty_$$"
PRE_RECORD_ENVELOPE_DB="cartulary_migration_pre_record_envelope_$$"
PRE_ASSESSMENTS_CORE02_DB="cartulary_migration_pre_assessments_core02_$$"
PENULTIMATE_DB="cartulary_migration_penultimate_$$"
MODE="all"

fail() {
  echo "migration verification failed: $*" >&2
  exit 1
}

usage() {
  cat >&2 <<'USAGE'
usage: check-migrations.sh [--mode input|scratch|all]

Modes:
  input    Validate migration filenames and static upgrade-path anchors only.
  scratch  Apply migrations against scratch Postgres databases.
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

previous_migration_version_before_anchor() {
  local anchor_glob="$1"
  local boundary_name="$2"
  local anchor_index="-1"
  local match_count="0"
  local index
  local name

  for index in "${!MIGRATION_FILES[@]}"; do
    name="$(basename "${MIGRATION_FILES[$index]}")"
    # anchor_glob is an intentional migration filename glob, not a literal.
    # shellcheck disable=SC2053
    if [[ "$name" == $anchor_glob ]]; then
      anchor_index="$index"
      match_count=$((match_count + 1))
    fi
  done

  if ((match_count == 0)); then
    fail "missing migration anchor for ${boundary_name}: ${anchor_glob}"
  fi
  if ((match_count > 1)); then
    fail "multiple migration anchors for ${boundary_name}: ${anchor_glob}"
  fi
  if ((anchor_index == 0)); then
    fail "migration anchor for ${boundary_name} has no preceding migration: ${anchor_glob}"
  fi

  migration_version_at_index "$((anchor_index - 1))"
}

cleanup() {
  local db_name
  for db_name in "$EMPTY_DB" "$PRE_RECORD_ENVELOPE_DB" "$PRE_ASSESSMENTS_CORE02_DB" "$PENULTIMATE_DB"; do
    docker compose -f "$COMPOSE_FILE" exec -T postgres \
      psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"$db_name\";" >/dev/null 2>&1 || true
  done
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

  PRE_RECORD_ENVELOPE_VERSION=""
  PRE_ASSESSMENTS_CORE02_VERSION=""
  if [ "$MIGRATION_COUNT" -ge 2 ]; then
    PRE_RECORD_ENVELOPE_VERSION="$(previous_migration_version_before_anchor "*_phase4_record_envelope_backfill.sql" "pre-record-envelope boundary")"
    PRE_ASSESSMENTS_CORE02_VERSION="$(previous_migration_version_before_anchor "*_phase4_assessments_core02.sql" "pre-assessments-Core02 boundary")"
  fi
}

run_input_validation() {
  "$NODE_BIN" "$ROOT_DIR/tools/harness/backend/migration-history-cli.mjs"
  load_and_validate_migration_inputs
  echo "migration input drift: validated ${MIGRATION_COUNT} migration files and static upgrade-path anchors"
}

run_scratch_apply() {
  load_and_validate_migration_inputs
  trap cleanup EXIT

  docker compose -f "$COMPOSE_FILE" up -d postgres >/dev/null
  "$DEV_SERVICES_SCRIPT" wait-postgres

  echo "migration verification: empty database apply to head"
  create_database "$EMPTY_DB"
  run_migrate "$EMPTY_DB" up

  if [[ -n "$PRE_RECORD_ENVELOPE_VERSION" ]]; then
    run_upgrade_boundary "$PRE_RECORD_ENVELOPE_DB" "pre-record-envelope boundary" "$PRE_RECORD_ENVELOPE_VERSION"
  else
    echo "upgrade-path coverage limited: fewer than two migrations exist; skipping pre-record-envelope boundary" >&2
  fi
  if [[ -n "$PRE_ASSESSMENTS_CORE02_VERSION" ]]; then
    run_upgrade_boundary "$PRE_ASSESSMENTS_CORE02_DB" "pre-assessments-Core02 boundary" "$PRE_ASSESSMENTS_CORE02_VERSION"
  else
    echo "upgrade-path coverage limited: fewer than two migrations exist; skipping pre-assessments-Core02 boundary" >&2
  fi

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
    cd "$ROOT_DIR"
    export CARTULARY_CONFIG_FILE="$CONFIG_FILE"
    export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="postgres://cartulary:cartulary@localhost:5432/$db_name?sslmode=disable"
    if [[ -n "$MIGRATE_BIN" && -x "$MIGRATE_BIN" ]]; then
      "$MIGRATE_BIN" "$command" "$@"
    else
      "$GO_BIN" run ./cmd/migrate "$command" "$@"
    fi
  )
}

run_upgrade_boundary() {
  local db_name="$1"
  local boundary_name="$2"
  local boundary_version="$3"

  echo "migration verification: upgrade path from ${boundary_name}"
  create_database "$db_name"
  run_migrate "$db_name" up-to "$boundary_version"
  run_migrate "$db_name" up
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
