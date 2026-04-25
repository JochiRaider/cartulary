#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"
source "${ROOT_DIR}/scripts/lib/web-e2e-lifecycle.sh"

COMPOSE_FILE="${ROOT_DIR}/docker-compose.dev.yml"
DEV_SERVICES_SCRIPT="${ROOT_DIR}/scripts/dev-services.sh"
GO_BIN="${GO:-go}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
SERVER_BIN="${CARTULARY_SERVER_BIN:-}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
TEST_SERVICES_BIN="${CARTULARY_TEST_SERVICES_BIN:-}"
USE_REPO_ROOT_RUNTIME_ARTIFACTS_ENV="CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES"

KEEP_RUNTIME_ROOT=0
TARGET_ARTIFACT_DIR=""
RUNTIME_ROOT_BASE=""
SERVER_LOG=""
WEB_LOG=""
STACK_ENV_FILE=""
STACK_JSON_FILE=""
PLAYWRIGHT_STATE_DIR=""
TEST_SERVICES_ENV_FILE=""
TEST_SERVICES_METADATA_FILE=""
E2E_DB=""
E2E_DSN=""
BACKEND_PORT=""
FRONTEND_PORT=""
API_ORIGIN=""
PUBLIC_ORIGIN=""
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
    STACK_ENV_FILE="${TARGET_ARTIFACT_DIR}/stack.env"
    STACK_JSON_FILE="${TARGET_ARTIFACT_DIR}/stack.json"
    rm -rf "${RUNTIME_ROOT_BASE}"
    rm -f "${SERVER_LOG}" "${WEB_LOG}" "${STACK_ENV_FILE}" "${STACK_JSON_FILE}"
    KEEP_RUNTIME_ROOT=1
  else
    RUNTIME_ROOT_BASE="$(mktemp -d /tmp/cartulary-web-e2e-runtime-XXXXXX)"
    SERVER_LOG="/tmp/cartulary-e2e-server-$$.log"
    WEB_LOG="/tmp/cartulary-e2e-web-$$.log"
    STACK_ENV_FILE="${RUNTIME_ROOT_BASE}/stack.env"
    STACK_JSON_FILE="${RUNTIME_ROOT_BASE}/stack.json"
  fi

  PLAYWRIGHT_STATE_DIR="${RUNTIME_ROOT_BASE}/playwright-state"
  E2E_DB="cartulary_web_e2e_$$"
  E2E_DSN="postgres://cartulary:cartulary@localhost:5432/${E2E_DB}?sslmode=disable"
  TEST_SERVICES_ENV_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.env"
  TEST_SERVICES_METADATA_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.json"

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

validate_port_number() {
  local port="$1"
  local name="$2"

  if [[ ! "${port}" =~ ^[0-9]+$ ]] || (( port < 1 || port > 65535 )); then
    echo "${name} port must be an integer from 1 through 65535, got ${port}" >&2
    return 1
  fi
}

allocate_available_port() {
  local outvar="$1"
  local name="$2"
  local configured_port="$3"
  local excluded_port="${4:-}"
  local -n port_ref="$outvar"

  port_ref=""

  if [[ -n "${configured_port}" ]]; then
    validate_port_number "${configured_port}" "${name}" || return $?
    if [[ -n "${excluded_port}" && "${configured_port}" == "${excluded_port}" ]]; then
      echo "${name} port ${configured_port} must differ from the other browser e2e stack port" >&2
      return 1
    fi
    if port_in_use "${configured_port}"; then
      echo "${name} port ${configured_port} is already in use; choose another CARTULARY_WEB_E2E_*_PORT override" >&2
      ss -ltnp "sport = :${configured_port}" >&2 || true
      return 1
    fi
    port_ref="${configured_port}"
    return 0
  fi

  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  local candidate=""
  for _ in $(seq 1 50); do
    candidate="$("${node_bin}" -e 'const net = require("node:net"); const server = net.createServer(); server.on("error", (error) => { console.error(error.message); process.exit(1); }); server.listen(0, "127.0.0.1", () => { console.log(server.address().port); server.close(); });')"
    validate_port_number "${candidate}" "${name}" || return $?
    if [[ -n "${excluded_port}" && "${candidate}" == "${excluded_port}" ]]; then
      continue
    fi
    if ! port_in_use "${candidate}"; then
      port_ref="${candidate}"
      return 0
    fi
  done

  echo "failed to allocate an available ${name} port for browser e2e" >&2
  return 1
}

resolve_owned_stack_ports() {
  allocate_available_port BACKEND_PORT "backend" "${CARTULARY_WEB_E2E_BACKEND_PORT:-}" "" || return $?
  allocate_available_port FRONTEND_PORT "frontend" "${CARTULARY_WEB_E2E_FRONTEND_PORT:-}" "${BACKEND_PORT}" || return $?

  if [[ "${BACKEND_PORT}" == "${FRONTEND_PORT}" ]]; then
    echo "backend and frontend ports must differ for browser e2e" >&2
    return 1
  fi

  API_ORIGIN="http://127.0.0.1:${BACKEND_PORT}"
  PUBLIC_ORIGIN="http://127.0.0.1:${FRONTEND_PORT}"
  export CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}"
  export CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}"
  export CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}"
  export CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}"
}

write_stack_metadata() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  mkdir -p "$(dirname "${STACK_ENV_FILE}")"
  cat >"${STACK_ENV_FILE}" <<EOF
