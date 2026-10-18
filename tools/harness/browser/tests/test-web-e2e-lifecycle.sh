#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
START_SCRIPT="$ROOT_DIR/tools/harness/browser/start-web-e2e.sh"
RESET_SCRIPT="$ROOT_DIR/tools/harness/browser/reset-web-e2e-stack.sh"
ATTACH_SCRIPT="$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"
EVIDENCE_HELPER="$ROOT_DIR/tools/harness/browser/browser-session-evidence.mjs"
PLAYWRIGHT_CONFIG="$ROOT_DIR/apps/web/playwright.shared.config.ts"
NODE_BIN="${NODE_BIN:-$ROOT_DIR/.cache/cartulary/node-runtime/bin/node}"
cleanup_pids=()

fail() {
  printf 'test-web-e2e-lifecycle: %s\n' "$*" >&2
  exit 1
}

assert_file_contains() {
  local file="$1"
  local value="$2"
  local label="$3"
  grep -Fq "$value" "$file" || fail "$label: expected $file to contain [$value]"
}

assert_file_not_contains() {
  local file="$1"
  local value="$2"
  local label="$3"
  if grep -Fq "$value" "$file"; then
    fail "$label: expected $file to omit [$value]"
  fi
}

assert_text_contains() {
  local text="$1"
  local value="$2"
  local label="$3"
  [[ "$text" == *"$value"* ]] || fail "$label: expected output to contain [$value]"
}

assert_json() {
  local file="$1"
  local expression="$2"
  local label="$3"
  # shellcheck disable=SC2016
  "$NODE_BIN" -e \
    'const fs=require("node:fs"); const value=JSON.parse(fs.readFileSync(process.argv[1],"utf8")); if (!Function("value", `return (${process.argv[2]})`)(value)) process.exit(1);' \
    "$file" "$expression" || fail "$label"
}

# shellcheck disable=SC2329
fixture_cleanup() {
  local pid
  for pid in "${cleanup_pids[@]}"; do
    kill -TERM "$pid" >/dev/null 2>&1 || true
    wait "$pid" >/dev/null 2>&1 || true
  done
  rm -rf "$tmp_dir"
}

[[ -x "$NODE_BIN" ]] || NODE_BIN=node
tmp_dir="$(mktemp -d)"
trap fixture_cleanup EXIT

# The managed lifecycle is session-owned and has no development-stack route.
assert_file_contains "$START_SCRIPT" 'CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID' "authenticated managed suite guard"
assert_file_not_contains "$START_SCRIPT" 'CARTULARY_TEST_SERVICES_ACTIVE' "ambient active state removed"
# shellcheck disable=SC2016
assert_file_contains "$START_SCRIPT" '_shared/test-services/${SUITE_ID}/browser-sessions/${BROWSER_SESSION_ID}' "session artifact cardinality"
assert_file_contains "$START_SCRIPT" 'finalize startup diagnostics' "terminal diagnostic precedes publication"
assert_file_contains "$START_SCRIPT" 'publish immutable v8 stack' "v7 publication"
assert_file_contains "$START_SCRIPT" 'v8 browser stack publication requires bound backend and frontend readiness' "missing readiness fails publication"
assert_file_contains "$START_SCRIPT" 'snapshot_service_scope || return $?' "admission publication failure propagates"
assert_file_contains "$START_SCRIPT" 'verify_stack_publication' "terminal publication verification"
# shellcheck disable=SC2016
assert_file_contains "$START_SCRIPT" 'artifact identity ${BROWSER_SESSION_ID} is already in use' "session artifact collision rejection"
# shellcheck disable=SC2016
assert_file_not_contains "$START_SCRIPT" 'rm -f \
    "${STACK_ENV_FILE}" \
    "${STACK_JSON_FILE}"' "immutable public session evidence is never cleared"
