#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
POSTGRES_READY_TIMEOUT_SECONDS="${CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS:-180}"
OBJECT_STORE_READY_TIMEOUT_SECONDS="${CARTULARY_OBJECT_STORE_READY_TIMEOUT_SECONDS:-120}"
SEAWEEDFS_S3_PORT="${SEAWEEDFS_S3_PORT:-8333}"
OBJECT_STORE_ENDPOINT="${OBJECT_STORE_ENDPOINT:-127.0.0.1:${SEAWEEDFS_S3_PORT}}"
OBJECT_STORE_BUCKET="${OBJECT_STORE_BUCKET:-cartulary}"
SEAWEEDFS_S3_ACCESS_KEY_ID="${SEAWEEDFS_S3_ACCESS_KEY_ID:-cartulary-local}"
SEAWEEDFS_S3_SECRET_ACCESS_KEY="${SEAWEEDFS_S3_SECRET_ACCESS_KEY:-cartulary-local-secret}"
OBJECT_STORE_SECURE="${OBJECT_STORE_SECURE:-false}"
OBJECT_STORE_CORS_ORIGIN="${OBJECT_STORE_CORS_ORIGIN:-http://127.0.0.1:5173}"
OBJECT_STORE_CORS_ALLOWED_ORIGINS="${OBJECT_STORE_CORS_ALLOWED_ORIGINS:-http://localhost:5173,http://127.0.0.1:5173}"
SEAWEEDFS_S3_IMAGE="${SEAWEEDFS_S3_IMAGE:-docker.io/chrislusf/seaweedfs:4.17}"
SEAWEEDFS_S3_IMAGE_DIGEST="${SEAWEEDFS_S3_IMAGE_DIGEST:-sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9}"
GO_BIN="${GO:-go}"
GO_CACHE="${GOCACHE:-${GO_CACHE_DIR:-/tmp/cartulary-go-build}}"
GO_MOD_CACHE="${GOMODCACHE:-${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}}"
export OBJECT_STORE_CORS_ALLOWED_ORIGINS

usage() {
  echo "usage: dev-services.sh up|services-down|db-down|wait-postgres|wait-object-store|wait|init-object-store|db-up|db-reset|object-store-reset" >&2
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

probe_object_store() {
  local mode="${1:-probe}"

  cd "$ROOT_DIR"
  env GOCACHE="$GO_CACHE" GOMODCACHE="$GO_MOD_CACHE" \
    "$GO_BIN" run ./tools/objectstoreprobe \
      --mode "$mode" \
      --endpoint "$OBJECT_STORE_ENDPOINT" \
      --access-key-id "$SEAWEEDFS_S3_ACCESS_KEY_ID" \
      --secret-access-key "$SEAWEEDFS_S3_SECRET_ACCESS_KEY" \
      --secure "$OBJECT_STORE_SECURE" \
      --bucket "$OBJECT_STORE_BUCKET" \
      --origin "$OBJECT_STORE_CORS_ORIGIN" \
      --service-name seaweedfs-s3 \
      --image "$SEAWEEDFS_S3_IMAGE" \
      --image-digest "$SEAWEEDFS_S3_IMAGE_DIGEST"
}

wait_object_store() {
  local start_time="$SECONDS"
  local status="unknown"
  local health="unknown"

  while (( SECONDS - start_time < OBJECT_STORE_READY_TIMEOUT_SECONDS )); do
    if probe_object_store probe >/dev/null 2>&1; then
      return 0
    fi

    read -r status health < <(container_state seaweedfs-s3)
    if [[ "$status" == "exited" || "$status" == "dead" || "$status" == "missing" ]]; then
      echo "seaweedfs-s3 container is ${status} during readiness wait (health=${health})" >&2
      compose logs --no-color --tail 120 seaweedfs-s3 >&2 || true
      return 1
    fi

    sleep 1
  done

  echo "seaweedfs-s3 did not become ready after ${OBJECT_STORE_READY_TIMEOUT_SECONDS}s (state=${status} health=${health})" >&2
  probe_object_store probe >&2 || true
  compose logs --no-color --tail 120 seaweedfs-s3 >&2 || true
  return 1
}

init_object_store() {
  wait_object_store
}

services_up() {
  compose up -d --remove-orphans postgres seaweedfs-s3
  wait_postgres
  wait_object_store
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
  init_object_store
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
  printf '%s\n' 'db-reset: database reset only; object storage is not reset.'
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
    dry_run_line "start-service" "compose:${COMPOSE_FILE}:seaweedfs-s3" "local_dev_object_store_required_for_object_store_reset"
    dry_run_line "reset-object-store" "object-store-bucket:${OBJECT_STORE_BUCKET}" "delete_objects_and_preserve_bucket"
    return 0
  fi

  compose up -d --remove-orphans seaweedfs-s3
  wait_object_store
  probe_object_store reset
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
  wait-object-store)
    wait_object_store
    ;;
  wait)
    wait_postgres
    wait_object_store
    ;;
  init-object-store)
    init_object_store
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
