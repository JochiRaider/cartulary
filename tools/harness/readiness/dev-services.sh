#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
POSTGRES_READY_TIMEOUT_SECONDS="${CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS:-180}"
LOCAL_POSTGRES_HOST="${CARTULARY_LOCAL_POSTGRES_HOST:-localhost}"
LOCAL_POSTGRES_PORT="${CARTULARY_LOCAL_POSTGRES_PORT:-5432}"
LOCAL_POSTGRES_DATABASE="${CARTULARY_LOCAL_POSTGRES_DATABASE:-cartulary}"
LOCAL_POSTGRES_USER="${CARTULARY_LOCAL_POSTGRES_USER:-cartulary}"
LOCAL_POSTGRES_PASSWORD="${CARTULARY_LOCAL_POSTGRES_PASSWORD:-cartulary}"
LOCAL_POSTGRES_SSLMODE="${CARTULARY_LOCAL_POSTGRES_SSLMODE:-disable}"
POSTGRES_PRIMARY_DSN="${CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN:-postgres://${LOCAL_POSTGRES_USER}:${LOCAL_POSTGRES_PASSWORD}@${LOCAL_POSTGRES_HOST}:${LOCAL_POSTGRES_PORT}/${LOCAL_POSTGRES_DATABASE}?sslmode=${LOCAL_POSTGRES_SSLMODE}}"
OBJECT_STORE_READY_TIMEOUT_SECONDS="${CARTULARY_OBJECT_STORE_READY_TIMEOUT_SECONDS:-120}"
SEAWEEDFS_S3_PORT="${SEAWEEDFS_S3_PORT:-8333}"
SEAWEEDFS_S3_UPSTREAM_PORT="${SEAWEEDFS_S3_UPSTREAM_PORT:-18333}"
OBJECT_STORE_ENDPOINT="${OBJECT_STORE_ENDPOINT:-127.0.0.1:${SEAWEEDFS_S3_PORT}}"
OBJECT_STORE_BUCKET="${OBJECT_STORE_BUCKET:-cartulary}"
SEAWEEDFS_S3_ACCESS_KEY_ID="${SEAWEEDFS_S3_ACCESS_KEY_ID:-cartulary-local}"
SEAWEEDFS_S3_SECRET_ACCESS_KEY="${SEAWEEDFS_S3_SECRET_ACCESS_KEY:-cartulary-local-secret}"
OBJECT_STORE_SECURE="${OBJECT_STORE_SECURE:-false}"
OBJECT_STORE_CORS_ORIGIN="${OBJECT_STORE_CORS_ORIGIN:-http://localhost:5173}"
OBJECT_STORE_CORS_ALLOWED_ORIGINS="${OBJECT_STORE_CORS_ALLOWED_ORIGINS:-$OBJECT_STORE_CORS_ORIGIN}"
OBJECT_STORE_CORS_PROXY_LISTEN="${OBJECT_STORE_CORS_PROXY_LISTEN:-127.0.0.1:${SEAWEEDFS_S3_PORT}}"
OBJECT_STORE_CORS_PROXY_UPSTREAM="${OBJECT_STORE_CORS_PROXY_UPSTREAM:-http://127.0.0.1:${SEAWEEDFS_S3_UPSTREAM_PORT}}"
SEAWEEDFS_S3_IMAGE="${SEAWEEDFS_S3_IMAGE:-docker.io/chrislusf/seaweedfs:4.17}"
SEAWEEDFS_S3_IMAGE_DIGEST="${SEAWEEDFS_S3_IMAGE_DIGEST:-sha256:186de7ef977a20343ee9a5544073f081976a29e2d29ecf8379891e7bf177fbe9}"
GO_BIN="${GO:-go}"
GO_CACHE="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
GO_MOD_CACHE="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
GO_TMP="${GO_TMP_DIR:?GO_TMP_DIR is required}"
RUNTIME_DIR="${CARTULARY_RUNTIME_DIR:-$ROOT_DIR/.cartulary/runtime}"
OBJECT_STORE_CORS_PROXY_PID_FILE="${OBJECT_STORE_CORS_PROXY_PID_FILE:-$RUNTIME_DIR/seaweedfs-s3-cors-proxy.pid}"
OBJECT_STORE_CORS_PROXY_STATE_DIR="${OBJECT_STORE_CORS_PROXY_STATE_DIR:-$RUNTIME_DIR/object-store-proxy}"
OBJECT_STORE_CORS_PROXY_LOCK_FILE="${OBJECT_STORE_CORS_PROXY_LOCK_FILE:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/operation.lock}"
OBJECT_STORE_CORS_PROXY_LOCK_METADATA_FILE="${OBJECT_STORE_CORS_PROXY_LOCK_METADATA_FILE:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/operation.json}"
OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE="${OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/start-attempt.json}"
OBJECT_STORE_CORS_PROXY_LEASE_FILE="${OBJECT_STORE_CORS_PROXY_LEASE_FILE:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/ready-lease.json}"
OBJECT_STORE_CORS_PROXY_LOG_DIR="${OBJECT_STORE_CORS_PROXY_LOG_DIR:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/logs}"
OBJECT_STORE_CORS_PROXY_BIN="${OBJECT_STORE_CORS_PROXY_BIN:-$OBJECT_STORE_CORS_PROXY_STATE_DIR/s3corsproxy}"
OBJECT_STORE_CORS_PROXY_LOG_FILE=""
export OBJECT_STORE_CORS_ALLOWED_ORIGINS
export SEAWEEDFS_S3_UPSTREAM_PORT