# shellcheck disable=SC2016
assert_file_contains "$START_SCRIPT" 'vite preview --host 127.0.0.1 --port "${FRONTEND_PORT}" --strictPort' "strict preview"
assert_file_contains "$START_SCRIPT" 'TEST_SERVICE_FRONTEND_PORT_START=19000' "service-backed frontend range starts below the default ephemeral range"
assert_file_contains "$START_SCRIPT" 'TEST_SERVICE_FRONTEND_PORT_END=19199' "service-backed frontend range ends below the default ephemeral range"
assert_file_not_contains "$START_SCRIPT" 'retrying with' "strict-port collision retry"
assert_file_not_contains "$START_SCRIPT" 'dev-services.sh' "no development fallback"
assert_file_not_contains "$START_SCRIPT" '127.0.0.1:8333' "no development proxy dependency"
assert_file_not_contains "$START_SCRIPT" 'source: "standalone"' "no standalone database evidence"
assert_file_not_contains "$PLAYWRIGHT_CONFIG" "webServer:" "canonical Playwright config has no webServer"
assert_file_not_contains "$PLAYWRIGHT_CONFIG" "reuseExistingServer" "canonical Playwright config has no listener reuse"
assert_file_contains "$PLAYWRIGHT_CONFIG" 'updateSnapshots: "none"' "ordinary snapshot validation is read-only"
assert_file_contains "$ATTACH_SCRIPT" 'browser-session-evidence.mjs' "Playwright uses the v7 attachment validator"

# Source the adapter only for fail-closed entrypoint checks.
# shellcheck source=tools/harness/browser/start-web-e2e.sh
source "$START_SCRIPT"

if [[ "${1:-}" == --artifact-fixture ]]; then
  # shellcheck source=tools/harness/browser/tests/web-e2e-artifact-fixture.sh
  source "$ROOT_DIR/tools/harness/browser/tests/web-e2e-artifact-fixture.sh"
  exit 0
fi

(
  # Exercise failure under a conditional, where shell errexit cannot help.
  assert_port_free() { return 0; }
  using_test_services_stack() { return 0; }
  require_test_services_bin() { return 0; }
  write_startup_diagnostics() { printf '%s\n' "$*" >"$tmp_dir/prepare-failure"; }
  snapshot_service_scope() { touch "$tmp_dir/incorrect-admission"; }
  record_startup_event() { touch "$tmp_dir/incorrect-ready"; }
  TEST_SERVICES_BIN="$tmp_dir/failed-preparation"
  TEST_SERVICES_ENV_FILE="$tmp_dir/unpublished.env"
  TEST_SERVICES_METADATA_FILE="$tmp_dir/unpublished.json"
  # This isolated prerequisite fixture intentionally leaves the parent environment unchanged.
  # shellcheck disable=SC2030
  CARTULARY_PGTEST_TEMPLATE_DB=template
  printf '#!/bin/sh\nexit 37\n' >"$TEST_SERVICES_BIN"
  chmod 700 "$TEST_SERVICES_BIN"
  printf 'touch "%s"\n' "$tmp_dir/incorrect-env-read" >"$TEST_SERVICES_ENV_FILE"
  preparation_status=0
  browser_prepare_database || preparation_status=$?
  [[ "$preparation_status" == 37 ]] || fail "database prerequisite failure was replaced"
  [[ ! -e "$tmp_dir/incorrect-env-read" && ! -e "$tmp_dir/incorrect-admission" && ! -e "$tmp_dir/incorrect-ready" ]] || fail "database preparation consumed dependent state after failure"
  assert_file_contains "$tmp_dir/prepare-failure" 'database_preparation infra service_start_error' "first preparation failure"
  printf '#!/bin/sh\nexit 0\n' >"$TEST_SERVICES_BIN"
  preparation_status=0
  browser_prepare_database || preparation_status=$?
  [[ "$preparation_status" == 1 ]] || fail "missing metadata was admitted"
)

