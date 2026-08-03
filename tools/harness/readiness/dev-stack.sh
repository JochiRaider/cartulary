#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
source "${ROOT_DIR}/tools/harness/readiness/process-lifecycle.sh"

GO_BIN="${GO:-go}"
CONFIG_FILE="${CONFIG_FILE:-${ROOT_DIR}/configs/dev/config.toml}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
DEV_ARTIFACT_DIR="${CARTULARY_DEV_STACK_ARTIFACT_DIR:-${ROOT_DIR}/tmp/dev-stack}"
SERVER_LOG="${CARTULARY_DEV_STACK_SERVER_LOG:-${DEV_ARTIFACT_DIR}/server.log}"
WEB_LOG="${CARTULARY_DEV_STACK_WEB_LOG:-${DEV_ARTIFACT_DIR}/web.log}"
BACKEND_PORT="${CARTULARY_DEV_STACK_BACKEND_PORT:-8080}"
FRONTEND_PORT="${CARTULARY_DEV_STACK_FRONTEND_PORT:-5173}"
BACKEND_READY_URL="${CARTULARY_DEV_STACK_BACKEND_READY_URL:-http://127.0.0.1:${BACKEND_PORT}/readyz}"
FRONTEND_READY_URL="${CARTULARY_DEV_STACK_FRONTEND_READY_URL:-http://127.0.0.1:${FRONTEND_PORT}}"
READY_TIMEOUT_SECONDS="${CARTULARY_DEV_STACK_READY_TIMEOUT_SECONDS:-180}"
POSTGRES_READY_HOST="${CARTULARY_DEV_STACK_POSTGRES_HOST:-127.0.0.1}"
POSTGRES_READY_PORT="${CARTULARY_DEV_STACK_POSTGRES_PORT:-5432}"
LOCAL_POSTGRES_HOST="${CARTULARY_LOCAL_POSTGRES_HOST:-localhost}"
LOCAL_POSTGRES_PORT="${CARTULARY_LOCAL_POSTGRES_PORT:-5432}"
LOCAL_POSTGRES_DATABASE="${CARTULARY_LOCAL_POSTGRES_DATABASE:-cartulary}"
LOCAL_POSTGRES_USER="${CARTULARY_LOCAL_POSTGRES_USER:-cartulary}"
LOCAL_POSTGRES_PASSWORD="${CARTULARY_LOCAL_POSTGRES_PASSWORD:-cartulary}"
LOCAL_POSTGRES_SSLMODE="${CARTULARY_LOCAL_POSTGRES_SSLMODE:-disable}"
POSTGRES_PRIMARY_DSN="${CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN:-postgres://${LOCAL_POSTGRES_USER}:${LOCAL_POSTGRES_PASSWORD}@${LOCAL_POSTGRES_HOST}:${LOCAL_POSTGRES_PORT}/${LOCAL_POSTGRES_DATABASE}?sslmode=${LOCAL_POSTGRES_SSLMODE}}"
OBJECT_STORE_READY_HOST="${CARTULARY_DEV_STACK_OBJECT_STORE_HOST:-127.0.0.1}"
OBJECT_STORE_READY_PORT="${CARTULARY_DEV_STACK_OBJECT_STORE_PORT:-${SEAWEEDFS_S3_UPSTREAM_PORT:-18333}}"
OBJECT_STORE_ENDPOINT="${OBJECT_STORE_ENDPOINT:-localhost:${OBJECT_STORE_READY_PORT}}"
OBJECT_STORE_BUCKET="${OBJECT_STORE_BUCKET:-cartulary}"
SEAWEEDFS_S3_ACCESS_KEY_ID="${SEAWEEDFS_S3_ACCESS_KEY_ID:-cartulary-local}"
SEAWEEDFS_S3_SECRET_ACCESS_KEY="${SEAWEEDFS_S3_SECRET_ACCESS_KEY:-cartulary-local-secret}"
REVISIONS_CONFLICT_TOKEN_SECRET="$(dd if=/dev/urandom bs=32 count=1 status=none | base64 | tr '+/' '-_' | tr -d '=\n')"

SERVER_PGID=""
VITE_PGID=""
cleanup_done=0

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

  echo "${name} port ${port} remained in use after dev stack cleanup" >&2
  ss -ltnp "sport = :${port}" >&2 || true
  return 1
}