usage() {
  echo "usage: dev-services.sh up|services-down|wait-postgres|wait-object-store|wait|init-object-store|db-up|db-migrate|db-reset|object-store-reset" >&2
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
  env GOCACHE="$GO_CACHE" GOMODCACHE="$GO_MOD_CACHE" GOTMPDIR="$GO_TMP" \
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

proxy_listener_in_use() {
  local port="${OBJECT_STORE_CORS_PROXY_LISTEN##*:}"
  command -v ss >/dev/null 2>&1 &&
    ss -ltn "sport = :${port}" 2>/dev/null | tail -n +2 | grep -q .
}

proxy_command() {
  "$OBJECT_STORE_CORS_PROXY_BIN" "$@" \
    --listen "$OBJECT_STORE_CORS_PROXY_LISTEN" \
    --upstream "$OBJECT_STORE_CORS_PROXY_UPSTREAM" \
    --origin "$OBJECT_STORE_CORS_ORIGIN"
}

build_object_store_proxy() {
  cd "$ROOT_DIR"
  env GOCACHE="$GO_CACHE" GOMODCACHE="$GO_MOD_CACHE" GOTMPDIR="$GO_TMP" \
    "$GO_BIN" build -o "$OBJECT_STORE_CORS_PROXY_BIN" ./tools/s3corsproxy
}

write_proxy_lock_metadata() {
  local operation="$1"
  local temporary="${OBJECT_STORE_CORS_PROXY_LOCK_METADATA_FILE}.tmp.$$"
  umask 077
  printf '{"schema_id":"cartulary.local_object_store_proxy_operation.v1","operation":"%s","owner_pid":%d,"started_at":"%s"}\n' \
    "$operation" "$$" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >"$temporary"
  chmod 600 "$temporary"
  sync "$temporary"
  mv -f "$temporary" "$OBJECT_STORE_CORS_PROXY_LOCK_METADATA_FILE"
  sync "$OBJECT_STORE_CORS_PROXY_STATE_DIR"
}

with_proxy_lock() {
  local operation="$1"
  shift
  local lock_status=0
  local lock_fd
  if ! command -v flock >/dev/null 2>&1; then
    echo "local object-store proxy lifecycle requires flock" >&2
    return 1
  fi
  umask 077
  mkdir -p "$OBJECT_STORE_CORS_PROXY_STATE_DIR" "$OBJECT_STORE_CORS_PROXY_LOG_DIR"
  chmod 700 "$OBJECT_STORE_CORS_PROXY_STATE_DIR" "$OBJECT_STORE_CORS_PROXY_LOG_DIR"
  exec {lock_fd}>"$OBJECT_STORE_CORS_PROXY_LOCK_FILE"
  flock -x "$lock_fd"
  write_proxy_lock_metadata "$operation"
  "$@" || lock_status=$?
  rm -f "$OBJECT_STORE_CORS_PROXY_LOCK_METADATA_FILE"
  sync "$OBJECT_STORE_CORS_PROXY_STATE_DIR"
  flock -u "$lock_fd"
  eval "exec ${lock_fd}>&-"
  return "$lock_status"
}

discard_proxy_state() {
  local state_file="$1"
  [[ -f "$state_file" ]] || return 0
  proxy_command discard --state-file "$state_file"
}

resolve_existing_proxy_state() {
  local state_file="$1"
  local state_status=0
  [[ -f "$state_file" ]] || return 1

  if proxy_command status --state-file "$state_file" >/dev/null 2>&1; then
    return 0
  else
    state_status=$?
  fi

  case "$state_status" in
    3)
      proxy_command stop --state-file "$state_file"
      ;;
    4)
      discard_proxy_state "$state_file"
      ;;
    *)
      if proxy_listener_in_use; then
        echo "resource_conflict: unproven listener owns ${OBJECT_STORE_CORS_PROXY_LISTEN}" >&2
        return 2
      fi
      discard_proxy_state "$state_file"
      ;;
  esac
  return 1
}