assert_stack_publication_failure() {
  local label="$1"
  local backend_ready_at="$2"
  local frontend_ready_at="$3"
  local diagnostic_state="$4"
  local expected_diagnostic="$5"
  local case_root="$tmp_dir/publication-$label"
  local public_root="$case_root/public"
  local runtime_root="$case_root/private-runtime"
  local terminal_diagnostic="$case_root/terminal-startup.json"
  local output=""
  local status=0

  mkdir -p "$public_root" "$runtime_root"
  if [[ "$diagnostic_state" == "present" ]]; then
    printf '%s\n' '{"schema_id":"cartulary.browser_startup_diagnostics.v3","status":"ready"}' >"$terminal_diagnostic"
    chmod 600 "$terminal_diagnostic"
  fi

  set +e
  output="$(
    (
      set +e
      BACKEND_READY_AT="$backend_ready_at"
      FRONTEND_READY_AT="$frontend_ready_at"
      STARTUP_DIAGNOSTIC_FILE="$terminal_diagnostic"
      TARGET_ARTIFACT_DIR="$public_root"
      # Publication guards are intentionally isolated in this subshell.
      # shellcheck disable=SC2030
      STACK_ENV_FILE="$case_root/stack.env"
      STACK_JSON_FILE="$public_root/stack.json"
      STACK_LEASE_FILE="$public_root/stack-lease.json"
      SERVICE_ADMISSION_FILE="$public_root/service-admission.json"
      SESSION_ENV_FILE="$case_root/session.env"
      SESSION_LEASE_FILE="$case_root/session-lease.json"
      RUNTIME_ROOT_BASE="$runtime_root"
      PLAYWRIGHT_STATE_DIR="$runtime_root/playwright-state"
      TEST_ROUTE_TOKEN_FILE="$runtime_root/test-route-token"
      TEST_SERVICES_ENV_FILE="$runtime_root/test-services.env"
      TEST_SERVICES_METADATA_FILE=""
      TEST_SERVICES_BIN=""
      SERVER_PGID=""
      VITE_PGID=""
      CHILD_PGID=""
      BACKEND_PORT=39881
      FRONTEND_PORT=39882
      PORT_LEASE_DIRS=()
      KEEP_RUNTIME_ROOT=0
      cleanup_done=0
      write_stack_metadata
      publication_status=$?
      if ! cleanup; then
        echo "publication guard cleanup failed" >&2
        exit 99
      fi
      exit "$publication_status"
    ) 2>&1
  )"
  status=$?
  set -e

  [[ "$status" -eq 1 ]] || fail "$label: expected publication status 1, got $status"
  assert_text_contains "$output" "$expected_diagnostic" "$label bounded diagnostic"
  for unpublished in \
    "$public_root/stack.json" \
    "$public_root/stack-lease.json" \
    "$public_root/service-admission.json" \
    "$case_root/stack.env" \
    "$case_root/session.env" \
    "$case_root/session-lease.json"; do
    [[ ! -e "$unpublished" ]] || fail "$label: unexpectedly published $unpublished"
  done
  [[ ! -e "$runtime_root" ]] || fail "$label: normal cleanup did not remove the private runtime"
}

assert_stack_publication_failure \
  "missing-backend-readiness" "" "2026-08-19T12:00:01Z" "present" \
  "v8 browser stack publication requires bound backend and frontend readiness"
assert_stack_publication_failure \
  "missing-frontend-readiness" "2026-08-19T12:00:00Z" "" "present" \
  "v8 browser stack publication requires bound backend and frontend readiness"
assert_stack_publication_failure \
  "missing-both-readiness" "" "" "present" \
  "v8 browser stack publication requires bound backend and frontend readiness"
assert_stack_publication_failure \
  "missing-terminal-diagnostic" "2026-08-19T12:00:00Z" "2026-08-19T12:00:01Z" "missing" \
  "v8 browser stack publication requires terminal startup diagnostics"

