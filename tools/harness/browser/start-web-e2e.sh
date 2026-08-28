#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
# shellcheck source=tools/harness/browser/browser-lifecycle-adapter.sh
source "${ROOT_DIR}/tools/harness/browser/browser-lifecycle-adapter.sh"

GO_BIN="${GO:-go}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
SERVER_HARNESS_BIN="${CARTULARY_SERVER_HARNESS_BIN:-}"
TEST_SERVICES_BIN="${CARTULARY_TEST_SERVICES_BIN:-}"
TEST_SERVICE_FRONTEND_PORT_START=19000
TEST_SERVICE_FRONTEND_PORT_END=19199
TEST_SERVICE_FRONTEND_STAGE_WIDTH=100
WEB_DIST_INDEX="${ROOT_DIR}/apps/web/dist/index.html"
SESSION_EVIDENCE_HELPER="${ROOT_DIR}/tools/harness/browser/browser-session-evidence.mjs"
FRONTEND_MODE="preview"
FRONTEND_COMMAND_KIND="vite-preview"

KEEP_RUNTIME_ROOT=0
TARGET_ARTIFACT_DIR=""
RUNTIME_ROOT_BASE=""
SERVER_LOG=""
WEB_LOG=""
STACK_ENV_FILE=""
STACK_JSON_FILE=""
STARTUP_DIAGNOSTIC_FILE=""
STACK_LEASE_FILE=""
SERVICE_ADMISSION_FILE=""
RUN_ROOT=""
PRIVATE_SESSION_ROOT="${CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT:-}"
SUITE_ID="${CARTULARY_TEST_SUITE_ID:-}"
BROWSER_SESSION_ID="${CARTULARY_BROWSER_SESSION_GROUP:-}"
PLAYWRIGHT_STATE_DIR=""
TEST_ROUTE_TOKEN=""
TEST_ROUTE_TOKEN_FILE=""
BACKEND_READY_AT=""
FRONTEND_READY_AT=""
BACKEND_IDENTITY_SERVER_PID=""
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
PORT_LEASE_DIRS=()
cleanup_done=0
SESSION_MODE="wrap"
SESSION_ENV_FILE=""
SESSION_LEASE_FILE=""
RUNTIME_PROFILE_ID="${CARTULARY_BROWSER_RUNTIME_PROFILE_ID:-default}"
RUNTIME_PROFILE_KIND=""
RUNTIME_PROFILE_FINGERPRINT=""
RUNTIME_PROFILE_KEY_RING_MANIFEST=""
RUNTIME_PROFILE_CURSOR_SECRET=""
RUNTIME_PROFILE_SAFE_DIGEST_SECRET=""
REVISIONS_CONFLICT_TOKEN_SECRET=""
BACKEND_GENERATION_HEAD=""
BACKEND_RESTART_SECRET_FILE=""
EXPECTED_RUNTIME_PROFILE_FINGERPRINT=""
RESET_LABEL=""
RESET_DATABASE_DIAGNOSTIC_FILE=""
RESET_OBJECT_STORE_MARKER_FILE=""
RESET_STATE_MARKER_FILE=""
RESET_BACKEND_READY_MARKER_FILE=""

usage() {
  echo "usage: start-web-e2e.sh [-- <command...>]" >&2
  echo "       start-web-e2e.sh --session-start --env-file <path> --lease-file <path>" >&2
  echo "       start-web-e2e.sh --session-stop --lease-file <path>" >&2
  echo "       start-web-e2e.sh --session-reset-backend --lease-file <path> --label <label> --database-result-file <path> --object-store-marker-file <path> --state-marker-file <path> --backend-ready-marker-file <path>" >&2
}

