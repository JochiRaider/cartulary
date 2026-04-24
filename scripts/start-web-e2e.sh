#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"
source "${ROOT_DIR}/scripts/lib/web-e2e-lifecycle.sh"

COMPOSE_FILE="${ROOT_DIR}/docker-compose.dev.yml"
GO_BIN="${GO:-go}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
SERVER_BIN="${CARTULARY_SERVER_BIN:-}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
USE_REPO_ROOT_RUNTIME_ARTIFACTS_ENV="CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES"
MINIO_READY_URL="http://127.0.0.1:9000/minio/health/ready"
POSTGRES_READY_TIMEOUT_SECONDS="${CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS:-180}"

KEEP_RUNTIME_ROOT=0
TARGET_ARTIFACT_DIR=""
RUNTIME_ROOT_BASE=""
SERVER_LOG=""
WEB_LOG=""
PLAYWRIGHT_STATE_DIR=""
E2E_DB=""
E2E_DSN=""
child_command=()
SERVER_PGID=""
VITE_PGID=""
CHILD_PGID=""
cleanup_done=0

usage() {
  echo "usage: start-web-e2e.sh [-- <command...>]" >&2
}

parse_child_command() {
  child_command=()

  if [[ "$#" -eq 0 ]]; then
    return 0
  fi

  if [[ "$1" != "--" ]]; then
    usage
    return 2
  fi
  shift

  if [[ "$#" -eq 0 ]]; then
    usage
    return 2
  fi

  child_command=("$@")
}

prepare_runtime_root() {
  TARGET_ARTIFACT_DIR="${CARTULARY_WEB_E2E_ARTIFACT_DIR:-}"
  if [[ -z "${TARGET_ARTIFACT_DIR}" && -n "${CARTULARY_TEST_TARGET:-}" ]]; then
    TARGET_ARTIFACT_DIR="$(ensure_target_artifact_dir)/owned-stack"
  fi

  if [[ -n "${TARGET_ARTIFACT_DIR}" ]]; then
    mkdir -p "${TARGET_ARTIFACT_DIR}"
    RUNTIME_ROOT_BASE="${TARGET_ARTIFACT_DIR}/runtime-root"
    SERVER_LOG="${TARGET_ARTIFACT_DIR}/server.log"
    WEB_LOG="${TARGET_ARTIFACT_DIR}/web.log"
    rm -rf "${RUNTIME_ROOT_BASE}"
    rm -f "${SERVER_LOG}" "${WEB_LOG}"
    KEEP_RUNTIME_ROOT=1
  else
    RUNTIME_ROOT_BASE="$(mktemp -d /tmp/cartulary-web-e2e-runtime-XXXXXX)"
    SERVER_LOG="/tmp/cartulary-e2e-server-$$.log"
    WEB_LOG="/tmp/cartulary-e2e-web-$$.log"
  fi

  PLAYWRIGHT_STATE_DIR="${RUNTIME_ROOT_BASE}/playwright-state"
  E2E_DB="cartulary_web_e2e_$$"
  E2E_DSN="postgres://cartulary:cartulary@localhost:5432/${E2E_DB}?sslmode=disable"

  mkdir -p \
    "${RUNTIME_ROOT_BASE}/database-storage" \
    "${RUNTIME_ROOT_BASE}/object-storage" \
    "${PLAYWRIGHT_STATE_DIR}" \
    "${RUNTIME_ROOT_BASE}/backup-storage" \
    "${RUNTIME_ROOT_BASE}/reference-pack-storage" \
    "${RUNTIME_ROOT_BASE}/temporary-work" \
    "${RUNTIME_ROOT_BASE}/export-outputs"

  export CARTULARY_PLAYWRIGHT_STATE_DIR="${PLAYWRIGHT_STATE_DIR}"
  export CARTULARY_WEB_E2E_SERVER_LOG="${SERVER_LOG}"
  export CARTULARY_WEB_E2E_WEB_LOG="${WEB_LOG}"
  export CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}"
}

use_repo_root_runtime_artifacts() {
  [[ "${!USE_REPO_ROOT_RUNTIME_ARTIFACTS_ENV:-0}" == "1" ]]
}

# Browser E2E defaults to the current source tree so repo-root build artifacts
# cannot silently drift from the code under test.
resolve_runtime_command() {
  local outvar="$1"
  local label="$2"
  local configured_path="$3"
  local repo_root_artifact="$4"
  shift 4
  local -n resolved_ref="$outvar"

  resolved_ref=()

  if [[ -n "${configured_path}" ]]; then
    if [[ "${configured_path}" == "${repo_root_artifact}" ]] && ! use_repo_root_runtime_artifacts; then
      configured_path=""
    elif [[ ! -x "${configured_path}" ]]; then
      echo "${label} override ${configured_path} is not executable" >&2
      return 1
    fi
  fi

  if [[ -n "${configured_path}" ]]; then
    resolved_ref=("${configured_path}")
    return 0
  fi

  resolved_ref=("${GO_BIN}" run "$@")
}