verification_case_root="$tmp_dir/publication-verification-failure"
mkdir -p "$verification_case_root/private-runtime"
set +e
verification_failure_output="$(
  (
    set +e
    # publish_session_lease invokes this test override indirectly.
    # shellcheck disable=SC2329
    write_session_files() { return 0; }
    verify_stack_publication() {
      echo "browser session publication verification failed" >&2
      return 19
    }
    # Verification failure state is intentionally isolated in this subshell.
    # shellcheck disable=SC2030
    RUNTIME_ROOT_BASE="$verification_case_root/private-runtime"
    SESSION_ENV_FILE="$verification_case_root/session.env"
    SESSION_LEASE_FILE="$verification_case_root/session-lease.json"
    TEST_SERVICES_BIN=""
    TEST_SERVICES_METADATA_FILE=""
    SERVER_PGID=""
    VITE_PGID=""
    CHILD_PGID=""
    BACKEND_PORT=39883
    FRONTEND_PORT=39884
    PORT_LEASE_DIRS=()
    KEEP_RUNTIME_ROOT=0
    cleanup_done=0
    publish_session_lease
    publication_status=$?
    if ! cleanup; then
      echo "publication verification cleanup failed" >&2
      exit 99
    fi
    exit "$publication_status"
  ) 2>&1
)"
verification_failure_status=$?
set -e
[[ "$verification_failure_status" -eq 19 ]] ||
  fail "publication verification failure: expected status 19, got $verification_failure_status"
assert_text_contains \
  "$verification_failure_output" \
  "browser session publication verification failed" \
  "publication verification bounded diagnostic"
for unpublished in \
  "$verification_case_root/session.env" \
  "$verification_case_root/session-lease.json"; do
  [[ ! -e "$unpublished" ]] ||
    fail "publication verification failure: unexpectedly published $unpublished"
done
if [[ -d "$verification_case_root/private-runtime" ]] &&
  find "$verification_case_root/private-runtime" -mindepth 1 -print -quit | grep -q .; then
  fail "publication verification failure: normal cleanup retained private runtime material"
fi

# Parent graph selection and scheduling inputs must not leak into the nested
# readiness Make target, whose public contract does not declare them.
fake_make_dir="$tmp_dir/fake-make"
fake_make_log="$tmp_dir/fake-make.log"
mkdir -p "$fake_make_dir"
cat >"$fake_make_dir/make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
for name in \
  CARTULARY_HARNESS_IDENTITY_PREPARED \
  CARTULARY_TEST_RUN_ID \
  CARTULARY_TEST_TARGET \
  CARTULARY_MAKE_INPUT_SOURCES \
  CARTULARY_HARNESS_CACHE_MODE \
  CARTULARY_HARNESS_CAPACITY_OVERRIDE \
  JSON \
  OWNER \
  ROWS \
  SCOPE \
  PLAYWRIGHT_WORKERS \
  VITEST_MAX_WORKERS; do
  if [[ -n "${!name+x}" ]]; then
    printf 'leaked %s\n' "$name" >&2
    exit 91
  fi
done
printf '%s\n' "$*" >"${CARTULARY_FAKE_MAKE_LOG:?}"
EOF
chmod 700 "$fake_make_dir/make"
lifecycle_parent_run_id="${CARTULARY_TEST_RUN_ID:-}"
export CARTULARY_HARNESS_IDENTITY_PREPARED=1
export CARTULARY_TEST_RUN_ID=parent-run
export CARTULARY_TEST_TARGET=test-slice
export CARTULARY_MAKE_INPUT_SOURCES='CARTULARY_HARNESS_CACHE_MODE=cli OWNER=cli ROWS=cli'
export CARTULARY_HARNESS_CACHE_MODE=off
export CARTULARY_HARNESS_CAPACITY_OVERRIDE=cpu_tokens=1
export JSON=1
export OWNER=module.timeline
export ROWS=module.timeline.measurement.timeline_virtualized_grid_render_latency_961a4ec1d3
export SCOPE=services
export PLAYWRIGHT_WORKERS=1
export VITEST_MAX_WORKERS=1
CARTULARY_FAKE_MAKE_LOG="$fake_make_log" PATH="$fake_make_dir:$PATH" \
  browser_prepare_frontend_toolchain || fail "nested frontend toolchain public-input isolation"
