#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
# shellcheck source=tools/harness/browser/browser-lifecycle-adapter.sh
source "${ROOT_DIR}/tools/harness/browser/browser-lifecycle-adapter.sh"

COMPOSE_FILE="${ROOT_DIR}/docker-compose.dev.yml"
DEV_SERVICES_SCRIPT="${ROOT_DIR}/tools/harness/readiness/dev-services.sh"
GO_BIN="${GO:-go}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
SERVER_HARNESS_BIN="${CARTULARY_SERVER_HARNESS_BIN:-}"
MIGRATE_BIN="${CARTULARY_MIGRATE_BIN:-}"
TEST_SERVICES_BIN="${CARTULARY_TEST_SERVICES_BIN:-}"
TEST_SERVICE_FRONTEND_PORT_START=39000
TEST_SERVICE_FRONTEND_PORT_END=39199
TEST_SERVICE_FRONTEND_STAGE_WIDTH=100
TEST_SERVICE_FRONTEND_PREVIEW_RETRY_LIMIT=5
WEB_DIST_INDEX="${ROOT_DIR}/apps/web/dist/index.html"
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
PLAYWRIGHT_STATE_DIR=""
TEST_ROUTE_TOKEN=""
TEST_ROUTE_TOKEN_FILE=""
BACKEND_READY_AT=""
FRONTEND_READY_AT=""
BACKEND_IDENTITY_STATUS=""
BACKEND_IDENTITY_SERVER_PID=""
FRONTEND_OWNERSHIP_STATUS=""
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
FRONTEND_PORT_CONFIGURED=0
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

usage() {
  echo "usage: start-web-e2e.sh [-- <command...>]" >&2
  echo "       start-web-e2e.sh --session-start --env-file <path> --lease-file <path>" >&2
  echo "       start-web-e2e.sh --session-stop --lease-file <path>" >&2
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
  local session_artifact_suffix=""
  TARGET_ARTIFACT_DIR="${CARTULARY_WEB_E2E_ARTIFACT_DIR:-}"
  if [[ -z "${TARGET_ARTIFACT_DIR}" && -n "${CARTULARY_TEST_TARGET:-}" ]]; then
    if [[ -n "${CARTULARY_BROWSER_SESSION_GROUP:-}" ]]; then
      session_artifact_suffix="-$(slugify_phase_label "${CARTULARY_BROWSER_SESSION_GROUP}")"
    fi
    TARGET_ARTIFACT_DIR="$(ensure_target_artifact_dir)/owned-stack${session_artifact_suffix}"
  fi

  if [[ -n "${TARGET_ARTIFACT_DIR}" ]]; then
    phase_secure_mkdir "${TARGET_ARTIFACT_DIR}"
    RUNTIME_ROOT_BASE="${TARGET_ARTIFACT_DIR}/runtime-root"
    SERVER_LOG="${TARGET_ARTIFACT_DIR}/server.log"
    WEB_LOG="${TARGET_ARTIFACT_DIR}/web.log"
    STACK_ENV_FILE="${TARGET_ARTIFACT_DIR}/stack.env"
    STACK_JSON_FILE="${TARGET_ARTIFACT_DIR}/stack.json"
    STARTUP_DIAGNOSTIC_FILE="${TARGET_ARTIFACT_DIR}/startup-diagnostics.json"
    rm -rf "${RUNTIME_ROOT_BASE}"
    rm -f "${SERVER_LOG}" "${WEB_LOG}" "${STACK_ENV_FILE}" "${STACK_JSON_FILE}" "${STARTUP_DIAGNOSTIC_FILE}"
    KEEP_RUNTIME_ROOT=1
  else
    RUNTIME_ROOT_BASE="$(mktemp -d /tmp/cartulary-web-e2e-runtime-XXXXXX)"
    SERVER_LOG="/tmp/cartulary-e2e-server-$$.log"
    WEB_LOG="/tmp/cartulary-e2e-web-$$.log"
    STACK_ENV_FILE="${RUNTIME_ROOT_BASE}/stack.env"
    STACK_JSON_FILE="${RUNTIME_ROOT_BASE}/stack.json"
    STARTUP_DIAGNOSTIC_FILE="${RUNTIME_ROOT_BASE}/startup-diagnostics.json"
  fi

  PLAYWRIGHT_STATE_DIR="${RUNTIME_ROOT_BASE}/playwright-state"
  E2E_DB="cartulary_web_e2e_$$"
  E2E_DSN="postgres://cartulary:cartulary@localhost:5432/${E2E_DB}?sslmode=disable"
  TEST_SERVICES_ENV_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.env"
  TEST_SERVICES_METADATA_FILE="${RUNTIME_ROOT_BASE}/test-services-web-e2e.json"
  TEST_ROUTE_TOKEN_FILE="${RUNTIME_ROOT_BASE}/test-route-token"

  phase_secure_mkdir \
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
  export CARTULARY_WEB_E2E_FRONTEND_MODE="${FRONTEND_MODE}"
  export CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND="${FRONTEND_COMMAND_KIND}"
  export CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}"
  export CARTULARY_TEST_ROUTE_TOKEN_FILE="${TEST_ROUTE_TOKEN_FILE}"
  export CARTULARY_WEB_E2E_DB="${E2E_DB}"
  export CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT="${CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT:-localhost:8333}"
  export CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID="${CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID:-cartulary-local}"
  export CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY="${CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY:-cartulary-local-secret}"
  export CARTULARY_S3_OBJECT_PRIMARY_SECURE="${CARTULARY_S3_OBJECT_PRIMARY_SECURE:-false}"
  export CARTULARY_S3_OBJECT_PRIMARY_BUCKET="${CARTULARY_S3_OBJECT_PRIMARY_BUCKET:-cartulary}"
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