assert_dev_port_free() {
  local port="$1"
  local name="$2"

  if ! command -v ss >/dev/null 2>&1; then
    return 0
  fi

  if port_in_use "${port}"; then
    echo "${name} port ${port} is already in use; stop the existing listener before make dev" >&2
    ss -ltnp "sport = :${port}" >&2 || true
    return 1
  fi
}

tcp_port_ready() {
  local host="$1"
  local port="$2"

  (exec 3<>"/dev/tcp/${host}/${port}") >/dev/null 2>&1
}

assert_backing_services_ready() {
  local failed=0

  if [[ "${CARTULARY_DEV_STACK_SKIP_SERVICE_PREFLIGHT:-0}" == "1" ]]; then
    return 0
  fi

  if ! tcp_port_ready "${POSTGRES_READY_HOST}" "${POSTGRES_READY_PORT}"; then
    echo "Postgres is not reachable at ${POSTGRES_READY_HOST}:${POSTGRES_READY_PORT}; run make db-up or make services-up before make dev" >&2
    failed=1
  fi

  if ! tcp_port_ready "${OBJECT_STORE_READY_HOST}" "${OBJECT_STORE_READY_PORT}"; then
    echo "SeaweedFS S3 is not reachable at ${OBJECT_STORE_READY_HOST}:${OBJECT_STORE_READY_PORT}; run make db-up or make services-up before make dev" >&2
    failed=1
  fi

  if [[ "${failed}" -ne 0 ]]; then
    return 1
  fi
}

stop_owned_process_group() {
  local group_id="$1"
  local port="$2"
  local name="$3"

  if [[ -z "${group_id}" ]]; then
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

  stop_owned_process_group "${VITE_PGID:-}" "${FRONTEND_PORT}" "frontend"
  stop_owned_process_group "${SERVER_PGID:-}" "${BACKEND_PORT}" "backend"
}

exit_for_requested_shutdown() {
  local context="$1"

  if ! lifecycle_shutdown_requested; then
    return 0
  fi

  echo "received $(lifecycle_signal_name) during ${context}; shutting down dev stack" >&2
  return "$(lifecycle_signal_exit_status)"
}

wait_for_http() {
  local url="$1"
  local name="$2"
  local start_time="${SECONDS}"
  local shutdown_status=0

  while (( SECONDS - start_time < READY_TIMEOUT_SECONDS )); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi

    if exit_for_requested_shutdown "${name} readiness"; then
      :
    else
      shutdown_status=$?
      return "${shutdown_status}"
    fi

    if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
      echo "backend exited before ${name} readiness; inspect ${SERVER_LOG}, run make db-up for backing services, and run make db-migrate for current-line schema upgrades" >&2
      if [[ "${name}" == "backend" ]]; then
        echo "If the backend log reports prod_ddl_rebaseline_v1/historical_migration_lineage, reset the local database with CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset or use an owner-approved export/import path." >&2
      fi
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
      echo "frontend exited before ${name} readiness; inspect ${WEB_LOG}" >&2
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi

    sleep 1
  done

  echo "timed out waiting for ${name} at ${url}; inspect ${SERVER_LOG} and ${WEB_LOG}, run make db-up for backing services, and run make db-migrate for current-line schema upgrades" >&2
  if [[ "${name}" == "backend" ]]; then
    echo "If the backend log reports prod_ddl_rebaseline_v1/historical_migration_lineage, reset the local database with CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset or use an owner-approved export/import path." >&2
  fi
  cat "${SERVER_LOG}" >&2 || true
  cat "${WEB_LOG}" >&2 || true
  return 1
}

resolve_backend_command() {
  local outvar="$1"
  local -n backend_command_ref="$outvar"

  if [[ -n "${CARTULARY_DEV_STACK_BACKEND_COMMAND:-}" ]]; then
    # shellcheck disable=SC2034
    backend_command_ref=(bash -lc "${CARTULARY_DEV_STACK_BACKEND_COMMAND}")
    return 0
  fi

  # shellcheck disable=SC2034
  backend_command_ref=("${GO_BIN}" run ./cmd/server)
}