parse_child_command() {
  child_command=()

  if [[ "$#" -eq 0 ]]; then
    return 0
  fi

  case "$1" in
    --session-start)
      SESSION_MODE="start"
      shift
      while [[ "$#" -gt 0 ]]; do
        case "$1" in
          --env-file)
            SESSION_ENV_FILE="${2:-}"
            shift 2
            ;;
          --lease-file)
            SESSION_LEASE_FILE="${2:-}"
            shift 2
            ;;
          *)
            usage
            return 2
            ;;
        esac
      done
      if [[ -z "${SESSION_ENV_FILE}" || -z "${SESSION_LEASE_FILE}" ]]; then
        usage
        return 2
      fi
      return 0
      ;;
    --session-stop)
      SESSION_MODE="stop"
      shift
      while [[ "$#" -gt 0 ]]; do
        case "$1" in
          --lease-file)
            SESSION_LEASE_FILE="${2:-}"
            shift 2
            ;;
          *)
            usage
            return 2
            ;;
        esac
      done
      if [[ -z "${SESSION_LEASE_FILE}" ]]; then
        usage
        return 2
      fi
      return 0
      ;;
    --session-reset-backend)
      SESSION_MODE="reset"
      shift
      while [[ "$#" -gt 0 ]]; do
        case "$1" in
          --lease-file)
            SESSION_LEASE_FILE="${2:-}"
            shift 2
            ;;
          --label)
            RESET_LABEL="${2:-}"
            shift 2
            ;;
          --database-result-file)
            RESET_DATABASE_DIAGNOSTIC_FILE="${2:-}"
            shift 2
            ;;
          --object-store-marker-file)
            RESET_OBJECT_STORE_MARKER_FILE="${2:-}"
            shift 2
            ;;
          --state-marker-file)
            RESET_STATE_MARKER_FILE="${2:-}"
            shift 2
            ;;
          --backend-ready-marker-file)
            RESET_BACKEND_READY_MARKER_FILE="${2:-}"
            shift 2
            ;;
          *)
            usage
            return 2
            ;;
        esac
      done
      if [[ -z "${SESSION_LEASE_FILE}" || ! "${RESET_LABEL}" =~ ^[A-Za-z0-9_.-]+$ || -z "${RESET_DATABASE_DIAGNOSTIC_FILE}" || -z "${RESET_OBJECT_STORE_MARKER_FILE}" || -z "${RESET_STATE_MARKER_FILE}" || -z "${RESET_BACKEND_READY_MARKER_FILE}" ]]; then
        usage
        return 2
      fi
      return 0
      ;;
  esac

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
  local results_root="${CARTULARY_TEST_RESULTS_DIR:-}"
  local run_id="${CARTULARY_TEST_RUN_ID:-}"
  local suite_runtime_root="${CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:-}"
  local suite_runtime_real=""
  local private_session_real=""
  local suite_runtime_lease_id="${CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:-}"
  local suite_runtime_run_id="${CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:-}"
  local owner_marker=""

  if [[ "${CARTULARY_BROWSER_SERVICE_REQUIREMENT:-}" != "test-services" ]]; then
    echo "browser lifecycle adapter requires CARTULARY_BROWSER_SERVICE_REQUIREMENT=test-services" >&2
    return 2
  fi
  if [[ ! "${SUITE_ID}" =~ ^[A-Za-z0-9_.-]+$ ]]; then
    echo "managed browser session requires a safe CARTULARY_TEST_SUITE_ID" >&2
    return 2
  fi
  if [[ ! "${BROWSER_SESSION_ID}" =~ ^[A-Za-z0-9_.-]+$ ]]; then
    echo "managed browser session requires a safe CARTULARY_BROWSER_SESSION_GROUP" >&2
    return 2
  fi
  if [[ -z "${results_root}" || ! "${run_id}" =~ ^[A-Za-z0-9_.-]+$ ]]; then
    echo "managed browser session requires CARTULARY_TEST_RESULTS_DIR and a safe CARTULARY_TEST_RUN_ID" >&2
    return 2
  fi
  if [[ "${results_root}" = /* ]]; then
    RUN_ROOT="${results_root}/${run_id}"
  else
    RUN_ROOT="${ROOT_DIR}/${results_root}/${run_id}"
  fi
  if [[ -z "${suite_runtime_root}" || -z "${PRIVATE_SESSION_ROOT}" || -z "${suite_runtime_lease_id}" || -z "${suite_runtime_run_id}" ]]; then
    echo "managed browser session requires an external suite runtime root and private session root" >&2
    return 2
  fi
  if [[ ! -d "${suite_runtime_root}" || ! -d "${PRIVATE_SESSION_ROOT}" || -L "${suite_runtime_root}" || -L "${PRIVATE_SESSION_ROOT}" ]]; then
    echo "browser private runtime roots must be existing non-symlink directories" >&2
    return 2
  fi
  suite_runtime_real="$(realpath "${suite_runtime_root}")"
  private_session_real="$(realpath "${PRIVATE_SESSION_ROOT}")"
  case "${private_session_real}/" in
    "${suite_runtime_real}/"*) ;;
    *)
      echo "browser private session root must be beneath the suite runtime root" >&2
      return 2
      ;;
  esac
  case "${suite_runtime_real}/" in
    "${ROOT_DIR}/"*|"${RUN_ROOT}/"*)
      echo "suite runtime root must be outside repository and retained result roots" >&2
      return 2
      ;;
  esac
  if [[ "$(stat -c '%a' "${suite_runtime_real}")" != "700" || "$(stat -c '%a' "${private_session_real}")" != "700" ]]; then
    echo "browser private runtime roots must be owner-only 0700 directories" >&2
    return 2
  fi
  if [[ "$(stat -c '%u' "${suite_runtime_real}")" != "$(id -u)" || "$(stat -c '%u' "${private_session_real}")" != "$(id -u)" ]]; then
    echo "browser private runtime roots must be owned by the current user" >&2
    return 2
  fi
  owner_marker="${suite_runtime_real}/runtime-owner.json"
  if [[ ! -f "${owner_marker}" || -L "${owner_marker}" || "$(stat -c '%a' "${owner_marker}")" != "600" || "$(stat -c '%u' "${owner_marker}")" != "$(id -u)" ]]; then
    echo "suite runtime owner marker must be an owner-only 0600 regular file" >&2
    return 2
  fi
  if ! grep -Fq "\"lease_id\":\"${suite_runtime_lease_id}\"" "${owner_marker}" || ! grep -Fq "\"run_id\":\"${suite_runtime_run_id}\"" "${owner_marker}"; then
    echo "suite runtime owner marker does not match the active browser lease" >&2
    return 2
  fi
  PRIVATE_SESSION_ROOT="${private_session_real}"
  TARGET_ARTIFACT_DIR="${RUN_ROOT}/_shared/test-services/${SUITE_ID}/browser-sessions/${BROWSER_SESSION_ID}"
  if [[ -e "${TARGET_ARTIFACT_DIR}" || -L "${TARGET_ARTIFACT_DIR}" ]]; then
    echo "browser session artifact identity ${BROWSER_SESSION_ID} is already in use" >&2
    return 2
  fi
  step_secure_mkdir "${TARGET_ARTIFACT_DIR}" "${PRIVATE_SESSION_ROOT}/logs"
  RUNTIME_ROOT_BASE="${PRIVATE_SESSION_ROOT}/runtime-root"
  SERVER_LOG="${PRIVATE_SESSION_ROOT}/logs/server.log"
  WEB_LOG="${PRIVATE_SESSION_ROOT}/logs/web.log"
  STACK_ENV_FILE="${PRIVATE_SESSION_ROOT}/stack.env"
  STACK_JSON_FILE="${TARGET_ARTIFACT_DIR}/stack-v6.json"
  STARTUP_DIAGNOSTIC_FILE="${TARGET_ARTIFACT_DIR}/startup-diagnostics.json"
  STACK_LEASE_FILE="${TARGET_ARTIFACT_DIR}/browser-stack-lease.json"
  SERVICE_ADMISSION_FILE="${TARGET_ARTIFACT_DIR}/service-admission.json"
  rm -rf "${RUNTIME_ROOT_BASE}"
  rm -f "${STACK_ENV_FILE}"
  KEEP_RUNTIME_ROOT=1

  PLAYWRIGHT_STATE_DIR="${RUNTIME_ROOT_BASE}/playwright-state"
  E2E_DB="cartulary_web_e2e_$$"
  E2E_DSN="postgres://cartulary:cartulary@localhost:5432/${E2E_DB}?sslmode=disable"
  TEST_SERVICES_ENV_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.env"
  TEST_SERVICES_METADATA_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.json"
  TEST_ROUTE_TOKEN_FILE="${RUNTIME_ROOT_BASE}/test-route-token"
  BACKEND_GENERATION_HEAD="${RUNTIME_ROOT_BASE}/backend-generation-head.json"
  BACKEND_RESTART_SECRET_FILE="${RUNTIME_ROOT_BASE}/backend-restart-secrets.json"

  step_secure_mkdir \
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
  export CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS="${STARTUP_DIAGNOSTIC_FILE}"
  export CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR="${TARGET_ARTIFACT_DIR}"
  export CARTULARY_WEB_E2E_FRONTEND_MODE="${FRONTEND_MODE}"
  export CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND="${FRONTEND_COMMAND_KIND}"
  export CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}"
  export CARTULARY_TEST_ROUTE_TOKEN_FILE="${TEST_ROUTE_TOKEN_FILE}"
  export CARTULARY_WEB_E2E_DB="${E2E_DB}"
}

prepare_runtime_profile() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  local profile_row=""
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  profile_row="$(
    "${node_bin}" "${ROOT_DIR}/tools/harness/browser/browser-runtime-profile.mjs" \
      resolve "${ROOT_DIR}/tools/execution_topology_manifest.json" "${RUNTIME_PROFILE_ID}"
  )"
  IFS=$'\t' read -r RUNTIME_PROFILE_ID RUNTIME_PROFILE_KIND RUNTIME_PROFILE_KEY_RING_MANIFEST RUNTIME_PROFILE_FINGERPRINT <<<"${profile_row}"
  if [[ "${RUNTIME_PROFILE_KEY_RING_MANIFEST}" == "-" ]]; then
    RUNTIME_PROFILE_KEY_RING_MANIFEST=""
  fi
  if [[ "${RUNTIME_PROFILE_KIND}" == "network_flow_claimed" ]]; then
    RUNTIME_PROFILE_CURSOR_SECRET="$(dd if=/dev/urandom bs=32 count=1 status=none | base64 | tr '+/' '-_' | tr -d '=\n')"
    RUNTIME_PROFILE_SAFE_DIGEST_SECRET="$(dd if=/dev/urandom bs=32 count=1 status=none | base64 | tr '+/' '-_' | tr -d '=\n')"
  fi
  export CARTULARY_BROWSER_RUNTIME_PROFILE_ID="${RUNTIME_PROFILE_ID}"
  export CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID="${RUNTIME_PROFILE_ID}"
  export CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT="${RUNTIME_PROFILE_FINGERPRINT}"
}

write_backend_restart_secrets() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  CARTULARY_BACKEND_RESTART_SECRET_FILE="${BACKEND_RESTART_SECRET_FILE}" \
  CARTULARY_BACKEND_RESTART_DSN="${E2E_DSN}" \
  CARTULARY_BACKEND_RESTART_REVISIONS_SECRET="${REVISIONS_CONFLICT_TOKEN_SECRET}" \
  CARTULARY_BACKEND_RESTART_CURSOR_SECRET="${RUNTIME_PROFILE_CURSOR_SECRET}" \
  CARTULARY_BACKEND_RESTART_SAFE_DIGEST_SECRET="${RUNTIME_PROFILE_SAFE_DIGEST_SECRET}" \
  CARTULARY_BACKEND_RESTART_S3_ACCESS_KEY_ID="${CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID}" \
  CARTULARY_BACKEND_RESTART_S3_SECRET_ACCESS_KEY="${CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY}" \
    "${node_bin}" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const destination = process.env.CARTULARY_BACKEND_RESTART_SECRET_FILE;
const temporary = `${destination}.tmp-${process.pid}`;
const payload = {
  dsn: process.env.CARTULARY_BACKEND_RESTART_DSN,
  revisions_secret: process.env.CARTULARY_BACKEND_RESTART_REVISIONS_SECRET,
  cursor_secret: process.env.CARTULARY_BACKEND_RESTART_CURSOR_SECRET,
  safe_digest_secret: process.env.CARTULARY_BACKEND_RESTART_SAFE_DIGEST_SECRET,
  s3_access_key_id: process.env.CARTULARY_BACKEND_RESTART_S3_ACCESS_KEY_ID,
  s3_secret_access_key: process.env.CARTULARY_BACKEND_RESTART_S3_SECRET_ACCESS_KEY,
};
fs.mkdirSync(path.dirname(destination), { recursive: true, mode: 0o700 });
fs.writeFileSync(temporary, `${JSON.stringify(payload)}\n`, { mode: 0o600 });
fs.renameSync(temporary, destination);
fs.chmodSync(destination, 0o600);
EOF
}

load_backend_restart_secrets() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  eval "$("${node_bin}" - "${BACKEND_RESTART_SECRET_FILE}" <<'EOF'
const fs = require("node:fs");
const value = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const q = (input) => JSON.stringify(String(input ?? ""));
console.log(`E2E_DSN=${q(value.dsn)}`);
console.log(`REVISIONS_CONFLICT_TOKEN_SECRET=${q(value.revisions_secret)}`);
console.log(`RUNTIME_PROFILE_CURSOR_SECRET=${q(value.cursor_secret)}`);
console.log(`RUNTIME_PROFILE_SAFE_DIGEST_SECRET=${q(value.safe_digest_secret)}`);
console.log(`CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID=${q(value.s3_access_key_id)}`);
console.log(`CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY=${q(value.s3_secret_access_key)}`);
EOF
  )"
}

write_stack_metadata() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  if [[ -z "${BACKEND_READY_AT}" || -z "${FRONTEND_READY_AT}" ]]; then
    echo "v6 browser stack publication requires bound backend and frontend readiness" >&2
    return 1
  fi
  if [[ ! -f "${STARTUP_DIAGNOSTIC_FILE}" ]]; then
    echo "v6 browser stack publication requires terminal startup diagnostics" >&2
    return 1
  fi
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  export CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}"
  export CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}"
  export CARTULARY_WEB_E2E_SERVER_PGID="${SERVER_PGID}"
  export CARTULARY_WEB_E2E_VITE_PGID="${VITE_PGID}"
  export CARTULARY_WEB_E2E_BACKEND_READY_AT="${BACKEND_READY_AT}"
  export CARTULARY_WEB_E2E_FRONTEND_READY_AT="${FRONTEND_READY_AT}"
  export CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID="${BACKEND_IDENTITY_SERVER_PID}"
  export CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE="${TEST_SERVICES_METADATA_FILE}"
  export CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD="${BACKEND_GENERATION_HEAD}"
  export CARTULARY_WEB_E2E_BACKEND_RESTART_SECRET_FILE="${BACKEND_RESTART_SECRET_FILE}"
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" lease || return $?
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" stack >/dev/null || return $?
  export CARTULARY_WEB_E2E_STACK_JSON_FILE="${STACK_JSON_FILE}"

  step_secure_mkdir "$(dirname "${STACK_ENV_FILE}")"
  cat >"${STACK_ENV_FILE}" <<EOF
CARTULARY_WEB_E2E_API_ORIGIN=${API_ORIGIN}
CARTULARY_WEB_E2E_PUBLIC_ORIGIN=${PUBLIC_ORIGIN}
CARTULARY_WEB_E2E_BACKEND_PORT=${BACKEND_PORT}
CARTULARY_WEB_E2E_FRONTEND_PORT=${FRONTEND_PORT}
CARTULARY_WEB_E2E_RUNTIME_ROOT=${RUNTIME_ROOT_BASE}
CARTULARY_WEB_E2E_SERVER_LOG=${SERVER_LOG}
CARTULARY_WEB_E2E_WEB_LOG=${WEB_LOG}
CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS=${STARTUP_DIAGNOSTIC_FILE}
CARTULARY_WEB_E2E_STACK_JSON_FILE=${STACK_JSON_FILE}
CARTULARY_WEB_E2E_FRONTEND_MODE=${FRONTEND_MODE}
CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND=${FRONTEND_COMMAND_KIND}
CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID=${RUNTIME_PROFILE_ID}
CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT=${RUNTIME_PROFILE_FINGERPRINT}
CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD=${BACKEND_GENERATION_HEAD}
CARTULARY_WEB_E2E_BACKEND_RESTART_SECRET_FILE=${BACKEND_RESTART_SECRET_FILE}
EOF
  chmod 600 "${STACK_ENV_FILE}" 2>/dev/null || true
  verify_stack_publication
}

publish_stack_metadata() {
  run_timing_span "setup" "browser-e2e publish immutable v6 stack" write_stack_metadata || return $?
}

verify_stack_publication() {
  local artifact=""
  local label=""

  for artifact in "${SERVICE_ADMISSION_FILE}" "${STACK_LEASE_FILE}" "${STACK_JSON_FILE}"; do
    label="$(basename "${artifact}")"
    if [[ ! -f "${artifact}" || -L "${artifact}" ]]; then
      echo "browser session publication requires regular ${label}" >&2
      return 1
    fi
    if [[ "$(stat -c '%a' "${artifact}")" != "600" || "$(stat -c '%u' "${artifact}")" != "$(id -u)" ]]; then
      echo "browser session publication requires owner-only ${label}" >&2
      return 1
    fi
  done
  return 0
}

write_startup_diagnostics() {
  if [[ "${SESSION_MODE}" == "reset" ]]; then
    return 0
  fi
  local status="$1"
  local step="$2"
  local failure_class="${3:-}"
  local failure_reason="${4:-}"
  local message="${5:-}"
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  if [[ "${status}" == "fail" ]]; then
    "${node_bin}" "${SESSION_EVIDENCE_HELPER}" terminal \
      failed "${message:-browser session startup failed during ${step}}" \
      "${failure_class:-infra}" "${failure_reason:-service_start_error}" || true
    return 0
  fi
  if [[ ! -f "${STARTUP_DIAGNOSTIC_FILE}" ]]; then
    local state="${step}"
    case "${step}" in
      frontend_artifact) state="initializing" ;;
      backend_readiness) state="backend_ready" ;;
      frontend_readiness) state="frontend_ready" ;;
    esac
    "${node_bin}" "${SESSION_EVIDENCE_HELPER}" event \
      "${state}" "${message:-browser session completed ${step}}" || true
  fi
  return 0
}

record_startup_event() {
  local state="$1"
  local message="$2"
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" event "${state}" "${message}"
}

finalize_startup_ready() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" terminal \
    ready "browser session ${BROWSER_SESSION_ID} is ready"
}

snapshot_service_scope() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" write-service-admission
}

require_frontend_preview_artifacts() {
  if [[ -f "${WEB_DIST_INDEX}" ]]; then
    return 0
  fi

  local message="built frontend artifact missing at ${WEB_DIST_INDEX}; run make build-web before browser e2e"
  echo "${message}" >&2
  write_startup_diagnostics "fail" "frontend_artifact" "config" "configuration_error" "${message}" || true
  return 2
}

browser_prepare_frontend_toolchain() {
  env \
    -u CARTULARY_HARNESS_IDENTITY_PREPARED \
    -u CARTULARY_TEST_RUN_ID \
    -u CARTULARY_TEST_TARGET \
    -u CARTULARY_MAKE_INPUT_SOURCES \
    -u CARTULARY_HARNESS_CACHE_MODE \
    -u CARTULARY_HARNESS_CAPACITY_OVERRIDE \
    -u JSON \
    -u OWNER \
    -u ROWS \
    -u SERVICE_BACKED_ONLY \
    -u PLAYWRIGHT_WORKERS \
    -u VITEST_MAX_WORKERS \
    MAKEFLAGS= \
    CARTULARY_FRONTEND_TOOLCHAIN_QUIET=1 \
    CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
    make -s -C "${ROOT_DIR}" --no-print-directory frontend-toolchain
}

bounded_private_failure_message() {
  local label="$1"
  local log_file="$2"
  local redacted=""

  if [[ ! -s "${log_file}" ]]; then
    printf '%s process failed to establish a session\n' "${label}"
    return 0
  fi

  redacted="$(tail -c 4096 "${log_file}" | step_redact_stream)"
  CARTULARY_PRIVATE_FAILURE_TEXT="${redacted}" \
  CARTULARY_PRIVATE_FAILURE_LABEL="${label}" \
  CARTULARY_PRIVATE_FAILURE_TEST_ROUTE_TOKEN="${TEST_ROUTE_TOKEN:-}" \
  CARTULARY_PRIVATE_FAILURE_REVISIONS_TOKEN="${REVISIONS_CONFLICT_TOKEN_SECRET:-}" \
  CARTULARY_PRIVATE_FAILURE_DSN="${E2E_DSN:-}" \
  CARTULARY_PRIVATE_FAILURE_S3_ACCESS_KEY="${CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID:-}" \
  CARTULARY_PRIVATE_FAILURE_S3_SECRET_KEY="${CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY:-}" \
    "${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}" <<'EOF'
const secrets = [
  process.env.CARTULARY_PRIVATE_FAILURE_TEST_ROUTE_TOKEN,
  process.env.CARTULARY_PRIVATE_FAILURE_REVISIONS_TOKEN,
  process.env.CARTULARY_PRIVATE_FAILURE_DSN,
  process.env.CARTULARY_PRIVATE_FAILURE_S3_ACCESS_KEY,
  process.env.CARTULARY_PRIVATE_FAILURE_S3_SECRET_KEY,
].filter((value) => typeof value === "string" && value.length >= 8);
let text = process.env.CARTULARY_PRIVATE_FAILURE_TEXT ?? "";
for (const secret of secrets) text = text.replaceAll(secret, "[REDACTED]");
text = text.replace(
  /\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis):\/\/[^\s]+/giu,
  "[REDACTED]",
);
text = text.replace(/[\u0000-\u001f\u007f]+/gu, " ").trim();
if (text.length > 1500) text = text.slice(-1500);
process.stdout.write(`${process.env.CARTULARY_PRIVATE_FAILURE_LABEL} process failed to establish a session: ${text}\n`);
EOF
}

using_test_services_stack() {
  [[ -n "${SUITE_ID}" && -n "${CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:-}" && \
    -n "${CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:-}" && \
    -n "${CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:-}" && \
    -f "${CARTULARY_HARNESS_SUITE_RUNTIME_ROOT}/runtime-owner.json" && \
    -f "${CARTULARY_HARNESS_SUITE_RUNTIME_ROOT}/test-services/service-lease.json" && \
    -n "${CARTULARY_PGTEST_TEMPLATE_DB:-}" && \
    -n "${CARTULARY_PGTEST_SCHEMA_HASH:-}" && \
    -n "${CARTULARY_S3TEST_ENDPOINT:-}" ]]
}

require_test_services_bin() {
  if [[ -z "${TEST_SERVICES_BIN}" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN is required for an authenticated managed-service stack" >&2
    return 1
  fi
  if [[ ! -x "${TEST_SERVICES_BIN}" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN ${TEST_SERVICES_BIN} is not executable" >&2
    return 1
  fi
}

resolve_runtime_command() {
  local outvar="$1"
  local label="$2"
  local configured_path="$3"
  local -n resolved_ref="$outvar"

  resolved_ref=()

  if [[ -z "${configured_path}" ]]; then
    echo "${label} requires its scheduler-produced runtime binary" >&2
    return 1
  fi
  if [[ ! -x "${configured_path}" ]]; then
    echo "${label} runtime binary ${configured_path} is not executable" >&2
    return 1
  fi
  # shellcheck disable=SC2034
  resolved_ref=("${configured_path}")
}

backend_start_input_failure() {
  local name=""
  local value=""
  local -a missing=()
  for name in \
    SERVER_HARNESS_BIN \
    CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT \
    CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID \
    CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY \
    CARTULARY_S3_OBJECT_PRIMARY_SECURE \
    CARTULARY_S3_OBJECT_PRIMARY_BUCKET \
    GO_CACHE_DIR \
    GO_MOD_CACHE_DIR \
    GO_TMP_DIR; do
    value="${!name:-}"
    if [[ -z "${value}" ]]; then
      missing+=("${name}")
    fi
  done
  if [[ "${#missing[@]}" -gt 0 ]]; then
    printf 'backend start inputs are missing: %s\n' "$(IFS=,; printf '%s' "${missing[*]}")"
    return 0
  fi
  if [[ ! -x "${SERVER_HARNESS_BIN}" ]]; then
    printf 'backend scheduler-produced runtime binary is not executable\n'
  fi
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
  local status=0

  if [[ -z "${group_id}" ]]; then
    wait_for_port_release "${port}" "${name}" || status=$?
    return "${status}"
  fi

  stop_process_group "${group_id}" || status=$?
  wait_for_port_release "${port}" "${name}" || status=$?
  return "${status}"
}

remove_private_runtime_material() {
  local candidate=""
  local status=0
  local -a candidates=()

  if [[ -n "${TEST_ROUTE_TOKEN_FILE:-}" ]]; then
    candidates+=("${TEST_ROUTE_TOKEN_FILE}")
  fi
  if [[ -n "${TEST_SERVICES_ENV_FILE:-}" ]]; then
    candidates+=("${TEST_SERVICES_ENV_FILE}")
  fi
  if [[ -n "${SESSION_ENV_FILE:-}" ]]; then
    candidates+=("${SESSION_ENV_FILE}")
  fi
  if [[ -n "${PLAYWRIGHT_STATE_DIR:-}" ]]; then
    candidates+=(
      "${PLAYWRIGHT_STATE_DIR}/cartulary-playwright-admin-totp.txt"
      "${PLAYWRIGHT_STATE_DIR}/cartulary-playwright-worker-admins.json"
    )
  fi

  for candidate in "${candidates[@]}"; do
    if ! rm -f -- "${candidate}" >/dev/null 2>&1; then
      status=1
    fi
  done
  if [[ "${status}" -ne 0 ]]; then
    echo "browser e2e cleanup could not remove private runtime material" >&2
  fi
  return "${status}"
}

cleanup() {
  if [[ "${cleanup_done}" -eq 1 ]]; then
    return 0
  fi
  cleanup_done=1

  local step_start_time
  local step_start_ms
  local step_end_time
  local step_end_ms
  local step_duration_ms
  local cleanup_status=0
  local process_cleanup_complete=0
  local step_status=0
  local step_span_status="pass"

  step_start_time="$(step_now_utc)"
  step_start_ms="$(step_now_monotonic_ms)"

  if [[ -n "${CHILD_PGID:-}" ]]; then
    stop_process_group "${CHILD_PGID}" || cleanup_status=$?
  fi
  stop_owned_process_group "${VITE_PGID:-}" "${FRONTEND_PORT:-4173}" "frontend" || cleanup_status=$?
  stop_owned_process_group "${SERVER_PGID:-}" "${BACKEND_PORT:-8080}" "backend" || cleanup_status=$?
  release_port_leases || cleanup_status=$?

  step_end_time="$(step_now_utc)"
  step_end_ms="$(step_now_monotonic_ms)"
  step_duration_ms="$(step_elapsed_ms "${step_start_ms}" "${step_end_ms}")"
  if [[ "${cleanup_status}" -ne 0 ]]; then
    step_span_status="fail"
  fi
  emit_target_timing_span "teardown" "browser-e2e stop owned processes" "${step_start_time}" "${step_end_time}" "${step_duration_ms}" "${step_span_status}" "${cleanup_status}"
  if [[ "${cleanup_status}" -eq 0 ]]; then
    process_cleanup_complete=1
  fi

  if [[ -x "${TEST_SERVICES_BIN}" && -f "${TEST_SERVICES_METADATA_FILE}" ]]; then
    CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE="${process_cleanup_complete}" \
      "${TEST_SERVICES_BIN}" cleanup-web-e2e --metadata-file "${TEST_SERVICES_METADATA_FILE}" || cleanup_status=$?
    if [[ "${cleanup_status}" -eq 0 ]]; then
      rm -f -- "${TEST_SERVICES_METADATA_FILE}" || cleanup_status=$?
    fi
  fi
  remove_private_runtime_material || cleanup_status=$?
  if [[ "${KEEP_RUNTIME_ROOT}" -ne 1 ]]; then
    step_start_time="$(step_now_utc)"
    step_start_ms="$(step_now_monotonic_ms)"
    step_status=0
    rm -rf "${RUNTIME_ROOT_BASE}" || step_status=$?
    step_end_time="$(step_now_utc)"
    step_end_ms="$(step_now_monotonic_ms)"
    step_duration_ms="$(step_elapsed_ms "${step_start_ms}" "${step_end_ms}")"
    step_span_status="pass"
    if [[ "${step_status}" -ne 0 ]]; then
      step_span_status="fail"
      cleanup_status="${step_status}"
    fi
    emit_target_timing_span "teardown" "browser-e2e remove runtime root" "${step_start_time}" "${step_end_time}" "${step_duration_ms}" "${step_span_status}" "${step_status}"
  fi

  return "${cleanup_status}"
}

release_process_group_monitor() {
  local group_id="$1"
  local monitor_pid=""

  if [[ -z "${group_id}" ]]; then
    return 0
  fi
  monitor_pid="${CARTULARY_LIFECYCLE_GROUP_MONITORS[$group_id]:-}"
  if [[ -z "${monitor_pid}" ]]; then
    return 0
  fi
  kill "${monitor_pid}" >/dev/null 2>&1 || true
  wait "${monitor_pid}" >/dev/null 2>&1 || true
  unset "CARTULARY_LIFECYCLE_GROUP_MONITORS[$group_id]"
}

write_session_files() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  step_secure_mkdir "$(dirname "${SESSION_ENV_FILE}")" "$(dirname "${SESSION_LEASE_FILE}")"
  CARTULARY_WEB_E2E_SESSION_ENV_FILE="${SESSION_ENV_FILE}" \
  CARTULARY_WEB_E2E_SESSION_LEASE_FILE="${SESSION_LEASE_FILE}" \
  CARTULARY_PLAYWRIGHT_STATE_DIR="${PLAYWRIGHT_STATE_DIR}" \
  CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
  CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
  CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}" \
  CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}" \
  CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}" \
  CARTULARY_WEB_E2E_SERVER_LOG="${SERVER_LOG}" \
  CARTULARY_WEB_E2E_WEB_LOG="${WEB_LOG}" \
  CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS="${STARTUP_DIAGNOSTIC_FILE}" \
  CARTULARY_WEB_E2E_FRONTEND_MODE="${FRONTEND_MODE}" \
  CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND="${FRONTEND_COMMAND_KIND}" \
  CARTULARY_TEST_ROUTE_TOKEN_FILE="${TEST_ROUTE_TOKEN_FILE}" \
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID="${RUNTIME_PROFILE_ID}" \
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT="${RUNTIME_PROFILE_FINGERPRINT}" \
  CARTULARY_WEB_E2E_SERVER_PGID="${SERVER_PGID}" \
  CARTULARY_WEB_E2E_VITE_PGID="${VITE_PGID}" \
  CARTULARY_WEB_E2E_KEEP_RUNTIME_ROOT="${KEEP_RUNTIME_ROOT}" \
  CARTULARY_WEB_E2E_DB="${E2E_DB}" \
  CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE="${TEST_SERVICES_METADATA_FILE}" \
  CARTULARY_WEB_E2E_PGTEST_SCHEMA_HASH="${CARTULARY_PGTEST_SCHEMA_HASH}" \
  CARTULARY_WEB_E2E_PGTEST_TEMPLATE_DB="${CARTULARY_PGTEST_TEMPLATE_DB}" \
  CARTULARY_WEB_E2E_S3_ENDPOINT="${CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT}" \
  CARTULARY_WEB_E2E_S3_SECURE="${CARTULARY_S3_OBJECT_PRIMARY_SECURE}" \
  CARTULARY_WEB_E2E_S3_BUCKET="${CARTULARY_S3_OBJECT_PRIMARY_BUCKET}" \
  CARTULARY_WEB_E2E_FIXTURE_PROFILE_ID="${CARTULARY_FIXTURE_PROFILE_ID:-}" \
  CARTULARY_WEB_E2E_FIXTURE_SNAPSHOT_KEY="${CARTULARY_FIXTURE_SNAPSHOT_KEY:-}" \
  CARTULARY_WEB_E2E_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID="${CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID:-}" \
  CARTULARY_WEB_E2E_FIXTURE_ROW_ID="${CARTULARY_FIXTURE_ROW_ID:-}" \
  CARTULARY_WEB_E2E_FIXTURE_PREDICATE_ID="${CARTULARY_FIXTURE_PREDICATE_ID:-}" \
  CARTULARY_WEB_E2E_FIXTURE_CLONE_LEASE_ID="${CARTULARY_FIXTURE_CLONE_LEASE_ID:-}" \
  CARTULARY_WEB_E2E_FIXTURE_CLONE_ORDINAL="${CARTULARY_FIXTURE_CLONE_ORDINAL:-}" \
  CARTULARY_WEB_E2E_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE="${CARTULARY_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE:-}" \
    "${node_bin}" <<'EOF'
const fs = require("node:fs");

const env = {
  CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER: "1",
  CARTULARY_PLAYWRIGHT_STATE_DIR: process.env.CARTULARY_PLAYWRIGHT_STATE_DIR,
  CARTULARY_WEB_E2E_API_ORIGIN: process.env.CARTULARY_WEB_E2E_API_ORIGIN,
  CARTULARY_WEB_E2E_PUBLIC_ORIGIN: process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN,
  CARTULARY_WEB_E2E_BACKEND_PORT: process.env.CARTULARY_WEB_E2E_BACKEND_PORT,
  CARTULARY_WEB_E2E_FRONTEND_PORT: process.env.CARTULARY_WEB_E2E_FRONTEND_PORT,
  CARTULARY_WEB_E2E_RUNTIME_ROOT: process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT,
  CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE: process.env.CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE,
  CARTULARY_WEB_E2E_SERVER_LOG: process.env.CARTULARY_WEB_E2E_SERVER_LOG,
  CARTULARY_WEB_E2E_WEB_LOG: process.env.CARTULARY_WEB_E2E_WEB_LOG,
  CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS: process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS,
  CARTULARY_WEB_E2E_STACK_JSON_FILE: process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE,
  CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR: process.env.CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR,
  CARTULARY_WEB_E2E_FRONTEND_MODE: process.env.CARTULARY_WEB_E2E_FRONTEND_MODE,
  CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND: process.env.CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND,
  CARTULARY_TEST_ROUTE_TOKEN_FILE: process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE,
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID,
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT,
  CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD: process.env.CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD,
  CARTULARY_WEB_E2E_BACKEND_RESTART_SECRET_FILE: process.env.CARTULARY_WEB_E2E_BACKEND_RESTART_SECRET_FILE,
  CARTULARY_PGTEST_SCHEMA_HASH: process.env.CARTULARY_WEB_E2E_PGTEST_SCHEMA_HASH,
  CARTULARY_PGTEST_TEMPLATE_DB: process.env.CARTULARY_WEB_E2E_PGTEST_TEMPLATE_DB,
  CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT: process.env.CARTULARY_WEB_E2E_S3_ENDPOINT,
  CARTULARY_S3_OBJECT_PRIMARY_SECURE: process.env.CARTULARY_WEB_E2E_S3_SECURE,
  CARTULARY_S3_OBJECT_PRIMARY_BUCKET: process.env.CARTULARY_WEB_E2E_S3_BUCKET,
  CARTULARY_FIXTURE_PROFILE_ID: process.env.CARTULARY_WEB_E2E_FIXTURE_PROFILE_ID,
  CARTULARY_FIXTURE_SNAPSHOT_KEY: process.env.CARTULARY_WEB_E2E_FIXTURE_SNAPSHOT_KEY,
  CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID: process.env.CARTULARY_WEB_E2E_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID,
  CARTULARY_FIXTURE_ROW_ID: process.env.CARTULARY_WEB_E2E_FIXTURE_ROW_ID,
  CARTULARY_FIXTURE_PREDICATE_ID: process.env.CARTULARY_WEB_E2E_FIXTURE_PREDICATE_ID,
  CARTULARY_FIXTURE_CLONE_LEASE_ID: process.env.CARTULARY_WEB_E2E_FIXTURE_CLONE_LEASE_ID,
  CARTULARY_FIXTURE_CLONE_ORDINAL: process.env.CARTULARY_WEB_E2E_FIXTURE_CLONE_ORDINAL,
  CARTULARY_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE: process.env.CARTULARY_WEB_E2E_PERFORMANCE_FIXTURE_RUNTIME_BUNDLE,
};
for (const [key, value] of Object.entries(env)) {
  if (value === undefined || value === "") delete env[key];
}
const lease = {
  schema_id: "cartulary.web_e2e_session_lease.v1",
  env,
  session_env_file: process.env.CARTULARY_WEB_E2E_SESSION_ENV_FILE,
  backend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_BACKEND_PORT ?? "", 10),
  frontend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_FRONTEND_PORT ?? "", 10),
  runtime_root: process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT,
  server_log: process.env.CARTULARY_WEB_E2E_SERVER_LOG,
  web_log: process.env.CARTULARY_WEB_E2E_WEB_LOG,
  startup_diagnostics: process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS,
  frontend_mode: process.env.CARTULARY_WEB_E2E_FRONTEND_MODE,
  frontend_command_kind: process.env.CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND,
  server_pgid: process.env.CARTULARY_WEB_E2E_SERVER_PGID,
  vite_pgid: process.env.CARTULARY_WEB_E2E_VITE_PGID,
  keep_runtime_root: process.env.CARTULARY_WEB_E2E_KEEP_RUNTIME_ROOT === "1",
  e2e_db: process.env.CARTULARY_WEB_E2E_DB,
  test_services_metadata_file: process.env.CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE,
  runtime_profile_id: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID,
  runtime_profile_fingerprint: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT,
  backend_generation_head: process.env.CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD,
  backend_restart_secret_file: process.env.CARTULARY_WEB_E2E_BACKEND_RESTART_SECRET_FILE,
  backend_generation: 1,
};

fs.writeFileSync(process.env.CARTULARY_WEB_E2E_SESSION_ENV_FILE, `${JSON.stringify(env, null, 2)}\n`, { mode: 0o600 });
fs.writeFileSync(process.env.CARTULARY_WEB_E2E_SESSION_LEASE_FILE, `${JSON.stringify(lease, null, 2)}\n`, { mode: 0o600 });
fs.chmodSync(process.env.CARTULARY_WEB_E2E_SESSION_ENV_FILE, 0o600);
fs.chmodSync(process.env.CARTULARY_WEB_E2E_SESSION_LEASE_FILE, 0o600);
EOF
}

publish_session_lease() {
  run_timing_span "setup" "browser-e2e write session lease" write_session_files || return $?
  verify_stack_publication || return $?
}

load_session_lease() {
  local lease_file="$1"
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  eval "$("${node_bin}" - "${lease_file}" <<'EOF'
const fs = require("node:fs");
const lease = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const q = (value) => JSON.stringify(String(value ?? ""));
console.log(`SERVER_PGID=${q(lease.server_pgid)}`);
console.log(`VITE_PGID=${q(lease.vite_pgid)}`);
console.log(`BACKEND_PORT=${q(lease.backend_port)}`);
console.log(`FRONTEND_PORT=${q(lease.frontend_port)}`);
console.log(`RUNTIME_ROOT_BASE=${q(lease.runtime_root)}`);
console.log(`SERVER_LOG=${q(lease.server_log)}`);
console.log(`WEB_LOG=${q(lease.web_log)}`);
console.log(`PLAYWRIGHT_STATE_DIR=${q(lease.env?.CARTULARY_PLAYWRIGHT_STATE_DIR)}`);
console.log(`TEST_ROUTE_TOKEN_FILE=${q(lease.env?.CARTULARY_TEST_ROUTE_TOKEN_FILE)}`);
console.log(`TEST_SERVICES_ENV_FILE=${q(lease.runtime_root ? `${lease.runtime_root}/test-services-web-e2e.env` : "")}`);
console.log(`SESSION_ENV_FILE=${q(lease.session_env_file)}`);
console.log(`KEEP_RUNTIME_ROOT=${lease.keep_runtime_root ? "1" : "0"}`);
console.log(`E2E_DB=${q(lease.e2e_db)}`);
console.log(`TEST_SERVICES_METADATA_FILE=${q(lease.test_services_metadata_file)}`);
console.log(`API_ORIGIN=${q(lease.env?.CARTULARY_WEB_E2E_API_ORIGIN)}`);
console.log(`PUBLIC_ORIGIN=${q(lease.env?.CARTULARY_WEB_E2E_PUBLIC_ORIGIN)}`);
console.log(`STARTUP_DIAGNOSTIC_FILE=${q(lease.startup_diagnostics)}`);
console.log(`STACK_JSON_FILE=${q(lease.env?.CARTULARY_WEB_E2E_STACK_JSON_FILE)}`);
console.log(`TARGET_ARTIFACT_DIR=${q(lease.env?.CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR)}`);
console.log(`RUNTIME_PROFILE_ID=${q(lease.runtime_profile_id)}`);
console.log(`EXPECTED_RUNTIME_PROFILE_FINGERPRINT=${q(lease.runtime_profile_fingerprint)}`);
console.log(`BACKEND_GENERATION_HEAD=${q(lease.backend_generation_head)}`);
console.log(`BACKEND_RESTART_SECRET_FILE=${q(lease.backend_restart_secret_file)}`);
EOF
  )"
  CARTULARY_TEST_ROUTE_TOKEN_FILE="${TEST_ROUTE_TOKEN_FILE}"
  export CARTULARY_TEST_ROUTE_TOKEN_FILE
  adopt_port_lease_for_cleanup "${BACKEND_PORT}"
  adopt_port_lease_for_cleanup "${FRONTEND_PORT}"
}

stop_session() {
  local status=0

  if [[ ! -f "${SESSION_LEASE_FILE}" ]]; then
    echo "browser e2e session lease ${SESSION_LEASE_FILE} is missing" >&2
    return 1
  fi
  load_session_lease "${SESSION_LEASE_FILE}"
  KEEP_RUNTIME_ROOT=0
  cleanup || status=$?
  if ! rm -f -- "${SESSION_LEASE_FILE}" >/dev/null 2>&1; then
    status=1
  fi
  return "${status}"
}

on_exit() {
  local status=$?
  local cleanup_status=0

  trap - EXIT
  set +e
  if [[ "${status}" -ne 0 && -n "${STARTUP_DIAGNOSTIC_FILE}" && ! -f "${STARTUP_DIAGNOSTIC_FILE}" ]]; then
    write_startup_diagnostics \
      "fail" \
      "initializing" \
      "infra" \
      "service_start_error" \
      "browser session lifecycle exited before publishing ready evidence" || true
  fi
  cleanup
  cleanup_status=$?
  set -e

  if [[ "${cleanup_status}" -ne 0 ]]; then
    echo "browser e2e cleanup failed with status ${cleanup_status}" >&2
    if [[ "${status}" -eq 0 ]]; then
      exit "${cleanup_status}"
    fi
  fi

  exit "${status}"
}

exit_for_requested_shutdown() {
  local context="$1"

  if ! lifecycle_shutdown_requested; then
    return 0
  fi

  echo "received $(lifecycle_signal_name) during ${context}; shutting down browser e2e stack" >&2
  return "$(lifecycle_signal_exit_status)"
}

port_owned_by_process_group() {
  local port="$1"
  local group_id="$2"
  local pids
  local pid
  local pgid

  if [[ -z "${group_id}" ]]; then
    return 1
  fi
  if ! command -v ss >/dev/null 2>&1 || ! command -v ps >/dev/null 2>&1; then
    return 0
  fi

  pids="$(ss -ltnp "sport = :${port}" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2 | sort -u || true)"
  if [[ -z "${pids}" ]]; then
    return 1
  fi

  while IFS= read -r pid; do
    [[ -n "${pid}" ]] || continue
    pgid="$(ps -o pgid= -p "${pid}" 2>/dev/null | tr -d ' ' || true)"
    if [[ "${pgid}" == "${group_id}" ]]; then
      return 0
    fi
  done <<<"${pids}"

  return 1
}

print_port_diagnostics() {
  local port="$1"
  local name="$2"

  if command -v ss >/dev/null 2>&1; then
    echo "${name} port ${port} listener diagnostics:" >&2
    ss -ltnp "sport = :${port}" >&2 || true
  fi
}

probe_backend_identity() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
  CARTULARY_TEST_ROUTE_TOKEN="${TEST_ROUTE_TOKEN}" \
    "${node_bin}" <<'EOF'
const apiOrigin = process.env.CARTULARY_WEB_E2E_API_ORIGIN;
const requestOrigin = process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN || apiOrigin;
const token = process.env.CARTULARY_TEST_ROUTE_TOKEN;

(async () => {
  if (!apiOrigin || !token) {
    throw new Error("missing browser harness identity probe inputs");
  }

  const response = await fetch(`${apiOrigin}/api/v1/test/runtime/identity`, {
    headers: {
      "Origin": requestOrigin,
      "X-Cartulary-Test-Route-Token": token,
    },
    signal: AbortSignal.timeout(1000),
  });
  if (!response.ok) {
    throw new Error(`identity probe returned HTTP ${response.status}`);
  }
  const body = await response.json();
  const data = body?.data;
  if (
    data?.schema_id !== "cartulary.test.runtime_identity.v1" ||
    data?.runtime_marker !== "harness-owned" ||
    data?.test_routes_enabled !== true ||
    !Number.isInteger(data?.server_pid)
  ) {
    throw new Error(`identity probe returned unexpected payload ${JSON.stringify(body)}`);
  }
  process.stdout.write(`${data.server_pid}\n`);
})().catch((error) => {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
});
EOF
}

wait_for_http() {
  local url="$1"
  local name="$2"

  for _ in $(seq 1 240); do
    if exit_for_requested_shutdown "${name} readiness"; then
      :
    else
      return "$?"
    fi
    if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
      echo "backend exited before ${name} readiness" >&2
      write_startup_diagnostics "fail" "backend_readiness" "infra" "service_start_error" "backend exited before ${name} readiness" || true
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
      echo "frontend exited before ${name} readiness" >&2
      write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_start_error" "frontend exited before ${name} readiness" || true
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done

  echo "timed out waiting for ${name} at ${url}" >&2
  write_startup_diagnostics "fail" "${name}_readiness" "infra" "service_readiness_timeout" "timed out waiting for ${name} at ${url}" || true
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
    local message="${name} port ${port} is already in use by an unowned listener"
    echo "${message}" >&2
    ss -ltnp "sport = :${port}" >&2 || true
    write_startup_diagnostics \
      "fail" \
      "${name}_readiness" \
      "infra" \
      "resource_conflict" \
      "${message}" || true
    return 1
  fi
}

browser_start_services() {
  if ! using_test_services_stack; then
    echo "browser e2e refuses shared development services; run the Make-owned browser target" >&2
    return 2
  fi
  require_test_services_bin || return $?
  echo "browser e2e using active isolated test-service Postgres and object-store stack"
  record_startup_event "service_attached" \
    "attached browser session ${BROWSER_SESSION_ID} to suite ${SUITE_ID}"
}

browser_prepare_database() {
  assert_port_free "${BACKEND_PORT}" "backend"
  cd "${ROOT_DIR}"

  if ! using_test_services_stack; then
    echo "browser database preparation requires an active isolated test-services suite" >&2
    return 2
  fi
  require_test_services_bin || return $?
  if [[ -z "${CARTULARY_PGTEST_TEMPLATE_DB:-}" ]]; then
    echo "browser e2e active test-service mode requires CARTULARY_PGTEST_TEMPLATE_DB to clone the migrated suite template database" >&2
    return 1
  fi
  "${TEST_SERVICES_BIN}" prepare-web-e2e --env-file "${TEST_SERVICES_ENV_FILE}" --metadata-file "${TEST_SERVICES_METADATA_FILE}"
  # shellcheck disable=SC1090
  source "${TEST_SERVICES_ENV_FILE}"
  E2E_DSN="${CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN:?}"
  E2E_DB="$("${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}" -e \
    'const fs=require("node:fs"); process.stdout.write(String(JSON.parse(fs.readFileSync(process.argv[1],"utf8")).database_name));' \
    "${TEST_SERVICES_METADATA_FILE}")"
  export CARTULARY_WEB_E2E_DB="${E2E_DB}"
  export CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN="${E2E_DSN}"
  snapshot_service_scope || return $?
  record_startup_event "fixture_ready" \
    "prepared isolated browser database and object-store fixture"
}

browser_wait_backend_ready() {
  local identity_pid=""

  for _ in $(seq 1 240); do
    if exit_for_requested_shutdown "backend readiness"; then
      :
    else
      return "$?"
    fi
    if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
      echo "backend exited before readiness" >&2
      write_startup_diagnostics \
        "fail" \
        "backend_readiness" \
        "infra" \
        "service_start_error" \
        "$(bounded_private_failure_message "backend" "${SERVER_LOG}")" || true
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if port_owned_by_process_group "${BACKEND_PORT}" "${SERVER_PGID}" && identity_pid="$(probe_backend_identity 2>/dev/null)"; then
      if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
        echo "backend exited immediately after readiness identity probe" >&2
        write_startup_diagnostics \
          "fail" \
          "backend_readiness" \
          "infra" \
          "service_start_error" \
          "$(bounded_private_failure_message "backend" "${SERVER_LOG}")" || true
        cat "${SERVER_LOG}" >&2 || true
        return 1
      fi
      BACKEND_IDENTITY_SERVER_PID="${identity_pid}"
      return 0
    fi
    sleep 0.5
  done

  echo "timed out waiting for backend owned-runtime identity at ${API_ORIGIN}/api/v1/test/runtime/identity" >&2
  write_startup_diagnostics "fail" "backend_readiness" "infra" "service_readiness_timeout" "timed out waiting for backend owned-runtime identity at ${API_ORIGIN}/api/v1/test/runtime/identity" || true
  print_port_diagnostics "${BACKEND_PORT}" "backend"
  cat "${SERVER_LOG}" >&2 || true
  return 1
}

browser_wait_frontend_ready() {
  local failure_reason
  local failure_message

  for _ in $(seq 1 240); do
    if exit_for_requested_shutdown "frontend readiness"; then
      :
    else
      return "$?"
    fi
    if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
      failure_reason="service_start_error"
      failure_message="frontend exited before readiness"
      if [[ -f "${WEB_LOG}" ]] && grep -Eq 'Port [0-9]+ is already in use' "${WEB_LOG}"; then
        failure_reason="resource_conflict"
        failure_message="frontend port ${FRONTEND_PORT} became unavailable before readiness"
      fi
      echo "${failure_message}" >&2
      write_startup_diagnostics "fail" "frontend_readiness" "infra" "${failure_reason}" "${failure_message}" || true
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi
    if port_owned_by_process_group "${FRONTEND_PORT}" "${VITE_PGID}" && curl -fsS "${PUBLIC_ORIGIN}" >/dev/null 2>&1; then
      if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
        echo "frontend exited immediately after readiness probe" >&2
        write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_start_error" "frontend exited immediately after readiness probe" || true
        cat "${WEB_LOG}" >&2 || true
        return 1
      fi
      return 0
    fi
    sleep 0.5
  done

  echo "timed out waiting for frontend owned listener at ${PUBLIC_ORIGIN}" >&2
  write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_readiness_timeout" "timed out waiting for frontend owned listener at ${PUBLIC_ORIGIN}" || true
  print_port_diagnostics "${FRONTEND_PORT}" "frontend"
  cat "${WEB_LOG}" >&2 || true
  return 1
}

start_frontend_preview_process() {
  local pnpm_bin="$1"

  run_timing_span "frontend_startup" "browser-e2e start frontend process" \
  start_process_group VITE_PGID "${WEB_LOG}" \
    env \
    COREPACK_HOME="${NODE_RUNTIME_DIR}/corepack" \
    PATH="${NODE_RUNTIME_DIR}/bin:${PATH}" \
    CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
    CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
    "${pnpm_bin}" --dir apps/web exec vite preview --host 127.0.0.1 --port "${FRONTEND_PORT}" --strictPort
}

start_frontend_preview_ready() {
  local pnpm_bin="$1"

  start_frontend_preview_process "${pnpm_bin}"
  browser_wait_frontend_ready
}

browser_verify_frontend_ready() {
  if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
    echo "frontend exited before backend-ready verification" >&2
    write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_start_error" "frontend exited before backend-ready verification" || true
    cat "${WEB_LOG}" >&2 || true
    return 1
  fi
  if port_owned_by_process_group "${FRONTEND_PORT}" "${VITE_PGID}" && curl -fsS "${PUBLIC_ORIGIN}" >/dev/null 2>&1; then
    FRONTEND_READY_AT="${FRONTEND_READY_AT:-$(step_now_utc)}"
    return 0
  fi

  echo "frontend owned listener was not ready during backend-ready verification at ${PUBLIC_ORIGIN}" >&2
  write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_readiness_timeout" "frontend owned listener was not ready during backend-ready verification at ${PUBLIC_ORIGIN}" || true
  print_port_diagnostics "${FRONTEND_PORT}" "frontend"
  cat "${WEB_LOG}" >&2 || true
  return 1
}

start_backend_ready() {
  local backend_input_error=""
  backend_input_error="$(backend_start_input_failure)"
  if [[ -n "${backend_input_error}" ]]; then
    write_startup_diagnostics \
      "fail" \
      "backend_startup" \
      "infra" \
      "service_start_error" \
      "${backend_input_error}" || true
    return 1
  fi

  local -a server_command=()
  resolve_runtime_command server_command "backend" "${SERVER_HARNESS_BIN}"
  local -a backend_listen_command=(
    "${GO_BIN}" run ./tools/webstacklisten
    --listen "127.0.0.1:${BACKEND_PORT}"
    --
    "${server_command[@]}"
  )
  local -a runtime_profile_env=()
  if [[ "${RUNTIME_PROFILE_KIND}" == "network_flow_claimed" ]]; then
    runtime_profile_env=(
      CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED=true
      CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH="${ROOT_DIR}/${RUNTIME_PROFILE_KEY_RING_MANIFEST}"
      CARTULARY_SECRET_NETWORK_FLOW_CURSOR_ACTIVE="${RUNTIME_PROFILE_CURSOR_SECRET}"
      CARTULARY_SECRET_NETWORK_FLOW_SAFE_DIGEST_ACTIVE="${RUNTIME_PROFILE_SAFE_DIGEST_SECRET}"
    )
  fi

  local backend_start_status=0
  if run_timing_span "server_startup" "browser-e2e start backend process" \
    start_process_group SERVER_PGID "${SERVER_LOG}" \
      env \
      CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
      CARTULARY__APPLICATION__PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
      CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
      CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
      CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
      CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH="${ROOT_DIR}/configs/dev/revisions-conflict-token-key-ring.json" \
      CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_DEV_ACTIVE="${REVISIONS_CONFLICT_TOKEN_SECRET}" \
      CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN="${E2E_DSN}" \
      CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT="${CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT:?}" \
      CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID="${CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID:?}" \
      CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY="${CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY:?}" \
      CARTULARY_S3_OBJECT_PRIMARY_SECURE="${CARTULARY_S3_OBJECT_PRIMARY_SECURE:?}" \
      CARTULARY_S3_OBJECT_PRIMARY_BUCKET="${CARTULARY_S3_OBJECT_PRIMARY_BUCKET:?}" \
      CARTULARY_ENABLE_TEST_ROUTES=1 \
      CARTULARY_TEST_RUNTIME_MARKER=harness-owned \
      CARTULARY_TEST_ROUTE_TOKEN="${TEST_ROUTE_TOKEN}" \
      CARTULARY__ROOTS__BACKUP_STORAGE__PATH="${RUNTIME_ROOT_BASE}/backup-storage" \
      CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH="${RUNTIME_ROOT_BASE}/reference-pack-storage" \
      CARTULARY__ROOTS__TEMPORARY_WORK__PATH="${RUNTIME_ROOT_BASE}/temporary-work" \
      CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH="${RUNTIME_ROOT_BASE}/export-outputs" \
      "${runtime_profile_env[@]}" \
      GOCACHE="${GO_CACHE_DIR:?GO_CACHE_DIR is required}" \
      GOMODCACHE="${GO_MOD_CACHE_DIR:?GO_MOD_CACHE_DIR is required}" \
      GOTMPDIR="${GO_TMP_DIR:?GO_TMP_DIR is required}" \
      "${backend_listen_command[@]}"; then
    :
  else
    backend_start_status=$?
    write_startup_diagnostics \
      "fail" \
      "backend_startup" \
      "infra" \
      "service_start_error" \
      "$(bounded_private_failure_message "backend" "${SERVER_LOG}")" || true
    return "${backend_start_status}"
  fi

  CARTULARY_STEP_TIMING_BUCKET=server_startup run_step_command "browser-e2e startup backend ready" browser_wait_backend_ready
  BACKEND_READY_AT="$(step_now_utc)"
  if [[ "${SESSION_MODE}" != "reset" ]]; then
    record_startup_event "backend_ready" "backend ready at ${API_ORIGIN}"
  fi
}

wait_for_process_status() {
  local group_id="$1"

  if wait "${group_id}"; then
    return 0
  else
    return $?
  fi
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
      if wait_for_process_status "${SERVER_PGID}"; then
        server_status=0
      else
        server_status=$?
      fi
      echo "backend exited unexpectedly during browser e2e supervision (status=${server_status})" >&2
      cat "${SERVER_LOG}" >&2 || true
      if [[ -n "${CHILD_PGID:-}" ]]; then
        stop_process_group "${CHILD_PGID}" || true
      fi
      return 1
    fi

    if ! process_group_running "${VITE_PGID}"; then
      if wait_for_process_status "${VITE_PGID}"; then
        vite_status=0
      else
        vite_status=$?
      fi
      echo "frontend exited unexpectedly during browser e2e supervision (status=${vite_status})" >&2
      cat "${WEB_LOG}" >&2 || true
      if [[ -n "${CHILD_PGID:-}" ]]; then
        stop_process_group "${CHILD_PGID}" || true
      fi
      return 1
    fi

    if [[ -n "${CHILD_PGID:-}" ]] && ! process_group_running "${CHILD_PGID}"; then
      if wait_for_process_status "${CHILD_PGID}"; then
        child_status=0
      else
        child_status=$?
      fi
      return "${child_status}"
    fi

    sleep 1
  done
}

stop_backend_for_reset() {
  local drain_timeout_ms="${CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS:-}"
  local poll_count=""
  if [[ -z "${SERVER_PGID}" ]] || ! process_group_running "${SERVER_PGID}"; then
    echo "failure_class=harness reason=fixture_error browser reset backend is not running" >&2
    return 1
  fi
  if [[ ! "${drain_timeout_ms}" =~ ^[1-9][0-9]*$ ]]; then
    echo "failure_class=config reason=configuration_error browser reset drain deadline is invalid" >&2
    return 2
  fi
  poll_count="$(((drain_timeout_ms + 199) / 200))"
  kill -TERM -- "-${SERVER_PGID}" >/dev/null 2>&1 || true
  for _ in $(seq 1 "${poll_count}"); do
    if ! process_group_running "${SERVER_PGID}" && ! port_in_use "${BACKEND_PORT}"; then
      return 0
    fi
    sleep 0.2
  done
  echo "failure_class=timing reason=timeout_failure browser reset backend did not drain or release its listener" >&2
  return 13
}

next_backend_generation() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  "${node_bin}" - "${SESSION_LEASE_FILE}" <<'EOF'
const fs = require("node:fs");
const lease = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const current = Number.isSafeInteger(lease.backend_generation) ? lease.backend_generation : 1;
process.stdout.write(String(current + 1));
EOF
}

update_session_backend_lease() {
  local generation="$1"
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  CARTULARY_RESET_SERVER_PGID="${SERVER_PGID}" \
  CARTULARY_RESET_BACKEND_GENERATION="${generation}" \
  CARTULARY_RESET_BACKEND_GENERATION_HEAD="${BACKEND_GENERATION_HEAD}" \
    "${node_bin}" - "${SESSION_LEASE_FILE}" <<'EOF'
const fs = require("node:fs");
const file = process.argv[2];
const lease = JSON.parse(fs.readFileSync(file, "utf8"));
lease.server_pgid = process.env.CARTULARY_RESET_SERVER_PGID;
lease.backend_generation = Number.parseInt(process.env.CARTULARY_RESET_BACKEND_GENERATION, 10);
lease.backend_generation_head = process.env.CARTULARY_RESET_BACKEND_GENERATION_HEAD;
lease.env.CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD = process.env.CARTULARY_RESET_BACKEND_GENERATION_HEAD;
const temporary = `${file}.tmp-${process.pid}`;
fs.writeFileSync(temporary, `${JSON.stringify(lease, null, 2)}\n`, { mode: 0o600 });
fs.renameSync(temporary, file);
fs.chmodSync(file, 0o600);
EOF
}

reset_backend_failure_cleanup() {
  local status=$?
  trap - EXIT
  if [[ "${status}" -ne 0 && -n "${SERVER_PGID}" ]] && process_group_running "${SERVER_PGID}"; then
    stop_process_group "${SERVER_PGID}" || true
  fi
  exit "${status}"
}

reset_session_backend() {
  local generation=""
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
  local previous_server_pgid=""
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi
  load_session_lease "${SESSION_LEASE_FILE}"
  if [[ -z "${BACKEND_GENERATION_HEAD}" || -z "${BACKEND_RESTART_SECRET_FILE}" || -z "${STACK_JSON_FILE}" ]]; then
    echo "failure_class=config reason=configuration_error browser reset lease omits restart inputs" >&2
    return 2
  fi
  if [[ -n "${CARTULARY_FIXTURE_PROFILE_ID:-}" ]]; then
    echo "failure_class=config reason=configuration_error immutable performance fixture stacks cannot reset" >&2
    return 2
  fi
  TEST_ROUTE_TOKEN="$(tr -d '\r\n' <"${TEST_ROUTE_TOKEN_FILE}")"
  prepare_runtime_profile
  if [[ "${RUNTIME_PROFILE_FINGERPRINT}" != "${EXPECTED_RUNTIME_PROFILE_FINGERPRINT}" ]]; then
    echo "failure_class=config reason=configuration_error browser reset runtime profile fingerprint changed" >&2
    return 2
  fi
  load_backend_restart_secrets
  generation="$(next_backend_generation)"

  trap reset_backend_failure_cleanup EXIT
  previous_server_pgid="${SERVER_PGID}"
  stop_backend_for_reset
  release_process_group_monitor "${previous_server_pgid}"
  reclaim_port_lease_for_replacement "${BACKEND_PORT}" "${previous_server_pgid}"
  SERVER_PGID=""
  web_e2e_reset_database \
    "${ROOT_DIR}" \
    "${RESET_LABEL}" \
    "${RESET_DATABASE_DIAGNOSTIC_FILE}"
  printf '%s\n' "cleared" >"${RESET_OBJECT_STORE_MARKER_FILE}"
  web_e2e_reset_clear_playwright_state "${RESET_STATE_MARKER_FILE}"
  start_backend_ready
  printf '%s\n' "ready" >"${RESET_BACKEND_READY_MARKER_FILE}"
  transfer_port_lease_for_port "${BACKEND_PORT}" "${SERVER_PGID}"
  export CARTULARY_WEB_E2E_SERVER_PGID="${SERVER_PGID}"
  export CARTULARY_WEB_E2E_BACKEND_READY_AT="${BACKEND_READY_AT}"
  export CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID="${BACKEND_IDENTITY_SERVER_PID}"
  export CARTULARY_WEB_E2E_BACKEND_GENERATION_HEAD="${BACKEND_GENERATION_HEAD}"
  export CARTULARY_WEB_E2E_STACK_JSON_FILE="${STACK_JSON_FILE}"
  export CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR="${TARGET_ARTIFACT_DIR}"
  "${node_bin}" "${SESSION_EVIDENCE_HELPER}" \
    backend-generation "${RESET_LABEL}" "${generation}" >/dev/null
  update_session_backend_lease "${generation}"
  release_process_group_monitor "${SERVER_PGID}"
  trap - EXIT
}

main() {
  parse_child_command "$@"

  if [[ "${SESSION_MODE}" == "stop" ]]; then
    stop_session
    return $?
  fi
  if [[ "${SESSION_MODE}" == "reset" ]]; then
    reset_session_backend
    return $?
  fi

  prepare_runtime_profile

  # Establish the complete artifact identity in this shell before helper
  # functions are invoked through command substitutions. Exports performed
  # inside those substitutions cannot update the parent shell.
  ensure_harness_artifact_identity
  prepare_runtime_root

  trap on_exit EXIT
  lifecycle_reset_shutdown_state
  lifecycle_install_signal_traps
  record_startup_event "initializing" \
    "initializing browser session ${BROWSER_SESSION_ID} for runtime profile ${RUNTIME_PROFILE_ID}"

  run_timing_span "setup" "browser-e2e frontend toolchain" \
    browser_prepare_frontend_toolchain
  local pnpm_bin="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
  if [[ ! -x "${pnpm_bin}" ]]; then
    echo "repo-local pnpm was not found at ${pnpm_bin}; run make frontend-toolchain" >&2
    return 1
  fi

  CARTULARY_STEP_TIMING_BUCKET=setup run_step_command "browser-e2e allocate ports" resolve_owned_stack_ports
  CARTULARY_STEP_TIMING_BUCKET=setup run_step_command "browser-e2e prepare test route token" prepare_test_route_token
	REVISIONS_CONFLICT_TOKEN_SECRET="$(dd if=/dev/urandom bs=32 count=1 status=none | base64 | tr '+/' '-_' | tr -d '=\n')"
  CARTULARY_STEP_TIMING_BUCKET=frontend_startup run_step_command "browser-e2e validate frontend preview artifact" require_frontend_preview_artifacts

  CARTULARY_STEP_TIMING_BUCKET=service_wait run_step_command "browser-e2e startup services" browser_start_services
  CARTULARY_STEP_TIMING_BUCKET=migration run_step_command "browser-e2e startup database" browser_prepare_database
  write_backend_restart_secrets
  start_backend_ready
  CARTULARY_STEP_TIMING_BUCKET=frontend_startup run_step_command "browser-e2e startup frontend ready" start_frontend_preview_ready "${pnpm_bin}"
  FRONTEND_READY_AT="$(step_now_utc)"
  record_startup_event "frontend_ready" "frontend ready at ${PUBLIC_ORIGIN}"
  run_timing_span "setup" "browser-e2e finalize startup diagnostics" finalize_startup_ready || return $?
  publish_stack_metadata || return $?

  if [[ "${SESSION_MODE}" == "start" ]]; then
    publish_session_lease || return $?
    transfer_port_lease_for_port "${BACKEND_PORT}" "${SERVER_PGID}"
    transfer_port_lease_for_port "${FRONTEND_PORT}" "${VITE_PGID}"
    release_process_group_monitor "${SERVER_PGID}"
    release_process_group_monitor "${VITE_PGID}"
    trap - EXIT
    return 0
  fi

  if [[ "${#child_command[@]}" -gt 0 ]]; then
    start_process_group CHILD_PGID "" "${child_command[@]}"
  fi

  supervise_stack
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