write_stack_metadata() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  phase_secure_mkdir "$(dirname "${STACK_ENV_FILE}")"
  cat >"${STACK_ENV_FILE}" <<EOF
CARTULARY_WEB_E2E_API_ORIGIN=${API_ORIGIN}
CARTULARY_WEB_E2E_PUBLIC_ORIGIN=${PUBLIC_ORIGIN}
CARTULARY_WEB_E2E_BACKEND_PORT=${BACKEND_PORT}
CARTULARY_WEB_E2E_FRONTEND_PORT=${FRONTEND_PORT}
CARTULARY_WEB_E2E_RUNTIME_ROOT=${RUNTIME_ROOT_BASE}
CARTULARY_WEB_E2E_SERVER_LOG=${SERVER_LOG}
CARTULARY_WEB_E2E_WEB_LOG=${WEB_LOG}
CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS=${STARTUP_DIAGNOSTIC_FILE}
CARTULARY_WEB_E2E_FRONTEND_MODE=${FRONTEND_MODE}
CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND=${FRONTEND_COMMAND_KIND}
CARTULARY_TEST_ROUTE_TOKEN_FILE=${TEST_ROUTE_TOKEN_FILE}
CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID=${RUNTIME_PROFILE_ID}
CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT=${RUNTIME_PROFILE_FINGERPRINT}
EOF
  chmod 600 "${STACK_ENV_FILE}" 2>/dev/null || true

  CARTULARY_WEB_E2E_STACK_JSON_FILE="${STACK_JSON_FILE}" \
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
  CARTULARY_WEB_E2E_BACKEND_READY_AT="${BACKEND_READY_AT}" \
  CARTULARY_WEB_E2E_FRONTEND_READY_AT="${FRONTEND_READY_AT}" \
  CARTULARY_WEB_E2E_BACKEND_IDENTITY_STATUS="${BACKEND_IDENTITY_STATUS}" \
  CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID="${BACKEND_IDENTITY_SERVER_PID}" \
  CARTULARY_WEB_E2E_FRONTEND_OWNERSHIP_STATUS="${FRONTEND_OWNERSHIP_STATUS}" \
  CARTULARY_WEB_E2E_DB="${E2E_DB}" \
  CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE="${TEST_SERVICES_METADATA_FILE}" \
  CARTULARY_TEST_SERVICES_ACTIVE="${CARTULARY_TEST_SERVICES_ACTIVE:-}" \
    "${node_bin}" <<'EOF'
const fs = require("node:fs");

const fixtureIdentity = resolveFixtureIdentity();

const payload = {
  schema_id: "cartulary.web_e2e_stack.v3",
  api_origin: process.env.CARTULARY_WEB_E2E_API_ORIGIN,
  public_origin: process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN,
  backend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_BACKEND_PORT ?? "", 10),
  frontend_port: Number.parseInt(process.env.CARTULARY_WEB_E2E_FRONTEND_PORT ?? "", 10),
  frontend_mode: process.env.CARTULARY_WEB_E2E_FRONTEND_MODE,
  frontend_command_kind: process.env.CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND,
  runtime_root: process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT,
  server_log: process.env.CARTULARY_WEB_E2E_SERVER_LOG,
  web_log: process.env.CARTULARY_WEB_E2E_WEB_LOG,
  startup_diagnostics: stringOrUndefined(process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS),
  test_route_token_file: process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE,
  runtime_profile_id: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID,
  runtime_profile_fingerprint: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT,
  backend_process_group_id: numberOrUndefined(process.env.CARTULARY_WEB_E2E_SERVER_PGID),
  frontend_process_group_id: numberOrUndefined(process.env.CARTULARY_WEB_E2E_VITE_PGID),
  backend_ready_at: stringOrUndefined(process.env.CARTULARY_WEB_E2E_BACKEND_READY_AT),
  frontend_ready_at: stringOrUndefined(process.env.CARTULARY_WEB_E2E_FRONTEND_READY_AT),
  backend_identity: process.env.CARTULARY_WEB_E2E_BACKEND_IDENTITY_STATUS
    ? {
        status: process.env.CARTULARY_WEB_E2E_BACKEND_IDENTITY_STATUS,
        server_pid: numberOrUndefined(process.env.CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID),
        schema_id: "cartulary.test.runtime_identity.v1",
      }
    : undefined,
  frontend_ownership: process.env.CARTULARY_WEB_E2E_FRONTEND_OWNERSHIP_STATUS
    ? {
        status: process.env.CARTULARY_WEB_E2E_FRONTEND_OWNERSHIP_STATUS,
      }
    : undefined,
  database: fixtureIdentity,
};

