#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="${CARTULARY_COMPOSE_FILE:-$ROOT_DIR/docker-compose.dev.yml}"
POSTGRES_READY_TIMEOUT_SECONDS="${CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS:-180}"
MINIO_READY_TIMEOUT_SECONDS="${CARTULARY_MINIO_READY_TIMEOUT_SECONDS:-120}"
MINIO_BUCKET="${MINIO_BUCKET:-cartulary}"

usage() {
  echo "usage: dev-services.sh wait-postgres|wait-minio|wait|init-minio" >&2
}

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
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

case "${1:-}" in
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
  *)
    usage
    exit 2
    ;;
esac