CARTULARY_WEB_E2E_API_ORIGIN=${API_ORIGIN}
CARTULARY_WEB_E2E_PUBLIC_ORIGIN=${PUBLIC_ORIGIN}
CARTULARY_WEB_E2E_BACKEND_PORT=${BACKEND_PORT}
CARTULARY_WEB_E2E_FRONTEND_PORT=${FRONTEND_PORT}
CARTULARY_WEB_E2E_RUNTIME_ROOT=${RUNTIME_ROOT_BASE}
CARTULARY_WEB_E2E_SERVER_LOG=${SERVER_LOG}
CARTULARY_WEB_E2E_WEB_LOG=${WEB_LOG}
EOF

  CARTULARY_WEB_E2E_STACK_JSON_FILE="${STACK_JSON_FILE}" \
  CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
  CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
  CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}" \
  CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}" \
  CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}" \
  CARTULARY_WEB_E2E_SERVER_LOG="${SERVER_LOG}" \
  CARTULARY_WEB_E2E_WEB_LOG="${WEB_LOG}" \
    "${node_bin}" <<'EOF'
const fs = require("node:fs");

const payload = {
  schema_id: "cartulary.web_e2e_stack.v1",
  api_origin: process.env.CARTULARY_WEB_E2E_API_ORIGIN,
  public_origin: process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN,
  backend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_BACKEND_PORT ?? "", 10),
  frontend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_FRONTEND_PORT ?? "", 10),
  runtime_root: process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT,
  server_log: process.env.CARTULARY_WEB_E2E_SERVER_LOG,
  web_log: process.env.CARTULARY_WEB_E2E_WEB_LOG,
};

fs.writeFileSync(process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE, `${JSON.stringify(payload, null, 2)}\n`);
EOF
}

using_test_services_stack() {
  [[ "${CARTULARY_TEST_SERVICES_ACTIVE:-}" == "1" ]]
}

require_test_services_bin() {
  if [[ -z "${TEST_SERVICES_BIN}" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN is required when CARTULARY_TEST_SERVICES_ACTIVE=1" >&2
    return 1
  fi
  if [[ ! -x "${TEST_SERVICES_BIN}" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN ${TEST_SERVICES_BIN} is not executable" >&2
    return 1
  fi
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
  stop_owned_process_group "${VITE_PGID:-}" "${FRONTEND_PORT:-4173}" "frontend"
  stop_owned_process_group "${SERVER_PGID:-}" "${BACKEND_PORT:-8080}" "backend"
  if using_test_services_stack; then
    if [[ -x "${TEST_SERVICES_BIN}" && -f "${TEST_SERVICES_METADATA_FILE}" ]]; then
      "${TEST_SERVICES_BIN}" cleanup-web-e2e --metadata-file "${TEST_SERVICES_METADATA_FILE}" >/dev/null 2>&1 || true
    fi
  else
    docker compose -f "${COMPOSE_FILE}" exec -T postgres \
      psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null 2>&1 || true
  fi
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
  if using_test_services_stack; then
    require_test_services_bin || return $?
    echo "browser e2e using active test-service Postgres and MinIO stack"
    return 0
  fi

  docker compose -f "${COMPOSE_FILE}" up -d postgres minio >/dev/null
  "${DEV_SERVICES_SCRIPT}" wait
}

browser_prepare_database() {
  assert_port_free "${BACKEND_PORT}" "backend"
  assert_port_free "${FRONTEND_PORT}" "frontend"
  cd "${ROOT_DIR}"

  if using_test_services_stack; then
    require_test_services_bin || return $?
    "${TEST_SERVICES_BIN}" prepare-web-e2e --env-file "${TEST_SERVICES_ENV_FILE}" --metadata-file "${TEST_SERVICES_METADATA_FILE}"
    # shellcheck disable=SC1090
    source "${TEST_SERVICES_ENV_FILE}"
    E2E_DSN="${CARTULARY_POSTGRES_DSN:?}"
    return 0
  fi

  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "CREATE DATABASE \"${E2E_DB}\";" >/dev/null

  local -a migrate_command=()
  resolve_runtime_command migrate_command "migration" "${MIGRATE_BIN}" "${ROOT_DIR}/migrate" ./cmd/migrate

  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY__APPLICATION__PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
  CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
    "${migrate_command[@]}" up
}

browser_wait_backend_ready() {
  wait_for_http "${API_ORIGIN}/readyz" "backend"
}

browser_wait_frontend_ready() {
  wait_for_http "${PUBLIC_ORIGIN}" "frontend"
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

  run_phase_command "browser-e2e allocate ports" resolve_owned_stack_ports
  write_stack_metadata

  run_phase_command "browser-e2e startup services" browser_start_services
  run_phase_command "browser-e2e startup database" browser_prepare_database

  local -a server_command=()
  resolve_runtime_command server_command "backend" "${SERVER_BIN}" "${ROOT_DIR}/server" ./cmd/server

  start_process_group SERVER_PGID "${SERVER_LOG}" \
    env \
    CARTULARY_HTTP_ADDR="127.0.0.1:${BACKEND_PORT}" \
    CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
    CARTULARY__APPLICATION__PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
    CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
    CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
    CARTULARY_S3_ENDPOINT="${CARTULARY_S3_ENDPOINT:-}" \
    CARTULARY_S3_ACCESS_KEY_ID="${CARTULARY_S3_ACCESS_KEY_ID:-}" \
    CARTULARY_S3_SECRET_ACCESS_KEY="${CARTULARY_S3_SECRET_ACCESS_KEY:-}" \
    CARTULARY_S3_SECURE="${CARTULARY_S3_SECURE:-}" \
    CARTULARY_S3_BUCKET="${CARTULARY_S3_BUCKET:-}" \
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
    CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
    CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
    "${pnpm_bin}" --dir apps/web dev --host 127.0.0.1 --port "${FRONTEND_PORT}" --strictPort

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