start_object_store_proxy_locked() {
  local instance_id=""
  local start_time="$SECONDS"
  local child_pid=""
  local proxy_binary_built=0

  rm -f "$OBJECT_STORE_CORS_PROXY_PID_FILE"
  if [[ ! -x "$OBJECT_STORE_CORS_PROXY_BIN" ]]; then
    build_object_store_proxy
    proxy_binary_built=1
  fi
  if resolve_existing_proxy_state "$OBJECT_STORE_CORS_PROXY_LEASE_FILE"; then
    return 0
  elif [[ "$?" -eq 2 ]]; then
    return 1
  fi
  if resolve_existing_proxy_state "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE"; then
    return 0
  elif [[ "$?" -eq 2 ]]; then
    return 1
  fi
  if proxy_listener_in_use; then
    echo "resource_conflict: unproven listener owns ${OBJECT_STORE_CORS_PROXY_LISTEN}" >&2
    return 1
  fi

  if [[ "$proxy_binary_built" -ne 1 ]]; then
    build_object_store_proxy
  fi
  instance_id="$(cat /proc/sys/kernel/random/uuid)"
  instance_id="${instance_id//-/}"
  OBJECT_STORE_CORS_PROXY_LOG_FILE="$OBJECT_STORE_CORS_PROXY_LOG_DIR/${instance_id}.log"
  proxy_command attempt \
    --attempt-file "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" \
    --instance-id "$instance_id" \
    --log-path "$OBJECT_STORE_CORS_PROXY_LOG_FILE"
  # The path is recorded as metadata and independently receives process output.
  # shellcheck disable=SC2094
  nohup "$OBJECT_STORE_CORS_PROXY_BIN" serve \
    --listen "$OBJECT_STORE_CORS_PROXY_LISTEN" \
    --upstream "$OBJECT_STORE_CORS_PROXY_UPSTREAM" \
    --origin "$OBJECT_STORE_CORS_ORIGIN" \
    --attempt-file "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" \
    --instance-id "$instance_id" \
    --log-path "$OBJECT_STORE_CORS_PROXY_LOG_FILE" \
    >"$OBJECT_STORE_CORS_PROXY_LOG_FILE" 2>&1 &
  child_pid="$!"

  while (( SECONDS - start_time < 5 )); do
    if proxy_command status --state-file "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" >/dev/null 2>&1; then
      return 0
    fi
    if ! kill -0 "$child_pid" >/dev/null 2>&1; then
      break
    fi
    sleep 0.1
  done

  echo "resource_conflict: local object-store proxy did not complete its bind/identity handshake" >&2
  if [[ -f "$OBJECT_STORE_CORS_PROXY_LOG_FILE" ]]; then
    cat "$OBJECT_STORE_CORS_PROXY_LOG_FILE" >&2 || true
  fi
  return 1
}

