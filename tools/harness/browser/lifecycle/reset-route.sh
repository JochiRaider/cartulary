#!/usr/bin/env bash
# shellcheck shell=bash
# shellcheck disable=SC2154

web_e2e_reset_route_token() {
  local token="${CARTULARY_TEST_ROUTE_TOKEN:-}"
  if [[ -z "${token}" && -n "${CARTULARY_TEST_ROUTE_TOKEN_FILE:-}" && -f "${CARTULARY_TEST_ROUTE_TOKEN_FILE}" ]]; then
    token="$(tr -d '\r\n' <"${CARTULARY_TEST_ROUTE_TOKEN_FILE}")"
  fi
  if [[ -z "${token}" ]]; then
    echo "CARTULARY_TEST_ROUTE_TOKEN or CARTULARY_TEST_ROUTE_TOKEN_FILE is required for reset" >&2
    return 2
  fi
  printf '%s\n' "${token}"
}

web_e2e_reset_mark_tainted_on_partial_failure() {
  local response_file="$1"
  local taint_marker_file="$2"
  local runtime_root="$3"

  "${NODE_BIN:-node}" - "$response_file" "$taint_marker_file" "$runtime_root" <<'EOF' || true
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
}

web_e2e_reset_mark_tainted() {
  local taint_marker_file="$1"
  local runtime_root="$2"

  printf '%s\n' "partial_failure" >"$taint_marker_file"
  if [[ -n "$runtime_root" ]]; then
    printf '%s\n' "partial_failure" >"${runtime_root%/}/stack.tainted"
  fi
}

web_e2e_reset_database() {
  local root_dir="$1"
  local runtime_root="${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
  local test_services_bin="${CARTULARY_TEST_SERVICES_BIN:-}"

  if [[ -z "$runtime_root" ]]; then
    echo "CARTULARY_WEB_E2E_RUNTIME_ROOT is required for Recovery-purpose reset" >&2
    return 2
  fi
  if [[ -z "$test_services_bin" || ! -x "$test_services_bin" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN must name an executable for Recovery-purpose reset" >&2
    return 2
  fi
  "$test_services_bin" reset-web-e2e \
    --credential-root "$runtime_root" \
    --bootstrap-manifest "$root_dir/configs/dev/bootstrap-admin.json"
}

web_e2e_renew_functional_resources() {
  local root_dir="$1"
  local generation="$2"
  local runtime_root="${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
  local test_services_bin="${CARTULARY_TEST_SERVICES_BIN:-}"
  local metadata_file="${CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE:-}"

  if [[ -z "$runtime_root" ]]; then
    echo "CARTULARY_WEB_E2E_RUNTIME_ROOT is required for functional renewal" >&2
    return 2
  fi
  if [[ -z "$test_services_bin" || ! -x "$test_services_bin" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN must name an executable for functional renewal" >&2
    return 2
  fi
  if [[ -z "$metadata_file" ]]; then
    metadata_file="${runtime_root%/}/test-services-web-e2e.json"
  fi
  "$test_services_bin" renew-web-e2e \
    --credential-root "$runtime_root" \
    --bootstrap-manifest "$root_dir/configs/dev/bootstrap-admin.json" \
    --metadata-file "$metadata_file" \
    --generation "$generation"
}

web_e2e_reset_validate_response() {
  local response_file="$1"
  local data_file="$2"

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
}

web_e2e_reset_validate_schema() {
  local root_dir="$1"
  local data_file="$2"

  "${NODE_BIN:-${root_dir}/tmp/node-runtime/bin/node}" \
    "$root_dir/tools/harness/contract/harness-contract-cli.mjs" \
    validate-schema cartulary.test.runtime_reset.v1 "$data_file"
}

web_e2e_reset_clear_playwright_state() {
  local state_dir="${CARTULARY_PLAYWRIGHT_STATE_DIR:-}"
  local state_marker_file="$1"

  if [[ -n "${state_dir}" ]]; then
    mkdir -p "$state_dir"
    find "$state_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
    printf '%s\n' "$state_dir" >"$state_marker_file"
  fi
}

web_e2e_reset_request() {
  local api_origin="$1"
  local route_origin="$2"
  local token="$3"
  local response_file="$4"

  curl -sS \
    --max-time 35 \
    -X POST \
    -H 'Content-Type: application/json' \
    -H "Origin: ${route_origin}" \
    -H "X-Cartulary-Test-Route-Token: ${token}" \
    -o "$response_file" \
    -w '%{http_code}' \
    "${api_origin}/api/v1/test/runtime/reset"
}

web_e2e_reset_stack() {
  local root_dir="$1"
  local label="$2"
  local renew_generation="${3:-}"
  local support_dir
  local response_file
  local status_file
  local state_marker_file
  local taint_marker_file
  local data_file
  local api_origin
  local route_origin
  local test_route_token
  local status

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
  test_route_token="$(web_e2e_reset_route_token)" || return $?

  if [[ -n "$renew_generation" ]]; then
    web_e2e_renew_functional_resources "$root_dir" "$renew_generation" || return $?
  else
    web_e2e_reset_database "$root_dir" || return $?
  fi
  if ! status="$(web_e2e_reset_request "$api_origin" "$route_origin" "$test_route_token" "$response_file")"; then
    web_e2e_reset_mark_tainted "$taint_marker_file" "${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
    echo "test runtime reset request failed after database reset committed" >&2
    return 1
  fi
  printf '%s\n' "$status" >"$status_file"

  if [[ "$status" != "200" ]]; then
    web_e2e_reset_mark_tainted "$taint_marker_file" "${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
    web_e2e_reset_mark_tainted_on_partial_failure \
      "$response_file" \
      "$taint_marker_file" \
      "${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
    echo "test runtime reset returned HTTP ${status}" >&2
    cat "$response_file" >&2 || true
    return 1
  fi

  web_e2e_reset_validate_response "$response_file" "$data_file"
  web_e2e_reset_validate_schema "$root_dir" "$data_file"
  web_e2e_reset_clear_playwright_state "$state_marker_file"
}
