#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
START_SCRIPT="$ROOT_DIR/tools/harness/browser/start-web-e2e.sh"
RESET_SCRIPT="$ROOT_DIR/tools/harness/browser/reset-web-e2e-stack.sh"
ATTACH_SCRIPT="$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"
EVIDENCE_HELPER="$ROOT_DIR/tools/harness/browser/browser-session-evidence.mjs"
PLAYWRIGHT_CONFIG="$ROOT_DIR/apps/web/playwright.shared.config.ts"
NODE_BIN="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
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
cleanup() {
  local pid
  for pid in "${cleanup_pids[@]}"; do
    kill -TERM "$pid" >/dev/null 2>&1 || true
    wait "$pid" >/dev/null 2>&1 || true
  done
  rm -rf "$tmp_dir"
}

[[ -x "$NODE_BIN" ]] || NODE_BIN=node
tmp_dir="$(mktemp -d)"
trap cleanup EXIT

# The managed lifecycle is session-owned and has no development-stack route.
assert_file_contains "$START_SCRIPT" 'CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID' "authenticated managed suite guard"
assert_file_not_contains "$START_SCRIPT" 'CARTULARY_TEST_SERVICES_ACTIVE' "ambient active state removed"
# shellcheck disable=SC2016
assert_file_contains "$START_SCRIPT" '_shared/test-services/${SUITE_ID}/browser-sessions/${BROWSER_SESSION_ID}' "session artifact cardinality"
assert_file_contains "$START_SCRIPT" 'finalize startup diagnostics' "terminal diagnostic precedes publication"
assert_file_contains "$START_SCRIPT" 'publish immutable v6 stack' "v6 publication"
assert_file_contains "$START_SCRIPT" 'v6 browser stack publication requires bound backend and frontend readiness' "missing readiness fails publication"
# shellcheck disable=SC2016
assert_file_not_contains "$START_SCRIPT" 'if [[ -z "${BACKEND_READY_AT}" || -z "${FRONTEND_READY_AT}" ]]; then
    return 0' "missing readiness cannot silently publish"
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
assert_file_contains "$ATTACH_SCRIPT" 'browser-session-evidence.mjs' "Playwright uses the v4 attachment validator"

# Source the adapter only for fail-closed entrypoint checks.
# shellcheck source=tools/harness/browser/start-web-e2e.sh
source "$START_SCRIPT"

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
  SERVICE_BACKED_ONLY \
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
export CARTULARY_HARNESS_IDENTITY_PREPARED=1
export CARTULARY_TEST_RUN_ID=parent-run
export CARTULARY_TEST_TARGET=test-slice
export CARTULARY_MAKE_INPUT_SOURCES='CARTULARY_HARNESS_CACHE_MODE=cli OWNER=cli ROWS=cli'
export CARTULARY_HARNESS_CACHE_MODE=off
export CARTULARY_HARNESS_CAPACITY_OVERRIDE=cpu_tokens=1
export JSON=1
export OWNER=module.timeline
export ROWS=module.timeline.measurement.timeline_virtualized_grid_render_latency_961a4ec1d3
export SERVICE_BACKED_ONLY=1
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
  SERVICE_BACKED_ONLY \
  PLAYWRIGHT_WORKERS \
  VITEST_MAX_WORKERS

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
security_lease_id="00000000-0000-4000-8000-000000000002"
printf '{"schema_id":"cartulary.harness_suite_runtime_owner.v1","lease_id":"%s","run_id":"%s","owner_uid":%s,"created_at":"2026-08-14T00:00:00Z"}\n' \
  "$security_lease_id" "$security_run_id" "$(id -u)" >"$security_suite_root/runtime-owner.json"
chmod 600 "$security_suite_root/runtime-owner.json"
export CARTULARY_TEST_SERVICES_CALL_MODE=owned
export CARTULARY_TEST_RESULTS_DIR="$security_results_root"
export CARTULARY_TEST_RUN_ID="$security_run_id"
export CARTULARY_HARNESS_SUITE_RUNTIME_ROOT="$security_suite_root"
export CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID="$security_lease_id"
export CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID="$security_run_id"
PRIVATE_SESSION_ROOT="$security_session_root"
BROWSER_SESSION_ID=security-session
prepare_runtime_root || fail "valid external private browser runtime admission"
if prepare_runtime_root >/dev/null 2>&1; then
  fail "browser runtime must reject a reused public session identity"
