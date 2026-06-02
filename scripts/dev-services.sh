#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
POSTGRES_READY_TIMEOUT_SECONDS="${CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS:-180}"
MINIO_READY_TIMEOUT_SECONDS="${CARTULARY_MINIO_READY_TIMEOUT_SECONDS:-120}"
MINIO_BUCKET="${MINIO_BUCKET:-cartulary}"

usage() {
  echo "usage: dev-services.sh up|services-down|db-down|wait-postgres|wait-minio|wait|init-minio|db-up|db-reset|object-store-reset" >&2
}

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

cleanup_dry_run() {
  [[ "${CARTULARY_CLEANUP_DRY_RUN:-}" == "1" ]]
}

dry_run_line() {
  local action="$1"
  local identity="$2"
  local proof="$3"

  printf 'DRY-RUN %s %s %s\n' "$action" "$identity" "$proof"
}

require_destructive_confirm() {
  local target="$1"

  if cleanup_dry_run; then
    return 0
  fi

  if [[ "${CARTULARY_DESTRUCTIVE_CONFIRM:-}" != "$target" ]]; then
    printf 'refusing %s: set CARTULARY_DESTRUCTIVE_CONFIRM=%s or use CARTULARY_CLEANUP_DRY_RUN=1\n' "$target" "$target" >&2
    return 2
  fi
}

container_state() {
  local service="$1"
  local container_id

  container_id="$(compose ps -q "$service" 2>/dev/null || true)"
  if [[ -z "$container_id" ]]; then
    printf 'missing none'
    return 0
  fi

  docker inspect -f '{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container_id" 2>/dev/null || printf 'unknown unknown'
}

wait_postgres() {
  local start_time="$SECONDS"
  local status="unknown"
  local health="unknown"

  while (( SECONDS - start_time < POSTGRES_READY_TIMEOUT_SECONDS )); do
    if compose exec -T postgres pg_isready -U cartulary -d postgres >/dev/null 2>&1; then
      return 0
    fi

    read -r status health < <(container_state postgres)
    if [[ "$status" == "exited" || "$status" == "dead" || "$status" == "missing" ]]; then
      echo "postgres container is ${status} during readiness wait (health=${health})" >&2
      compose logs --no-color --tail 120 postgres >&2 || true
      return 1
    fi

    sleep 1
  done

  echo "postgres did not become ready after ${POSTGRES_READY_TIMEOUT_SECONDS}s (state=${status} health=${health})" >&2
  compose logs --no-color --tail 120 postgres >&2 || true
  return 1
}

# shellcheck disable=SC2016
minio_ready_command='mc alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null 2>&1 && mc ready local >/dev/null 2>&1'

wait_minio() {
  local start_time="$SECONDS"
  local status="unknown"
  local health="unknown"

  while (( SECONDS - start_time < MINIO_READY_TIMEOUT_SECONDS )); do
    if compose exec -T minio sh -c "$minio_ready_command"; then
      return 0
    fi

    read -r status health < <(container_state minio)
    if [[ "$status" == "exited" || "$status" == "dead" || "$status" == "missing" ]]; then
      echo "minio container is ${status} during readiness wait (health=${health})" >&2
      compose logs --no-color --tail 120 minio >&2 || true
      return 1
    fi

    sleep 1
  done

  echo "minio did not become ready after ${MINIO_READY_TIMEOUT_SECONDS}s (state=${status} health=${health})" >&2
  compose logs --no-color --tail 120 minio >&2 || true
  return 1
}

init_minio() {
  wait_minio
  # shellcheck disable=SC2016
  compose exec -T -e MINIO_BUCKET="$MINIO_BUCKET" minio sh -c '
    set -e
    mc alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null
    mc mb --ignore-existing "local/${MINIO_BUCKET}" >/dev/null
  '
}

services_up() {
  compose up -d postgres minio
  wait_postgres
  wait_minio
}

services_down() {
  if cleanup_dry_run; then
    dry_run_line "stop-services" "compose:${COMPOSE_FILE}" "local_dev_compose_services_preserve_named_volumes"
    return 0
  fi

  compose down --remove-orphans
}

db_up() {
  services_up
  init_minio
}

db_down() {
  services_down
}

db_reset() {
  local go_bin="${GO:-go}"
  local config_file="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
  local go_cache="${GOCACHE:-${GO_CACHE_DIR:-/tmp/cartulary-go-build}}"
  local go_mod_cache="${GOMODCACHE:-${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}}"

  require_destructive_confirm "db-reset"
  if cleanup_dry_run; then
    dry_run_line "start-service" "compose:${COMPOSE_FILE}:postgres" "local_dev_postgres_required_for_db_reset"
    dry_run_line "reset-database" "postgres:cartulary" "drop_recreate_and_migrate_local_database"
    return 0
  fi

  compose up -d postgres
  wait_postgres
  printf '%s\n' 'db-reset: database reset only; MinIO/object storage is not reset.'
  compose exec -T postgres psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS cartulary;"
  compose exec -T postgres psql -U cartulary -d postgres -c "CREATE DATABASE cartulary;"
  cd "$ROOT_DIR"
  env CARTULARY_CONFIG_FILE="$config_file" \
    CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="postgres://cartulary:cartulary@localhost:5432/cartulary?sslmode=disable" \
    GOCACHE="$go_cache" GOMODCACHE="$go_mod_cache" \
    "$go_bin" run ./cmd/migrate up
}

object_store_reset() {
  require_destructive_confirm "object-store-reset"
  if cleanup_dry_run; then
    dry_run_line "start-service" "compose:${COMPOSE_FILE}:minio" "local_dev_minio_required_for_object_store_reset"
    dry_run_line "reset-object-store" "minio-bucket:${MINIO_BUCKET}" "delete_objects_and_preserve_bucket"
    return 0
  fi

  compose up -d minio
  wait_minio
  # shellcheck disable=SC2016
  compose exec -T -e MINIO_BUCKET="$MINIO_BUCKET" minio sh -c '
    set -e
    mc alias set local http://127.0.0.1:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null
    mc mb --ignore-existing "local/${MINIO_BUCKET}" >/dev/null
    mc rm --recursive --force "local/${MINIO_BUCKET}" >/dev/null
    mc mb --ignore-existing "local/${MINIO_BUCKET}" >/dev/null
  '
}

case "${1:-}" in
  up)
    services_up
    ;;
  services-down)
    services_down
    ;;
  db-down)
    db_down
    ;;
  wait-postgres)
    wait_postgres
    ;;
  wait-minio)
    wait_minio
    ;;
  wait)
    wait_postgres
    wait_minio
    ;;
  init-minio)
    init_minio
    ;;
  db-up)
    db_up
    ;;
  db-reset)
    db_reset
    ;;
  object-store-reset)
    object_store_reset
    ;;
  *)
    usage
    exit 2
    ;;
esac