start_object_store_proxy() {
  with_proxy_lock start start_object_store_proxy_locked
}

promote_object_store_proxy_locked() {
  if [[ -f "$OBJECT_STORE_CORS_PROXY_LEASE_FILE" ]]; then
    proxy_command status --state-file "$OBJECT_STORE_CORS_PROXY_LEASE_FILE" >/dev/null
    return $?
  fi
  proxy_command promote \
    --attempt-file "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" \
    --lease-file "$OBJECT_STORE_CORS_PROXY_LEASE_FILE"
}

promote_object_store_proxy() {
  with_proxy_lock promote promote_object_store_proxy_locked
}

stop_object_store_proxy_locked() {
  local status=0
  rm -f "$OBJECT_STORE_CORS_PROXY_PID_FILE"
  if [[ -f "$OBJECT_STORE_CORS_PROXY_LEASE_FILE" ]]; then
    proxy_command stop --state-file "$OBJECT_STORE_CORS_PROXY_LEASE_FILE" || status=$?
  fi
  if [[ -f "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" ]]; then
    if ! proxy_command stop --state-file "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE"; then
      if proxy_listener_in_use; then
        echo "resource_conflict: refusing to signal an unproven proxy listener" >&2
        status=1
      else
        discard_proxy_state "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" || status=$?
      fi
    fi
  fi
  if [[ ! -f "$OBJECT_STORE_CORS_PROXY_LEASE_FILE" &&
        ! -f "$OBJECT_STORE_CORS_PROXY_ATTEMPT_FILE" ]] &&
    proxy_listener_in_use; then
    echo "resource_conflict: listener at ${OBJECT_STORE_CORS_PROXY_LISTEN} is not owned by the local proxy lifecycle" >&2
    status=1
  fi
  return "$status"
}

stop_object_store_proxy() {
  with_proxy_lock stop stop_object_store_proxy_locked
}

wait_object_store() {
  local start_time="$SECONDS"
  local status="unknown"
  local health="unknown"

  start_object_store_proxy
  while (( SECONDS - start_time < OBJECT_STORE_READY_TIMEOUT_SECONDS )); do
    if probe_object_store probe >/dev/null 2>&1; then
      promote_object_store_proxy
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
  if [[ -f "$OBJECT_STORE_CORS_PROXY_LOG_FILE" ]]; then
    cat "$OBJECT_STORE_CORS_PROXY_LOG_FILE" >&2 || true
  fi
  compose logs --no-color --tail 120 seaweedfs-s3 >&2 || true
  return 1
}

init_object_store() {
  compose up -d --remove-orphans seaweedfs-s3
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

  stop_object_store_proxy
  compose down --remove-orphans
}

db_up() {
  services_up
  init_object_store
}

run_local_migrate() {
  local go_bin="${GO:-go}"
  local config_file="${CONFIG_FILE:-$ROOT_DIR/configs/dev/config.toml}"
  local go_cache="${GO_CACHE_DIR:?GO_CACHE_DIR is required}"
  local go_mod_cache="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}"
  local go_tmp="${GO_TMP_DIR:?GO_TMP_DIR is required}"

  cd "$ROOT_DIR"
  env CARTULARY_CONFIG_FILE="$config_file" \
    CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="$POSTGRES_PRIMARY_DSN" \
    GOCACHE="$go_cache" GOMODCACHE="$go_mod_cache" GOTMPDIR="$go_tmp" \
    "$go_bin" run ./cmd/migrate up
}

db_migrate() {
  compose up -d postgres
  wait_postgres
  printf '%s\n' 'db-migrate: applying local database migrations only; object storage is not reset.'
  run_local_migrate
}

db_reset() {
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
  run_local_migrate
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
  db-migrate)
    db_migrate
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