fi
[[ "$RUNTIME_ROOT_BASE" == "$security_session_root/runtime-root" ]] ||
  fail "browser runtime root must be external to retained results"
[[ "$STACK_ENV_FILE" == "$security_session_root/stack.env" ]] ||
  fail "browser stack environment must remain private"
if find "$security_results_root/$security_run_id" -type f \( -name '*.env' -o -name '*.lease' -o -name '*.dsn' \) -print -quit | grep -q .; then
  fail "private browser runtime filename appeared below retained results"
fi

contained_suite_root="$security_results_root/$security_run_id/contained-runtime"
mkdir -p "$contained_suite_root/browser-stack-leases/contained-session"
chmod 700 "$contained_suite_root" "$contained_suite_root/browser-stack-leases/contained-session"
export CARTULARY_HARNESS_SUITE_RUNTIME_ROOT="$contained_suite_root"
PRIVATE_SESSION_ROOT="$contained_suite_root/browser-stack-leases/contained-session"
if prepare_runtime_root >/dev/null 2>&1; then
  fail "browser runtime must reject result-root containment"
fi

# Build one complete session artifact chain with live process proofs.
results_root="$tmp_dir/results"
run_id="run-test"
suite_id="suite-test"
session_id="session-default"
run_root="$results_root/$run_id"
suite_root="$run_root/_shared/test-services/$suite_id"
session_root="$suite_root/browser-sessions/$session_id"
private_suite_root="$tmp_dir/private-suite-runtime"
private_session_root="$private_suite_root/browser-stack-leases/$session_id"
runtime_root="$private_session_root/runtime-root"
mkdir -p "$session_root" "$private_session_root/logs" "$runtime_root/playwright-state"
chmod 700 "$private_suite_root" "$private_session_root" "$runtime_root" "$runtime_root/playwright-state"
printf '{"schema_id":"cartulary.test_services.scope.v2","target":"browser-e2e","suite_id":"%s","run_id":"%s","artifact_dir":"%s","readiness_generation":"sha256:%s","wrapper":{"owned_count":1,"pass_through_count":0},"preflight":{"docker_ok":true,"reaper_ready":true,"stale_containers_scanned":0,"stale_containers_removed":0,"stale_containers_deferred":0,"ryuk_disabled_for_suite_startup":true},"failures":{},"cleanup":{},"postgres":{"started":true,"startup":{"attempt_count":0,"retry_count":0,"slowest_attempt_duration_ms":0,"final_attempt":0,"final_retryable":false,"final_retry_blocked_by_context":false},"attached_harness_count":1,"created_database_count":1,"migrated_database_count":1,"template_clone_count":1},"object_store":{"started":true,"secure":false,"startup":{"attempt_count":0,"retry_count":0,"slowest_attempt_duration_ms":0,"final_attempt":0,"final_retryable":false,"final_retry_blocked_by_context":false},"attached_harness_count":1,"bucket_create_count":1,"bucket_cleanup_count":0},"browser_e2e":{"retired_fixture_count":0,"cleaned_fixture_count":0,"reclaimed_fixture_count":0},"fixture":{"total_count":2,"total_duration_ms":0,"strategy_aggregate_count":0},"started_services":{"names":["object_store","postgres"]}}\n' \
  "$suite_id" "$run_id" "$suite_root" "$(printf '4%.0s' {1..64})" >"$suite_root/service-scope.json"
printf '{"schema_id":"cartulary.run_manifest.v3","source_digest":"sha256:%s"}\n' \
  "$(printf '3%.0s' {1..64})" >"$run_root/run-manifest.json"
mkdir -p "$private_suite_root/test-services"
chmod 700 "$private_suite_root/test-services"
printf '{"schema_id":"cartulary.test_services.lease.v1","lease_id":"00000000-0000-4000-8000-000000000001","suite_id":"%s","run_id":"%s","cleanup_state":"not_started","resources":[{"kind":"container","service":"postgres","container_id":"postgres-container-proof"},{"kind":"container","service":"object_store","container_id":"object-store-container-proof"}]}\n' \
  "$suite_id" "$run_id" >"$private_suite_root/test-services/service-lease.json"