function stringOrUndefined(value) {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? undefined : normalized;
}

function numberOrUndefined(value) {
  const normalized = value?.trim() ?? "";
  if (normalized === "") {
    return undefined;
  }
  const parsed = Number.parseInt(normalized, 10);
  return Number.isInteger(parsed) ? parsed : undefined;
}

function resolveFixtureIdentity() {
  if (process.env.CARTULARY_TEST_SERVICES_ACTIVE === "1") {
    const metadataFile = process.env.CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE;
    if (metadataFile && fs.existsSync(metadataFile)) {
      const metadata = JSON.parse(fs.readFileSync(metadataFile, "utf8"));
      if (metadata.database_name) {
        return {
          logical_id: metadata.database_name,
          source: "testservices",
          bucket: metadata.bucket,
        };
      }
    }
    return undefined;
  }
  if (process.env.CARTULARY_WEB_E2E_DB) {
    return {
      logical_id: process.env.CARTULARY_WEB_E2E_DB,
      source: "standalone",
    };
  }
  return undefined;
}

for (const key of Object.keys(payload)) {
  if (payload[key] === undefined) {
    delete payload[key];
  }
}

fs.writeFileSync(process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE, `${JSON.stringify(payload, null, 2)}\n`, { mode: 0o600 });
fs.chmodSync(process.env.CARTULARY_WEB_E2E_STACK_JSON_FILE, 0o600);
EOF
}

write_startup_diagnostics() {
  local status="$1"
  local phase="$2"
  local failure_class="${3:-}"
  local failure_reason="${4:-}"
  local message="${5:-}"
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  if [[ -z "${STARTUP_DIAGNOSTIC_FILE}" ]]; then
    return 0
  fi
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  phase_secure_mkdir "$(dirname "${STARTUP_DIAGNOSTIC_FILE}")"
  CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS="${STARTUP_DIAGNOSTIC_FILE}" \
  CARTULARY_WEB_E2E_DIAGNOSTIC_STATUS="${status}" \
  CARTULARY_WEB_E2E_DIAGNOSTIC_PHASE="${phase}" \
  CARTULARY_WEB_E2E_DIAGNOSTIC_FAILURE_CLASS="${failure_class}" \
  CARTULARY_WEB_E2E_DIAGNOSTIC_FAILURE_REASON="${failure_reason}" \
  CARTULARY_WEB_E2E_DIAGNOSTIC_MESSAGE="${message}" \
  CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
  CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
  CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}" \
  CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}" \
  CARTULARY_WEB_E2E_RUNTIME_ROOT="${RUNTIME_ROOT_BASE}" \
  CARTULARY_WEB_E2E_SERVER_LOG="${SERVER_LOG}" \
  CARTULARY_WEB_E2E_WEB_LOG="${WEB_LOG}" \
  CARTULARY_WEB_E2E_FRONTEND_MODE="${FRONTEND_MODE}" \
  CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND="${FRONTEND_COMMAND_KIND}" \
  CARTULARY_WEB_E2E_SERVER_PGID="${SERVER_PGID}" \
  CARTULARY_WEB_E2E_VITE_PGID="${VITE_PGID}" \
    "${node_bin}" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");

const outputPath = process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS;
const message = stringOrUndefined(process.env.CARTULARY_WEB_E2E_DIAGNOSTIC_MESSAGE);
const backendLogSample = readLogSample(process.env.CARTULARY_WEB_E2E_SERVER_LOG);
const frontendLogSample = readLogSample(process.env.CARTULARY_WEB_E2E_WEB_LOG);
const combinedText = `${message ?? ""}\n${backendLogSample ?? ""}\n${frontendLogSample ?? ""}`;
const inotify = /ENOSPC|System limit for number of file watchers reached|inotify/i.test(combinedText)
  ? collectInotifyDiagnostics()
  : undefined;