assert_file_contains "$fake_make_log" 'frontend-toolchain' "nested frontend toolchain invocation"
unset \
  CARTULARY_HARNESS_IDENTITY_PREPARED \
  CARTULARY_TEST_RUN_ID \
  CARTULARY_TEST_TARGET \
  CARTULARY_MAKE_INPUT_SOURCES \
  CARTULARY_HARNESS_CACHE_MODE \
  CARTULARY_HARNESS_CAPACITY_OVERRIDE \
  JSON \
  OWNER \
  ROWS \
  SCOPE \
  PLAYWRIGHT_WORKERS \
  VITEST_MAX_WORKERS

export CARTULARY_TEST_RUN_ID="$lifecycle_parent_run_id"

private_failure_log="$tmp_dir/private-failure.log"
private_failure_token=opaque-runtime-token-that-must-not-survive
printf 'startup failed with postgres://user:password@localhost/db and token %s\n' \
  "$private_failure_token" >"$private_failure_log"
TEST_ROUTE_TOKEN="$private_failure_token"
private_failure_message="$(bounded_private_failure_message "backend" "$private_failure_log")"
if [[ "$private_failure_message" == *"postgres://"* ]]; then
  fail "bounded private failure diagnostic retained a DSN"
fi
if [[ "$private_failure_message" == *"$private_failure_token"* ]]; then
  fail "bounded private failure diagnostic retained the test route token"
fi
[[ "$private_failure_message" == backend\ process\ failed\ to\ establish\ a\ session:* ]] ||
  fail "bounded private failure diagnostic lost its safe failure class"
TEST_ROUTE_TOKEN=""

# A launcher may exit before a reporter descendant in the same process group.
# Group liveness and cleanup must follow the whole group, not the leader PID.
descendant_group=""
start_process_group descendant_group "" bash -c 'sleep 120 &'
sleep 0.3
process_group_running "$descendant_group" ||
  fail "process-group liveness lost a live descendant after leader exit"
stop_process_group "$descendant_group" ||
  fail "descendant process-group cleanup"
if process_group_running "$descendant_group"; then
  fail "descendant process group remained live after cleanup"
fi

# Port ownership survives the transient allocator shell and becomes stale only
# after the transferred live owner exits.
export CARTULARY_WEB_E2E_PORT_LEASE_ROOT="$tmp_dir/port-leases"
PORT_LEASE_DIRS=()
setsid sleep 120 &
lease_owner_pid=$!
cleanup_pids+=("$lease_owner_pid")
reserve_port_lease 39001 frontend || fail "initial port lease reservation"
transfer_port_lease_for_port 39001 "$lease_owner_pid" ||
  fail "port lease ownership transfer"
if reserve_port_lease 39001 frontend; then
  fail "live transferred port lease must remain exclusive"
fi
kill -TERM "$lease_owner_pid" >/dev/null 2>&1 || true
wait "$lease_owner_pid" >/dev/null 2>&1 || true
remove_stale_port_lease "$(port_lease_dir 39001)" ||
  fail "dead transferred port lease must become reclaimable"
[[ ! -d "$(port_lease_dir 39001)" ]] ||
  fail "dead transferred port lease was not removed"

CARTULARY_BROWSER_SERVICE_REQUIREMENT=test-services
SUITE_ID=suite-test
BROWSER_SESSION_ID=session-test
if prepare_runtime_root >/dev/null 2>&1; then
  fail "managed browser runtime must reject missing suite-runtime proof"
fi
if browser_start_services >/dev/null 2>&1; then
  fail "browser lifecycle must reject shared development services"
fi