chmod 600 "$private_suite_root/test-services/service-lease.json"
printf '{"database_name":"ct_web_test","bucket":"ct-web-test"}\n' >"$runtime_root/test-services-web-e2e.json"
printf 'backend\n' >"$private_session_root/logs/server.log"
printf 'frontend\n' >"$private_session_root/logs/web.log"

setsid sleep 120 &
backend_pid=$!
cleanup_pids+=("$backend_pid")
setsid sleep 120 &
frontend_pid=$!
cleanup_pids+=("$frontend_pid")

export CARTULARY_TEST_RESULTS_DIR="$results_root"
export CARTULARY_TEST_RUN_ID="$run_id"
export CARTULARY_TEST_SUITE_ID="$suite_id"
export CARTULARY_BROWSER_SESSION_GROUP="$session_id"
export CARTULARY_BROWSER_RUNTIME_PROFILE_ID=default
export CARTULARY_BROWSER_SERVICE_REQUIREMENT=test-services
export CARTULARY_TEST_SERVICES_CALL_MODE=owned
export CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR="$session_root"
export CARTULARY_WEB_E2E_RUNTIME_ROOT="$runtime_root"
export CARTULARY_PLAYWRIGHT_STATE_DIR="$runtime_root/playwright-state"
export CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID=00000000-0000-4000-8000-000000000001
export CARTULARY_HARNESS_SUITE_RUNTIME_ROOT="$private_suite_root"
export CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID="$run_id"
export CARTULARY_WEB_E2E_SERVER_LOG="$private_session_root/logs/server.log"
export CARTULARY_WEB_E2E_WEB_LOG="$private_session_root/logs/web.log"
export CARTULARY_WEB_E2E_API_ORIGIN=http://127.0.0.1:38080
export CARTULARY_WEB_E2E_PUBLIC_ORIGIN=http://127.0.0.1:34173
export CARTULARY_WEB_E2E_BACKEND_PORT=38080
export CARTULARY_WEB_E2E_FRONTEND_PORT=34173
export CARTULARY_WEB_E2E_SERVER_PGID="$backend_pid"
export CARTULARY_WEB_E2E_VITE_PGID="$frontend_pid"
export CARTULARY_WEB_E2E_BACKEND_IDENTITY_SERVER_PID="$backend_pid"
export CARTULARY_WEB_E2E_BACKEND_READY_AT=2026-07-27T12:00:00Z
export CARTULARY_WEB_E2E_FRONTEND_READY_AT=2026-07-27T12:00:01Z
CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT="sha256:$(printf '1%.0s' {1..64})"
export CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT
export CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE="$runtime_root/test-services-web-e2e.json"
export CARTULARY_PGTEST_TEMPLATE_DB=ct_suite_template
CARTULARY_PGTEST_SCHEMA_HASH="sha256:$(printf '2%.0s' {1..64})"
export CARTULARY_PGTEST_SCHEMA_HASH
export CARTULARY_S3TEST_ENDPOINT=127.0.0.1:39000
export CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT=127.0.0.1:39000
export CARTULARY_S3_OBJECT_PRIMARY_SECURE=false
export CARTULARY_S3_OBJECT_PRIMARY_BUCKET=ct-web-test

"$NODE_BIN" "$EVIDENCE_HELPER" event initializing "initializing test session"
"$NODE_BIN" "$EVIDENCE_HELPER" event service_attached "attached exact suite"
"$NODE_BIN" "$EVIDENCE_HELPER" write-service-admission
"$NODE_BIN" "$EVIDENCE_HELPER" event fixture_ready "fixture ready"
"$NODE_BIN" "$EVIDENCE_HELPER" event backend_ready "backend ready"
"$NODE_BIN" "$EVIDENCE_HELPER" event frontend_ready "frontend ready"
"$NODE_BIN" "$EVIDENCE_HELPER" terminal ready "session ready"
"$NODE_BIN" "$EVIDENCE_HELPER" lease
stack_file="$("$NODE_BIN" "$EVIDENCE_HELPER" stack)"
export CARTULARY_WEB_E2E_STACK_JSON_FILE="$stack_file"