const normalizedFailure = normalizeFailure(
  process.env.CARTULARY_WEB_E2E_DIAGNOSTIC_FAILURE_CLASS,
  process.env.CARTULARY_WEB_E2E_DIAGNOSTIC_FAILURE_REASON,
  combinedText,
);

const payload = {
  schema_id: "cartulary.browser_startup_diagnostics.v1",
  generated_at: new Date().toISOString(),
  target: stringOrUndefined(process.env.CARTULARY_TEST_TARGET),
  status: process.env.CARTULARY_WEB_E2E_DIAGNOSTIC_STATUS,
  startup_phase: process.env.CARTULARY_WEB_E2E_DIAGNOSTIC_PHASE,
  frontend_mode: process.env.CARTULARY_WEB_E2E_FRONTEND_MODE,
  frontend_command_kind: process.env.CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND,
  api_origin: stringOrUndefined(process.env.CARTULARY_WEB_E2E_API_ORIGIN),
  public_origin: stringOrUndefined(process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN),
  backend_port: numberOrUndefined(process.env.CARTULARY_WEB_E2E_BACKEND_PORT),
  frontend_port: numberOrUndefined(process.env.CARTULARY_WEB_E2E_FRONTEND_PORT),
  backend_process_group_id: numberOrUndefined(process.env.CARTULARY_WEB_E2E_SERVER_PGID),
  frontend_process_group_id: numberOrUndefined(process.env.CARTULARY_WEB_E2E_VITE_PGID),
  runtime_root: stringOrUndefined(process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT),
  failure_class: normalizedFailure.failure_class,
  failure_reason: normalizedFailure.failure_reason,
  message,
  logs: {
    backend: stringOrUndefined(process.env.CARTULARY_WEB_E2E_SERVER_LOG),
    frontend: stringOrUndefined(process.env.CARTULARY_WEB_E2E_WEB_LOG),
  },
  inotify,
};

for (const key of Object.keys(payload)) {
  if (payload[key] === undefined) {
    delete payload[key];
  }
}
if (payload.logs && Object.values(payload.logs).every((value) => value === undefined)) {
  delete payload.logs;
}

fs.mkdirSync(path.dirname(outputPath), { recursive: true, mode: 0o700 });
fs.writeFileSync(outputPath, `${JSON.stringify(payload, null, 2)}\n`, { mode: 0o600 });
fs.chmodSync(outputPath, 0o600);

function stringOrUndefined(value) {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? undefined : normalized;
}

function numberOrUndefined(value) {
  const normalized = value?.trim() ?? "";
  if (normalized === "") {
    return undefined;
  }
  const parsed = Number.parseInt(normalized, 10);
  return Number.isInteger(parsed) ? parsed : undefined;
}

function readLogSample(file) {
  if (!file || !fs.existsSync(file)) {
    return undefined;
  }
  const text = fs.readFileSync(file, "utf8");
  return text.split(/\r?\n/).slice(-40).join("\n");
}

function normalizeFailure(failureClass, failureReason, text) {
  const inputClass = stringOrUndefined(failureClass);
  const inputReason = stringOrUndefined(failureReason);
  if (!inputClass && !inputReason) {
    return {};
  }
  if (isResourceConflict(text)) {
    return {
      failure_class: "infra",
      failure_reason: "resource_conflict",
    };
  }
  return {
    failure_class: inputClass,
    failure_reason: inputReason,
  };
}

function isResourceConflict(text) {
  return /(?:port .* is already in use|is already in use|strictPort|eaddrinuse|address already in use|listener conflict|failed to allocate an available|port must differ)/i.test(text);
}

function readIntFile(file) {
  try {
    const parsed = Number.parseInt(fs.readFileSync(file, "utf8").trim(), 10);
    return Number.isInteger(parsed) ? parsed : undefined;
  } catch {
    return undefined;
  }
}

function collectInotifyDiagnostics() {
  const diagnostics = {
    platform: process.platform,
    max_user_watches: readIntFile("/proc/sys/fs/inotify/max_user_watches"),
    max_user_instances: readIntFile("/proc/sys/fs/inotify/max_user_instances"),
    remediation:
      "Raise fs.inotify.max_user_watches and fs.inotify.max_user_instances or stop stale watcher-heavy processes; the harness does not mutate host sysctl settings.",
  };
  if (process.platform !== "linux") {
    return diagnostics;
  }

  let currentInstances = 0;
  let currentWatches = 0;
  try {
    for (const pid of fs.readdirSync("/proc")) {
      if (!/^\d+$/.test(pid)) {
        continue;
      }
      const fdinfoDir = `/proc/${pid}/fdinfo`;
      let fdinfos = [];
      try {
        fdinfos = fs.readdirSync(fdinfoDir);
      } catch {
        continue;
      }
      for (const fdinfo of fdinfos) {
        let text = "";
        try {
          text = fs.readFileSync(`${fdinfoDir}/${fdinfo}`, "utf8");
        } catch {
          continue;
        }
        const matches = text.match(/^inotify\b/gm);
        if (matches) {
          currentInstances += 1;
          currentWatches += matches.length;
        }
      }
    }
    diagnostics.current_instances = currentInstances;
    diagnostics.current_watches = currentWatches;
  } catch {
    diagnostics.current_usage_available = false;
  }
  return diagnostics;
}
EOF
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