port_in_use() {
  local port="$1"

  if ! command -v ss >/dev/null 2>&1; then
    return 1
  fi

  ss -ltn "sport = :${port}" | tail -n +2 | grep -q .
}

wait_for_port_release() {
  local port="$1"
  local name="$2"

  if ! command -v ss >/dev/null 2>&1; then
    return 0
  fi

  for _ in $(seq 1 50); do
    if ! port_in_use "${port}"; then
      return 0
    fi
    sleep 0.2
  done

  echo "${name} port ${port} remained in use after browser e2e cleanup" >&2
  ss -ltnp "sport = :${port}" >&2 || true
  return 1
}

stop_owned_process_group() {
  local group_id="$1"
  local port="$2"
  local name="$3"

  if [[ -z "${group_id}" ]]; then
    wait_for_port_release "${port}" "${name}" || true
    return 0
  fi

  stop_process_group "${group_id}" || true
  wait_for_port_release "${port}" "${name}" || true
}

cleanup() {
  if [[ "${cleanup_done}" -eq 1 ]]; then
    return 0
  fi
  cleanup_done=1

  if [[ -n "${CHILD_PGID:-}" ]]; then
    stop_process_group "${CHILD_PGID}" || true
  fi
  stop_owned_process_group "${VITE_PGID:-}" 4173 "frontend"
  stop_owned_process_group "${SERVER_PGID:-}" 8080 "backend"
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null 2>&1 || true
  if [[ "${KEEP_RUNTIME_ROOT}" -ne 1 ]]; then
    rm -rf "${RUNTIME_ROOT_BASE}"
  fi
}

exit_for_requested_shutdown() {
  local context="$1"

  if ! lifecycle_shutdown_requested; then
    return 0
  fi

  echo "received $(lifecycle_signal_name) during ${context}; shutting down browser e2e stack" >&2
  return "$(lifecycle_signal_exit_status)"
}

wait_for_http() {
  local url="$1"
  local name="$2"

  for _ in $(seq 1 180); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    if exit_for_requested_shutdown "${name} readiness"; then
      :
    else
      return "$?"
    fi
    if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
      echo "backend exited before ${name} readiness" >&2
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
      echo "frontend exited before ${name} readiness" >&2
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi
    sleep 1
  done

  echo "timed out waiting for ${name} at ${url}" >&2
  cat "${SERVER_LOG}" >&2 || true
  cat "${WEB_LOG}" >&2 || true
  return 1
}

wait_for_postgres() {
  local start_time="$SECONDS"
  local container_id=""
  local state="unknown"
  local health="unknown"
  local shutdown_status=0

  while (( SECONDS - start_time < POSTGRES_READY_TIMEOUT_SECONDS )); do
    if docker compose -f "${COMPOSE_FILE}" exec -T postgres pg_isready -U cartulary -d postgres >/dev/null 2>&1; then
      return 0
    fi

    if exit_for_requested_shutdown "postgres readiness"; then
      :
    else
      shutdown_status=$?
      return "${shutdown_status}"
    fi

    container_id="$(docker compose -f "${COMPOSE_FILE}" ps -q postgres 2>/dev/null || true)"
    if [[ -n "${container_id}" ]]; then
      state="$(docker inspect -f '{{.State.Status}}' "${container_id}" 2>/dev/null || printf 'unknown')"
      health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "${container_id}" 2>/dev/null || printf 'unknown')"
      if [[ "${state}" == "exited" || "${state}" == "dead" ]]; then
        echo "postgres container is ${state} during browser e2e startup (health=${health})" >&2
        docker compose -f "${COMPOSE_FILE}" logs --no-color --tail 120 postgres >&2 || true
        return 1
      fi
    fi

    sleep 1
  done

  echo "postgres did not become ready for browser e2e after ${POSTGRES_READY_TIMEOUT_SECONDS}s (state=${state} health=${health})" >&2
  docker compose -f "${COMPOSE_FILE}" logs --no-color --tail 120 postgres >&2 || true
  return 1
}

wait_for_minio() {
  local shutdown_status=0

  for _ in $(seq 1 120); do
    if curl -fsS "${MINIO_READY_URL}" >/dev/null 2>&1; then
      return 0
    fi

    if exit_for_requested_shutdown "minio readiness"; then
      :
    else
      shutdown_status=$?
      return "${shutdown_status}"
    fi
    sleep 1
  done

  echo "minio did not become ready for browser e2e" >&2
  docker compose -f "${COMPOSE_FILE}" logs --no-color minio >&2 || true
  return 1
}

assert_port_free() {
  local port="$1"
  local name="$2"

  if ! command -v ss >/dev/null 2>&1; then
    return 0
  fi

  if ss -ltn "sport = :${port}" | tail -n +2 | grep -q .; then
    echo "${name} port ${port} is already in use; stop the existing listener before browser e2e" >&2
    ss -ltnp "sport = :${port}" >&2 || true
    return 1
  fi
}

browser_start_services() {
  docker compose -f "${COMPOSE_FILE}" up -d postgres minio >/dev/null
  wait_for_postgres
  wait_for_minio
}