assert_json "$stack_file" 'value.schema_id === "cartulary.web_e2e_stack.v6"' "v4 schema identity"
assert_json "$stack_file" 'value.suite_id === "suite-test" && value.browser_session_id === "session-default"' "v4 suite/session identity"
assert_json "$stack_file" 'value.postgres_identity.database_name === "ct_web_test" && value.object_store_identity.bucket === "ct-web-test"' "v4 isolated resource identity"
assert_json "$stack_file" 'value.frontend.frontend_command_kind === "vite-preview"' "v4 preview identity"
if grep -Eq 'access_key|secret|postgres://' "$stack_file"; then
  fail "v6 stack must not contain credentials or DSNs"
fi

attachment_exports="$("$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file")"
eval "$attachment_exports"
[[ "$CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED" == "1" ]] ||
  fail "valid v4 attachment must set the admission marker"
[[ "$CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER" == "1" ]] ||
  fail "valid v4 attachment must select external-server lifecycle semantics"
[[ "$CARTULARY_PLAYWRIGHT_STATE_DIR" == "$runtime_root/playwright-state" ]] ||
  fail "Playwright state must be scoped to the attached browser session"
attachment_json="$("$NODE_BIN" "$EVIDENCE_HELPER" attach-json "$stack_file")"
printf '%s\n' "$attachment_json" |
  "$NODE_BIN" -e '
    let input = "";
    process.stdin.on("data", (chunk) => { input += chunk; });
    process.stdin.on("end", () => {
      const value = JSON.parse(input);
      if (
        value.CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED !== "1" ||
        value.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER !== "1"
      ) {
        process.exit(1);
      }
    });
  ' ||
  fail "JSON attachment validation must return the closed admission environment"

# The Playwright-facing adapter accepts only the exact validated session.
# shellcheck source=tools/harness/browser/playwright-owned-stack.sh
source "$ATTACH_SCRIPT"
resolve_playwright_owned_stack_env "$ROOT_DIR"
printf '%s\n' "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" |
  grep -Fq 'CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED=1' ||
  fail "Playwright environment must carry the v4 admission marker"
printf '%s\n' "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" |
  grep -Fq "CARTULARY_PLAYWRIGHT_STATE_DIR=$runtime_root/playwright-state" ||
  fail "Playwright environment must carry session-scoped shared state"

if CARTULARY_BROWSER_RUNTIME_PROFILE_ID=network_flow_claimed \
  "$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file" >/dev/null 2>&1; then
  fail "profile-mismatched v4 attachment must fail"
fi
if "$NODE_BIN" "$EVIDENCE_HELPER" stack >/dev/null 2>&1; then
  fail "v6 stack publication must be immutable"
fi
printf '\n' >>"$session_root/startup-diagnostics.json"
if "$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file" >/dev/null 2>&1; then
  fail "mutated startup diagnostics must invalidate attachment"
fi

# A failed terminal state cannot regress to ready or accept more events.
failed_session_id="session-failed"
failed_session_root="$suite_root/browser-sessions/$failed_session_id"
mkdir -p "$failed_session_root"
export CARTULARY_BROWSER_SESSION_GROUP="$failed_session_id"
export CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR="$failed_session_root"
"$NODE_BIN" "$EVIDENCE_HELPER" event initializing "initializing failed session"
"$NODE_BIN" "$EVIDENCE_HELPER" terminal failed "configuration rejected" config configuration_error
if "$NODE_BIN" "$EVIDENCE_HELPER" terminal ready "must not regress" >/dev/null 2>&1; then
  fail "failed terminal diagnostics must be immutable"
fi
if "$NODE_BIN" "$EVIDENCE_HELPER" event service_attached "must not append" >/dev/null 2>&1; then
  fail "terminal diagnostics must close the event stream"
fi
assert_json "$failed_session_root/startup-diagnostics.json" \
  'value.schema_id === "cartulary.browser_startup_diagnostics.v2" && value.status === "failed"' \
  "failed v2 terminal diagnostic"

# The reset entrypoint owns every retained reset-boundary writer. Prove a
# caller's ambient 022 cannot weaken response, normalized-data, status, or
# Playwright-state evidence.
reset_fake_bin="$tmp_dir/reset-fake-bin"
reset_results_root="$tmp_dir/reset-results"
reset_runtime_root="$tmp_dir/reset-runtime"
reset_state_root="$tmp_dir/reset-state"
mkdir -p "$reset_fake_bin" "$reset_runtime_root" "$reset_state_root"
cat >"$reset_fake_bin/testservices" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ -n "${CARTULARY_RESET_FAKE_ARGS:-}" ]]; then
  printf '%s\n' "$@" >"$CARTULARY_RESET_FAKE_ARGS"
