#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/core/run-phase-common.sh"

label="reset"
if [[ "${1:-}" == "--label" ]]; then
  if [[ -z "${2:-}" ]]; then
    echo "usage: reset-web-e2e-stack.sh [--label <label>]" >&2
    exit 2
  fi
  label="$2"
  shift 2
fi
if [[ "$#" -ne 0 ]]; then
  echo "usage: reset-web-e2e-stack.sh [--label <label>]" >&2
  exit 2
fi

support_dir="$(prepare_target_support_dir reset-boundary)"
response_file="${support_dir}/${label}.json"
status_file="${support_dir}/${label}.status"
state_marker_file="${support_dir}/${label}.state-reset"
taint_marker_file="${support_dir}/${label}.tainted"
data_file="${support_dir}/${label}.data.json"
api_origin="${CARTULARY_WEB_E2E_API_ORIGIN:-http://127.0.0.1:8080}"
api_origin="${api_origin%/}"
route_origin="${CARTULARY_WEB_E2E_PUBLIC_ORIGIN:-$api_origin}"
route_origin="${route_origin%/}"
test_route_token="${CARTULARY_TEST_ROUTE_TOKEN:-}"
if [[ -z "${test_route_token}" && -n "${CARTULARY_TEST_ROUTE_TOKEN_FILE:-}" && -f "${CARTULARY_TEST_ROUTE_TOKEN_FILE}" ]]; then
  test_route_token="$(tr -d '\r\n' <"${CARTULARY_TEST_ROUTE_TOKEN_FILE}")"
fi
if [[ -z "${test_route_token}" ]]; then
  echo "CARTULARY_TEST_ROUTE_TOKEN or CARTULARY_TEST_ROUTE_TOKEN_FILE is required for reset" >&2
  exit 2
fi

status="$(
  curl -sS \
    --max-time 35 \
    -X POST \
    -H 'Content-Type: application/json' \
    -H "Origin: ${route_origin}" \
    -H "X-Cartulary-Test-Route-Token: ${test_route_token}" \
    -o "$response_file" \
    -w '%{http_code}' \
    "${api_origin}/api/v1/test/runtime/reset"
)"
printf '%s\n' "$status" >"$status_file"

if [[ "$status" != "200" ]]; then
  "${NODE_BIN:-node}" - "$response_file" "$taint_marker_file" "${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}" <<'EOF' || true
const fs = require("node:fs");
const responsePath = process.argv[2];
const taintMarker = process.argv[3];
const runtimeRoot = process.argv[4] || "";
let partial = false;
try {
  const envelope = JSON.parse(fs.readFileSync(responsePath, "utf8"));
  partial = envelope?.error?.details?.partial_failure === true;
} catch {}
if (partial) {
  fs.writeFileSync(taintMarker, "partial_failure\n");
  if (runtimeRoot) {
    fs.writeFileSync(`${runtimeRoot.replace(/\/+$/u, "")}/stack.tainted`, "partial_failure\n");
  }
}
EOF
  echo "test runtime reset returned HTTP ${status}" >&2
  cat "$response_file" >&2 || true
  exit 1
fi

"${NODE_BIN:-node}" - "$response_file" "$data_file" <<'EOF'
const fs = require("node:fs");

const responsePath = process.argv[2];
const dataPath = process.argv[3];
const envelope = JSON.parse(fs.readFileSync(responsePath, "utf8"));
const data = envelope.data ?? {};
const counts = data.post_reset_counts ?? {};
const failures = [];

if (data.schema_id !== "cartulary.test.runtime_reset.v1") {
  failures.push(`unexpected schema_id ${data.schema_id}`);
}
if (data.partial_failure !== false) {
  failures.push("partial_failure must be false on successful reset");
}
if (typeof data.reset_id !== "string" || data.reset_id.trim() === "") {
  failures.push("missing reset_id");
}
if (!Array.isArray(data.tables_reset) || data.tables_reset.length === 0) {
  failures.push("tables_reset must be non-empty");
}
if (data.migration_metadata_preserved !== true) {
  failures.push("migration metadata was not preserved");
}
if (data.bootstrap_admin_restored !== true) {
  failures.push("bootstrap admin was not restored");
}
if (data.object_count_after !== 0) {
  failures.push(`object_count_after must be 0, got ${data.object_count_after}`);
}
for (const [key, want] of [
  ["active_deployment_admins", 1],
  ["bootstrap_markers", 1],
  ["incidents", 0],
  ["records", 0],
  ["user_sessions", 0],
  ["route_idempotency", 0],
]) {
  if (counts[key] !== want) {
    failures.push(`post_reset_counts.${key} must be ${want}, got ${counts[key]}`);
  }
}

if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}
fs.writeFileSync(dataPath, `${JSON.stringify(data, null, 2)}\n`);
EOF
"${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" \
  "$ROOT_DIR/tools/harness/core/harness-contract-cli.mjs" \
  validate-schema cartulary.test.runtime_reset.v1 "$data_file"

if [[ -n "${CARTULARY_PLAYWRIGHT_STATE_DIR:-}" ]]; then
  mkdir -p "$CARTULARY_PLAYWRIGHT_STATE_DIR"
  find "$CARTULARY_PLAYWRIGHT_STATE_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
  printf '%s\n' "$CARTULARY_PLAYWRIGHT_STATE_DIR" >"$state_marker_file"
fi