# The private browser runtime is admitted only with an exact external suite
# ownership marker and never projects its env/lease/log paths below results.
security_results_root="$tmp_dir/security-results"
security_run_id="security-run"
security_suite_root="$tmp_dir/security-suite-runtime"
security_session_root="$security_suite_root/browser-stack-leases/security-session"
mkdir -p "$security_results_root/$security_run_id" "$security_session_root"
chmod 700 "$security_suite_root" "$security_session_root"
# Valid live-owner admission is exercised below through the native owner.
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$tmp_dir" <<'JS'
import assert from "node:assert/strict";
import path from "node:path";
import { mkdirSync } from "node:fs";
import { createRuntime } from "./tools/harness/runtime/session-client.mjs";
import { spawnSync } from "./tools/harness/workspace/child-process.mjs";
const [root, scratch] = process.argv.slice(2);
const runID = "live-security", runRoot = path.join(scratch, "live-results", runID);
mkdirSync(runRoot, {recursive:true, mode:0o700});
const runtime = createRuntime({repoRoot:root, runRoot, runID});
try {
  const privateSession = runtime.privatePath("browser-stack-leases", "security-session");
  mkdirSync(privateSession, {recursive:true, mode:0o700});
  const env = {...process.env, CARTULARY_TEST_RESULTS_DIR:path.dirname(runRoot), CARTULARY_TEST_RUN_ID:runID,
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:runtime.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:runID, CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT:privateSession,
    CARTULARY_TEST_SUITE_ID:"security-suite", CARTULARY_BROWSER_SESSION_GROUP:"security-session", CARTULARY_BROWSER_SERVICE_REQUIREMENT:"test-services"};
  const script = String.raw`set -euo pipefail
source tools/harness/browser/start-web-e2e.sh
prepare_runtime_root
[[ "$RUNTIME_ROOT_BASE" == "$PRIVATE_SESSION_ROOT/runtime-root" ]]
[[ "$STACK_ENV_FILE" == "$PRIVATE_SESSION_ROOT/stack.env" ]]
if prepare_runtime_root >/dev/null 2>&1; then exit 92; fi
if find "$RUN_ROOT" -type f \( -name '*.env' -o -name '*.lease' -o -name '*.dsn' \) -print -quit | grep -q .; then exit 93; fi
`;
  const child = spawnSync("bash", ["-c", script], {cwd:root,env,encoding:"utf8"});
  assert.equal(child.status, 0, child.stderr);
} finally { runtime.close(); }
JS

contained_suite_root="$security_results_root/$security_run_id/contained-runtime"
mkdir -p "$contained_suite_root/browser-stack-leases/contained-session"
chmod 700 "$contained_suite_root" "$contained_suite_root/browser-stack-leases/contained-session"
export CARTULARY_HARNESS_SUITE_RUNTIME_ROOT="$contained_suite_root"
PRIVATE_SESSION_ROOT="$contained_suite_root/browser-stack-leases/contained-session"
if prepare_runtime_root >/dev/null 2>&1; then
  fail "browser runtime must reject result-root containment"
fi

# Keep the native owner alive for the complete synthetic artifact chain.
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$tmp_dir" <<'JS'
import assert from "node:assert/strict";
import path from "node:path";
import { mkdirSync } from "node:fs";
import { createRuntime } from "./tools/harness/runtime/session-client.mjs";
import { spawnSync } from "./tools/harness/workspace/child-process.mjs";
const [root, scratch] = process.argv.slice(2);
const runID = "artifact-fixture", runRoot = path.join(scratch, "artifact-results", runID);
mkdirSync(runRoot, {recursive:true, mode:0o700});
const runtime = createRuntime({repoRoot:root, runRoot, runID});
try {
  const child = spawnSync("bash", ["tools/harness/browser/tests/test-web-e2e-lifecycle.sh", "--artifact-fixture"], {
    cwd:root, encoding:"utf8", env:{...process.env, ...runtime.producerEnvironment(),
      CARTULARY_TEST_RESULTS_DIR:path.dirname(runRoot), CARTULARY_TEST_RUN_ID:runID,
      CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:runtime.root, CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:runtime.leaseID,
      CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:runID,
      CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT:runtime.privatePath("browser-stack-leases", "session-default")}
  });
  assert.equal(child.status, 0, child.stderr);
} finally { runtime.close(); }
JS

printf '%s\n' "test-web-e2e-lifecycle: pass"

"$NODE_BIN" "$ROOT_DIR/tools/harness/browser/tests/test-design-review.mjs"