resolve_frontend_command() {
  local outvar="$1"
  local -n frontend_command_ref="$outvar"

  if [[ -n "${CARTULARY_DEV_STACK_FRONTEND_COMMAND:-}" ]]; then
    # shellcheck disable=SC2034
    frontend_command_ref=(bash -lc "${CARTULARY_DEV_STACK_FRONTEND_COMMAND}")
    return 0
  fi

  if [[ ! -x "${PNPM_BIN}" ]]; then
    echo "repo-local pnpm was not found at ${PNPM_BIN}; run make frontend-toolchain" >&2
    return 1
  fi

  if [[ "${CARTULARY_DEV_STACK_SKIP_INOTIFY_PREFLIGHT:-0}" != "1" ]]; then
    local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
    if [[ ! -x "${node_bin}" ]]; then
      node_bin="node"
    fi
    "${node_bin}" "${ROOT_DIR}/tools/harness/readiness/diagnose-inotify.mjs" --require-dev-watch-capacity || return $?
  fi

  # shellcheck disable=SC2034
  frontend_command_ref=("${PNPM_BIN}" --dir apps/web dev --host 127.0.0.1 --port "${FRONTEND_PORT}" --strictPort)
}

dev_wait_backend_ready() {
  if [[ "${CARTULARY_DEV_STACK_SKIP_READINESS:-0}" == "1" ]]; then
    return 0
  fi
  wait_for_http "${BACKEND_READY_URL}" "backend"
}

dev_wait_frontend_ready() {
  if [[ "${CARTULARY_DEV_STACK_SKIP_READINESS:-0}" == "1" ]]; then
    return 0
  fi
  wait_for_http "${FRONTEND_READY_URL}" "frontend"
}

supervise_dev_stack() {
  local shutdown_status=0

  while true; do
    if exit_for_requested_shutdown "dev stack supervision"; then
      :
    else
      shutdown_status=$?
      return "${shutdown_status}"
    fi

    if ! process_group_running "${SERVER_PGID}"; then
      # process_group_running reaps an exited group leader. Waiting for the
      # same PID again is both unnecessary and racy when a short-lived child
      # exits while several harness cases are completing concurrently.
      echo "backend exited during dev stack supervision" >&2
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi

    if ! process_group_running "${VITE_PGID}"; then
      echo "frontend exited during dev stack supervision" >&2
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi

    sleep 1
  done
}

main() {
  mkdir -p "${DEV_ARTIFACT_DIR}"
  rm -f "${SERVER_LOG}" "${WEB_LOG}"

  trap cleanup EXIT
  lifecycle_reset_shutdown_state
  lifecycle_install_signal_traps

  assert_backing_services_ready
  assert_dev_port_free "${BACKEND_PORT}" "backend"
  assert_dev_port_free "${FRONTEND_PORT}" "frontend"

  local -a backend_command=()
  local -a frontend_command=()
  resolve_backend_command backend_command
  resolve_frontend_command frontend_command

  cd "${ROOT_DIR}"

  start_process_group SERVER_PGID "${SERVER_LOG}" \
    env \
    CARTULARY_CONFIG_FILE="${CONFIG_FILE}" \
    CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
    CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH="${ROOT_DIR}/configs/dev/revisions-conflict-token-key-ring.json" \
    CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_DEV_ACTIVE="${REVISIONS_CONFLICT_TOKEN_SECRET}" \
    CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="${POSTGRES_PRIMARY_DSN}" \
    CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT="${OBJECT_STORE_ENDPOINT}" \
    CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID="${SEAWEEDFS_S3_ACCESS_KEY_ID}" \
    CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY="${SEAWEEDFS_S3_SECRET_ACCESS_KEY}" \
    CARTULARY_S3_OBJECT_PRIMARY_SECURE="false" \
    CARTULARY_S3_OBJECT_PRIMARY_BUCKET="${OBJECT_STORE_BUCKET}" \
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
    "${backend_command[@]}"

  echo "backend log: ${SERVER_LOG}"

  dev_wait_backend_ready

  start_process_group VITE_PGID "${WEB_LOG}" \
    env \
    COREPACK_HOME="${NODE_RUNTIME_DIR}/corepack" \
    PATH="${NODE_RUNTIME_DIR}/bin:${PATH}" \
    "${frontend_command[@]}"

  echo "frontend log: ${WEB_LOG}"

  dev_wait_frontend_ready

  supervise_dev_stack
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