cleanup_standalone_database() {
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres \
    -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${E2E_DB}' AND pid <> pg_backend_pid();" \
    -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null 2>&1
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
  local step_status=0
  local step_span_status="pass"

  step_start_time="$(phase_now_utc)"
  step_start_ms="$(phase_now_monotonic_ms)"

  if [[ -n "${CHILD_PGID:-}" ]]; then
    stop_process_group "${CHILD_PGID}" || cleanup_status=$?
  fi
  stop_owned_process_group "${VITE_PGID:-}" "${FRONTEND_PORT:-4173}" "frontend" || cleanup_status=$?
  stop_owned_process_group "${SERVER_PGID:-}" "${BACKEND_PORT:-8080}" "backend" || cleanup_status=$?
  release_port_leases || cleanup_status=$?

  step_end_time="$(phase_now_utc)"
  step_end_ms="$(phase_now_monotonic_ms)"
  step_duration_ms="$(phase_elapsed_ms "${step_start_ms}" "${step_end_ms}")"
  if [[ "${cleanup_status}" -ne 0 ]]; then
    step_span_status="fail"
  fi
  emit_target_timing_span "teardown" "browser-e2e stop owned processes" "${step_start_time}" "${step_end_time}" "${step_duration_ms}" "${step_span_status}" "${cleanup_status}"

  if using_test_services_stack; then
    if [[ -x "${TEST_SERVICES_BIN}" && -f "${TEST_SERVICES_METADATA_FILE}" ]]; then
      "${TEST_SERVICES_BIN}" cleanup-web-e2e --metadata-file "${TEST_SERVICES_METADATA_FILE}" || cleanup_status=$?
    fi
  else
    step_start_time="$(phase_now_utc)"
    step_start_ms="$(phase_now_monotonic_ms)"
    step_status=0
    cleanup_standalone_database || step_status=$?
    step_end_time="$(phase_now_utc)"
    step_end_ms="$(phase_now_monotonic_ms)"
    step_duration_ms="$(phase_elapsed_ms "${step_start_ms}" "${step_end_ms}")"
    step_span_status="pass"
    if [[ "${step_status}" -ne 0 ]]; then
      step_span_status="fail"
      cleanup_status="${step_status}"
    fi
    emit_target_timing_span "teardown" "browser-e2e cleanup standalone database" "${step_start_time}" "${step_end_time}" "${step_duration_ms}" "${step_span_status}" "${step_status}"
  fi
  if [[ "${KEEP_RUNTIME_ROOT}" -ne 1 ]]; then
    step_start_time="$(phase_now_utc)"
    step_start_ms="$(phase_now_monotonic_ms)"
    step_status=0
    rm -rf "${RUNTIME_ROOT_BASE}" || step_status=$?
    step_end_time="$(phase_now_utc)"
    step_end_ms="$(phase_now_monotonic_ms)"
    step_duration_ms="$(phase_elapsed_ms "${step_start_ms}" "${step_end_ms}")"
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

  phase_secure_mkdir "$(dirname "${SESSION_ENV_FILE}")" "$(dirname "${SESSION_LEASE_FILE}")"
  CARTULARY_WEB_E2E_SESSION_ENV_FILE="${SESSION_ENV_FILE}" \
  CARTULARY_WEB_E2E_SESSION_LEASE_FILE="${SESSION_LEASE_FILE}" \
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
  CARTULARY_WEB_E2E_TEST_SERVICES_ACTIVE="${CARTULARY_TEST_SERVICES_ACTIVE:-}" \
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
  CARTULARY_WEB_E2E_SERVER_LOG: process.env.CARTULARY_WEB_E2E_SERVER_LOG,
  CARTULARY_WEB_E2E_WEB_LOG: process.env.CARTULARY_WEB_E2E_WEB_LOG,
  CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS: process.env.CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS,
  CARTULARY_WEB_E2E_FRONTEND_MODE: process.env.CARTULARY_WEB_E2E_FRONTEND_MODE,
  CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND: process.env.CARTULARY_WEB_E2E_FRONTEND_COMMAND_KIND,
  CARTULARY_TEST_ROUTE_TOKEN_FILE: process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE,
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID,
  CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT,
};
const lease = {
  schema_id: "cartulary.web_e2e_session_lease.v1",
  env,
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
  test_services_active: process.env.CARTULARY_WEB_E2E_TEST_SERVICES_ACTIVE === "1",
  runtime_profile_id: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID,
  runtime_profile_fingerprint: process.env.CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT,
};