fi
exit 0
EOF
cat >"$reset_fake_bin/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
output_file=""
while [[ "$#" -gt 0 ]]; do
  if [[ "$1" == "-o" ]]; then
    output_file="$2"
    shift 2
    continue
  fi
  shift
done
[[ -n "$output_file" ]]
printf '%s\n' '{"data":{"schema_id":"cartulary.test.runtime_reset.v1","reset_id":"ambient-022","tables_reset":["records"],"mutable_table_count":1,"object_count_removed":0,"object_count_after":0,"migration_metadata_preserved":true,"bootstrap_admin_restored":true,"partial_failure":false,"post_reset_counts":{"active_deployment_admins":1,"bootstrap_markers":1,"incidents":0,"records":0,"user_sessions":0,"route_idempotency":0}}}' >"$output_file"
printf '200'
EOF
chmod 700 "$reset_fake_bin/testservices" "$reset_fake_bin/curl"
(
  umask 022
  unset CARTULARY_HARNESS_IDENTITY_PREPARED
  CARTULARY_TEST_TARGET=adhoc \
    CARTULARY_TEST_RESULTS_DIR="$reset_results_root" \
    CARTULARY_TEST_RUN_ID=ambient-reset \
    CARTULARY_TEST_ROUTE_TOKEN=opaque-reset-token \
    CARTULARY_TEST_SERVICES_BIN="$reset_fake_bin/testservices" \
    CARTULARY_WEB_E2E_RUNTIME_ROOT="$reset_runtime_root" \
    CARTULARY_PLAYWRIGHT_STATE_DIR="$reset_state_root" \
    NODE_BIN="$NODE_BIN" \
    PATH="$reset_fake_bin:$PATH" \
    "$RESET_SCRIPT" --label ambient-022
)
reset_support_root="$reset_results_root/ambient-reset/adhoc/reset-boundary"
for reset_artifact in \
  "$reset_support_root/ambient-022.json" \
  "$reset_support_root/ambient-022.data.json" \
  "$reset_support_root/ambient-022.status" \
  "$reset_support_root/ambient-022.state-reset"; do
  [[ -f "$reset_artifact" ]] || fail "reset ambient-022 artifact missing: $reset_artifact"
  [[ "$(stat -c '%a' "$reset_artifact")" == "600" ]] ||
    fail "reset ambient-022 artifact must be 0600: $reset_artifact"
done

reset_args_file="$tmp_dir/reset-renew-args"
reset_metadata_file="$reset_runtime_root/test-services-web-e2e.json"
printf '%s\n' '{}' >"$reset_metadata_file"
CARTULARY_RESET_FAKE_ARGS="$reset_args_file" \
  CARTULARY_TEST_TARGET=adhoc \
  CARTULARY_TEST_RESULTS_DIR="$reset_results_root" \
  CARTULARY_TEST_RUN_ID=ambient-reset \
  CARTULARY_TEST_ROUTE_TOKEN=opaque-reset-token \
  CARTULARY_TEST_SERVICES_BIN="$reset_fake_bin/testservices" \
  CARTULARY_WEB_E2E_RUNTIME_ROOT="$reset_runtime_root" \
  CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE="$reset_metadata_file" \
  CARTULARY_PLAYWRIGHT_STATE_DIR="$reset_state_root" \
  NODE_BIN="$NODE_BIN" \
  PATH="$reset_fake_bin:$PATH" \
  "$RESET_SCRIPT" --label functional-generation-2 --renew-generation 2
if [[ "$(tr '\n' ' ' <"$reset_args_file")" != \
  "renew-web-e2e --credential-root $reset_runtime_root --bootstrap-manifest $ROOT_DIR/configs/dev/bootstrap-admin.json --metadata-file $reset_metadata_file --generation 2 " ]]; then
  fail "functional renewal did not pass exact owned resource identity"
fi

printf '%s\n' "test-web-e2e-lifecycle: pass"