browser_prepare_database() {
  assert_port_free 8080 "backend"
  assert_port_free 4173 "frontend"
  cd "${ROOT_DIR}"

  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "CREATE DATABASE \"${E2E_DB}\";" >/dev/null

  local -a migrate_command=()
  resolve_runtime_command migrate_command "migration" "${MIGRATE_BIN}" "${ROOT_DIR}/migrate" ./cmd/migrate

  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY__APPLICATION__PUBLIC_ORIGIN="http://127.0.0.1:4173" \
  CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
    "${migrate_command[@]}" up
}

browser_wait_backend_ready() {
  wait_for_http "http://127.0.0.1:8080/readyz" "backend"
}

browser_wait_frontend_ready() {
  wait_for_http "http://127.0.0.1:4173" "frontend"
}

wait_for_process_status() {
  local group_id="$1"
  local status=0

  if wait "${group_id}"; then
    status=0
  else
    status=$?
  fi

  printf '%s\n' "${status}"
}

supervise_stack() {
  local child_status=0
  local shutdown_status=0
  local server_status=0
  local vite_status=0

  while true; do
    if exit_for_requested_shutdown "browser e2e supervision"; then
      :
    else
      shutdown_status=$?
      return "${shutdown_status}"
    fi

    if ! process_group_running "${SERVER_PGID}"; then
      server_status="$(wait_for_process_status "${SERVER_PGID}")"
      echo "backend exited unexpectedly during browser e2e supervision (status=${server_status})" >&2
      cat "${SERVER_LOG}" >&2 || true
      if [[ -n "${CHILD_PGID:-}" ]]; then
        stop_process_group "${CHILD_PGID}" || true
      fi
      return 1
    fi

    if ! process_group_running "${VITE_PGID}"; then
      vite_status="$(wait_for_process_status "${VITE_PGID}")"
      echo "frontend exited unexpectedly during browser e2e supervision (status=${vite_status})" >&2
      cat "${WEB_LOG}" >&2 || true
      if [[ -n "${CHILD_PGID:-}" ]]; then
        stop_process_group "${CHILD_PGID}" || true
      fi
      return 1
    fi

    if [[ -n "${CHILD_PGID:-}" ]] && ! process_group_running "${CHILD_PGID}"; then
      child_status="$(wait_for_process_status "${CHILD_PGID}")"
      return "${child_status}"
    fi

    sleep 1
  done
}

main() {
  parse_child_command "$@"
  prepare_runtime_root

  trap cleanup EXIT
  lifecycle_reset_shutdown_state
  lifecycle_install_signal_traps

  env MAKEFLAGS= CARTULARY_FRONTEND_TOOLCHAIN_QUIET=1 make -s -C "${ROOT_DIR}" --no-print-directory frontend-toolchain
  local pnpm_bin="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
  if [[ ! -x "${pnpm_bin}" ]]; then
    echo "repo-local pnpm was not found at ${pnpm_bin}; run make frontend-toolchain" >&2
    return 1
  fi

  run_phase_command "browser-e2e startup services" browser_start_services
  run_phase_command "browser-e2e startup database" browser_prepare_database

  local -a server_command=()
  resolve_runtime_command server_command "backend" "${SERVER_BIN}" "${ROOT_DIR}/server" ./cmd/server

  start_process_group SERVER_PGID "${SERVER_LOG}" \
    env \
    CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
    CARTULARY__APPLICATION__PUBLIC_ORIGIN="http://127.0.0.1:4173" \
    CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
    CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
    CARTULARY_ENABLE_TEST_ROUTES=1 \
    CARTULARY__ROOTS__DATABASE_STORAGE__PATH="${RUNTIME_ROOT_BASE}/database-storage" \
    CARTULARY__ROOTS__OBJECT_STORAGE__PATH="${RUNTIME_ROOT_BASE}/object-storage" \
    CARTULARY__ROOTS__BACKUP_STORAGE__PATH="${RUNTIME_ROOT_BASE}/backup-storage" \
    CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH="${RUNTIME_ROOT_BASE}/reference-pack-storage" \
    CARTULARY__ROOTS__TEMPORARY_WORK__PATH="${RUNTIME_ROOT_BASE}/temporary-work" \
    CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH="${RUNTIME_ROOT_BASE}/export-outputs" \
    GOCACHE=/tmp/cartulary-go-build \
    GOMODCACHE=/tmp/cartulary-go-mod \
    "${server_command[@]}"

  start_process_group VITE_PGID "${WEB_LOG}" \
    env \
    COREPACK_HOME="${NODE_RUNTIME_DIR}/corepack" \
    PATH="${NODE_RUNTIME_DIR}/bin:${PATH}" \
    "${pnpm_bin}" --dir apps/web dev --host 127.0.0.1 --port 4173 --strictPort

  run_phase_command "browser-e2e startup backend ready" browser_wait_backend_ready
  run_phase_command "browser-e2e startup frontend ready" browser_wait_frontend_ready

  if [[ "${#child_command[@]}" -gt 0 ]]; then
    start_process_group CHILD_PGID "" "${child_command[@]}"
  fi

  supervise_stack
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