fs.writeFileSync(process.env.CARTULARY_WEB_E2E_SESSION_ENV_FILE, `${JSON.stringify(env, null, 2)}\n`, { mode: 0o600 });
fs.writeFileSync(process.env.CARTULARY_WEB_E2E_SESSION_LEASE_FILE, `${JSON.stringify(lease, null, 2)}\n`, { mode: 0o600 });
fs.chmodSync(process.env.CARTULARY_WEB_E2E_SESSION_ENV_FILE, 0o600);
fs.chmodSync(process.env.CARTULARY_WEB_E2E_SESSION_LEASE_FILE, 0o600);
EOF
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
console.log(`CARTULARY_TEST_ROUTE_TOKEN_FILE=${q(lease.env?.CARTULARY_TEST_ROUTE_TOKEN_FILE)}`);
console.log(`KEEP_RUNTIME_ROOT=${lease.keep_runtime_root ? "1" : "0"}`);
console.log(`E2E_DB=${q(lease.e2e_db)}`);
console.log(`TEST_SERVICES_METADATA_FILE=${q(lease.test_services_metadata_file)}`);
console.log(`CARTULARY_TEST_SERVICES_ACTIVE=${lease.test_services_active ? "1" : ""}`);
EOF
  )"
  export CARTULARY_TEST_SERVICES_ACTIVE
  export CARTULARY_TEST_ROUTE_TOKEN_FILE
}

stop_session() {
  if [[ ! -f "${SESSION_LEASE_FILE}" ]]; then
    echo "browser e2e session lease ${SESSION_LEASE_FILE} is missing" >&2
    return 1
  fi
  load_session_lease "${SESSION_LEASE_FILE}"
  cleanup
  rm -f "${SESSION_LEASE_FILE}"
}

on_exit() {
  local status=$?
  local cleanup_status=0

  trap - EXIT
  set +e
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
    echo "${name} port ${port} is already in use; stop the existing listener before browser e2e" >&2
    ss -ltnp "sport = :${port}" >&2 || true
    return 1
  fi
}

browser_start_services() {
  if using_test_services_stack; then
    require_test_services_bin || return $?
    echo "browser e2e using active test-service Postgres and object-store stack"
    return 0
  fi

  OBJECT_STORE_CORS_ORIGIN="${PUBLIC_ORIGIN}" \
  OBJECT_STORE_CORS_ALLOWED_ORIGINS="${PUBLIC_ORIGIN}" \
    "${DEV_SERVICES_SCRIPT}" up >/dev/null
}

browser_prepare_database() {
  assert_port_free "${BACKEND_PORT}" "backend"
  cd "${ROOT_DIR}"

  if using_test_services_stack; then
    require_test_services_bin || return $?
    if [[ -z "${CARTULARY_PGTEST_TEMPLATE_DB:-}" ]]; then
      echo "browser e2e active test-service mode requires CARTULARY_PGTEST_TEMPLATE_DB to clone the migrated suite template database" >&2
      return 1
    fi
    "${TEST_SERVICES_BIN}" prepare-web-e2e --env-file "${TEST_SERVICES_ENV_FILE}" --metadata-file "${TEST_SERVICES_METADATA_FILE}"
    # shellcheck disable=SC1090
    source "${TEST_SERVICES_ENV_FILE}"
    E2E_DSN="${CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN:?}"
    export CARTULARY_WEB_E2E_DB="${E2E_DB}"
    export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="${E2E_DSN}"
    return 0
  fi

  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "CREATE DATABASE \"${E2E_DB}\";" >/dev/null

  local -a migrate_command=()
  resolve_runtime_command migrate_command "migration" "${MIGRATE_BIN}"

  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY__APPLICATION__PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
  CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="${E2E_DSN}" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
    "${migrate_command[@]}" up

  export CARTULARY_WEB_E2E_DB="${E2E_DB}"
  export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="${E2E_DSN}"
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
      write_startup_diagnostics "fail" "backend_readiness" "infra" "service_start_error" "backend exited before readiness" || true
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if port_owned_by_process_group "${BACKEND_PORT}" "${SERVER_PGID}" && identity_pid="$(probe_backend_identity 2>/dev/null)"; then
      if [[ -n "${SERVER_PGID:-}" ]] && ! process_group_running "${SERVER_PGID}" >/dev/null 2>&1; then
        echo "backend exited immediately after readiness identity probe" >&2
        write_startup_diagnostics "fail" "backend_readiness" "infra" "service_start_error" "backend exited immediately after readiness identity probe" || true
        cat "${SERVER_LOG}" >&2 || true
        return 1
      fi
      BACKEND_IDENTITY_STATUS="pass"
      BACKEND_IDENTITY_SERVER_PID="${identity_pid}"
      BACKEND_READY_AT="$(phase_now_utc)"
      write_stack_metadata
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
  for _ in $(seq 1 240); do
    if exit_for_requested_shutdown "frontend readiness"; then
      :
    else
      return "$?"
    fi
    if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
      echo "frontend exited before readiness" >&2
      write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_start_error" "frontend exited before readiness" || true
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
      FRONTEND_OWNERSHIP_STATUS="pass"
      FRONTEND_READY_AT="$(phase_now_utc)"
      write_stack_metadata
      write_startup_diagnostics "pass" "frontend_readiness" "" "" "frontend ready at ${PUBLIC_ORIGIN}" || true
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

startup_diagnostic_failure_reason() {
  local node_bin="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"

  if [[ -z "${STARTUP_DIAGNOSTIC_FILE}" || ! -f "${STARTUP_DIAGNOSTIC_FILE}" ]]; then
    return 1
  fi
  if [[ ! -x "${node_bin}" ]]; then
    node_bin="node"
  fi

  "${node_bin}" - "${STARTUP_DIAGNOSTIC_FILE}" <<'EOF'
const fs = require("node:fs");

try {
  const payload = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
  if (payload && typeof payload.failure_reason === "string") {
    process.stdout.write(`${payload.failure_reason}\n`);
    process.exit(0);
  }
} catch {
}
process.exit(1);
EOF
}

frontend_resource_conflict_retry_allowed() {
  local reason=""

  if ! using_test_services_stack; then
    return 1
  fi
  if [[ "${FRONTEND_PORT_CONFIGURED}" -eq 1 ]]; then
    return 1
  fi
  reason="$(startup_diagnostic_failure_reason 2>/dev/null || true)"
  [[ "${reason}" == "resource_conflict" ]]
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
  write_stack_metadata
}

retry_frontend_preview_port() {
  local previous_port="${FRONTEND_PORT}"

  if [[ -n "${VITE_PGID:-}" ]]; then
    stop_process_group "${VITE_PGID}" || true
    VITE_PGID=""
  fi
  allocate_available_port FRONTEND_PORT "frontend" "" "${BACKEND_PORT},${previous_port}" || return $?
  release_port_lease_for_port "${previous_port}"
  PUBLIC_ORIGIN="http://127.0.0.1:${FRONTEND_PORT}"
  export CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}"
  export CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}"
  WEB_LOG="${TARGET_ARTIFACT_DIR:-${RUNTIME_ROOT_BASE}}/web-${FRONTEND_PORT}.log"
  export CARTULARY_WEB_E2E_WEB_LOG="${WEB_LOG}"
  echo "frontend port ${previous_port} hit a resource conflict; retrying with ${FRONTEND_PORT}" >&2
  write_stack_metadata
}

start_frontend_preview_ready_with_retry() {
  local pnpm_bin="$1"
  local attempt
  local max_attempts=1

  if using_test_services_stack && [[ "${FRONTEND_PORT_CONFIGURED}" -ne 1 ]]; then
    max_attempts="${TEST_SERVICE_FRONTEND_PREVIEW_RETRY_LIMIT}"
  fi

  for attempt in $(seq 1 "${max_attempts}"); do
    start_frontend_preview_process "${pnpm_bin}"
    if browser_wait_frontend_ready; then
      return 0
    fi
    if (( attempt >= max_attempts )) || ! frontend_resource_conflict_retry_allowed; then
      return 1
    fi
    retry_frontend_preview_port || return $?
  done

  return 1
}

browser_verify_frontend_ready() {
  if [[ -n "${VITE_PGID:-}" ]] && ! process_group_running "${VITE_PGID}" >/dev/null 2>&1; then
    echo "frontend exited before backend-ready verification" >&2
    write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_start_error" "frontend exited before backend-ready verification" || true
    cat "${WEB_LOG}" >&2 || true
    return 1
  fi
  if port_owned_by_process_group "${FRONTEND_PORT}" "${VITE_PGID}" && curl -fsS "${PUBLIC_ORIGIN}" >/dev/null 2>&1; then
    FRONTEND_OWNERSHIP_STATUS="pass"
    FRONTEND_READY_AT="${FRONTEND_READY_AT:-$(phase_now_utc)}"
    write_stack_metadata
    write_startup_diagnostics "pass" "frontend_readiness" "" "" "frontend ready at ${PUBLIC_ORIGIN}" || true
    return 0
  fi

  echo "frontend owned listener was not ready during backend-ready verification at ${PUBLIC_ORIGIN}" >&2
  write_startup_diagnostics "fail" "frontend_readiness" "infra" "service_readiness_timeout" "frontend owned listener was not ready during backend-ready verification at ${PUBLIC_ORIGIN}" || true
  print_port_diagnostics "${FRONTEND_PORT}" "frontend"
  cat "${WEB_LOG}" >&2 || true
  return 1
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

  prepare_runtime_profile

  if [[ "${SESSION_MODE}" == "stop" ]]; then
    stop_session
    return $?
  fi

  prepare_runtime_root

  trap on_exit EXIT
  lifecycle_reset_shutdown_state
  lifecycle_install_signal_traps

  run_timing_span "setup" "browser-e2e frontend toolchain" \
    env -u CARTULARY_TEST_RUN_ID -u CARTULARY_TEST_TARGET MAKEFLAGS= CARTULARY_FRONTEND_TOOLCHAIN_QUIET=1 CARTULARY_SUPPRESS_CHILD_SUCCESS=1 make -s -C "${ROOT_DIR}" --no-print-directory frontend-toolchain
  local pnpm_bin="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
  if [[ ! -x "${pnpm_bin}" ]]; then
    echo "repo-local pnpm was not found at ${pnpm_bin}; run make frontend-toolchain" >&2
    return 1
  fi

  CARTULARY_PHASE_TIMING_BUCKET=setup run_phase_command "browser-e2e allocate ports" resolve_owned_stack_ports
  CARTULARY_PHASE_TIMING_BUCKET=setup run_phase_command "browser-e2e prepare test route token" prepare_test_route_token
  run_timing_span "setup" "browser-e2e write stack metadata" write_stack_metadata
  CARTULARY_PHASE_TIMING_BUCKET=frontend_startup run_phase_command "browser-e2e validate frontend preview artifact" require_frontend_preview_artifacts
  CARTULARY_PHASE_TIMING_BUCKET=frontend_startup run_phase_command "browser-e2e startup frontend ready" start_frontend_preview_ready_with_retry "${pnpm_bin}"

  CARTULARY_PHASE_TIMING_BUCKET=service_wait run_phase_command "browser-e2e startup services" browser_start_services
  CARTULARY_PHASE_TIMING_BUCKET=migration run_phase_command "browser-e2e startup database" browser_prepare_database

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

  run_timing_span "server_startup" "browser-e2e start backend process" \
  start_process_group SERVER_PGID "${SERVER_LOG}" \
    env \
    CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
    CARTULARY__APPLICATION__PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
    CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}" \
    CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}" \
    CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
    CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="${E2E_DSN}" \
    CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT="${CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT:-localhost:8333}" \
    CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID="${CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID:-cartulary-local}" \
    CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY="${CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY:-cartulary-local-secret}" \
    CARTULARY_S3_OBJECT_PRIMARY_SECURE="${CARTULARY_S3_OBJECT_PRIMARY_SECURE:-false}" \
    CARTULARY_S3_OBJECT_PRIMARY_BUCKET="${CARTULARY_S3_OBJECT_PRIMARY_BUCKET:-cartulary}" \
    CARTULARY_ENABLE_TEST_ROUTES=1 \
    CARTULARY_TEST_RUNTIME_MARKER=harness-owned \
    CARTULARY_TEST_ROUTE_TOKEN="${TEST_ROUTE_TOKEN}" \
    CARTULARY__ROOTS__BACKUP_STORAGE__PATH="${RUNTIME_ROOT_BASE}/backup-storage" \
    CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH="${RUNTIME_ROOT_BASE}/reference-pack-storage" \
    CARTULARY__ROOTS__TEMPORARY_WORK__PATH="${RUNTIME_ROOT_BASE}/temporary-work" \
    CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH="${RUNTIME_ROOT_BASE}/export-outputs" \
    "${runtime_profile_env[@]}" \
    GOCACHE=/tmp/cartulary-go-build \
    GOMODCACHE=/tmp/cartulary-go-mod \
    "${backend_listen_command[@]}"
  write_stack_metadata

  CARTULARY_PHASE_TIMING_BUCKET=server_startup run_phase_command "browser-e2e startup backend ready" browser_wait_backend_ready
  CARTULARY_PHASE_TIMING_BUCKET=frontend_startup run_phase_command "browser-e2e verify frontend ready" browser_verify_frontend_ready

  if [[ "${SESSION_MODE}" == "start" ]]; then
    run_timing_span "setup" "browser-e2e write session lease" write_session_files
    release_process_group_monitor "${SERVER_PGID}"
    release_process_group_monitor "${VITE_PGID}"
    release_port_leases || true
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
