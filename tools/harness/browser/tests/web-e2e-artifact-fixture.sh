#!/usr/bin/env bash
# Sourced only by test-web-e2e-lifecycle.sh inside a native fixture owner.
# shellcheck disable=SC2154,SC2153
# Build one complete session artifact chain with live process proofs.
results_root="${CARTULARY_TEST_RESULTS_DIR:?}"
run_id="${CARTULARY_TEST_RUN_ID:?}"
suite_id="suite-test"
session_id="session-default"
run_root="$results_root/$run_id"
suite_root="$run_root/_shared/test-services/$suite_id"
session_root="$suite_root/browser-sessions/$session_id"
private_suite_root="${CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:?}"
private_session_root="${CARTULARY_WEB_E2E_PRIVATE_SESSION_ROOT:?}"
runtime_root="$private_session_root/runtime-root"
mkdir -p "$session_root" "$private_session_root/logs" "$runtime_root/playwright-state"
chmod 700 "$private_suite_root" "$private_session_root" "$runtime_root" "$runtime_root/playwright-state"
"${GO:-go}" run ./tools/browserpki --directory "$runtime_root/tls"
cp "$runtime_root/tls/ca.pem" "$runtime_root/tls/trust.pem"
export CARTULARY_WEB_E2E_TLS_ROOT_CERTIFICATE="$runtime_root/tls/trust.pem"
printf '{"schema_id":"cartulary.test_services.scope.v2","target":"browser-e2e","suite_id":"%s","run_id":"%s","artifact_dir":"%s","readiness_generation":"sha256:%s","wrapper":{"owned_count":1,"pass_through_count":0},"preflight":{"docker_ok":true,"reaper_ready":true,"stale_containers_scanned":0,"stale_containers_removed":0,"stale_containers_deferred":0,"ryuk_disabled_for_suite_startup":true},"failures":{},"cleanup":{},"postgres":{"started":true,"startup":{"attempt_count":0,"retry_count":0,"slowest_attempt_duration_ms":0,"final_attempt":0,"final_retryable":false,"final_retry_blocked_by_context":false},"attached_harness_count":1,"created_database_count":1,"migrated_database_count":1,"template_clone_count":1},"object_store":{"started":true,"secure":false,"startup":{"attempt_count":0,"retry_count":0,"slowest_attempt_duration_ms":0,"final_attempt":0,"final_retryable":false,"final_retry_blocked_by_context":false},"attached_harness_count":1,"bucket_create_count":1,"bucket_cleanup_count":0},"browser_e2e":{"retired_fixture_count":0,"cleaned_fixture_count":0,"reclaimed_fixture_count":0},"fixture":{"total_count":2,"total_duration_ms":0,"strategy_aggregate_count":0},"started_services":{"names":["object_store","postgres"]}}\n' \
  "$suite_id" "$run_id" "$suite_root" "$(printf '4%.0s' {1..64})" >"$suite_root/service-scope.json"
printf '{"schema_id":"cartulary.harness_run_manifest.v3","toolchain_digest":"sha256:2222222222222222222222222222222222222222222222222222222222222222","source_digest":"sha256:%s"}\n' \
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
export CARTULARY_BROWSER_STAGE=webserver-backed
export CARTULARY_BROWSER_RUNTIME_PROFILE_ID=default
export CARTULARY_BROWSER_SERVICE_REQUIREMENT=test-services
export CARTULARY_TEST_SERVICES_CALL_MODE=owned
export CARTULARY_WEB_E2E_SESSION_ARTIFACT_DIR="$session_root"
export CARTULARY_WEB_E2E_RUNTIME_ROOT="$runtime_root"
export CARTULARY_PLAYWRIGHT_STATE_DIR="$runtime_root/playwright-state"

export CARTULARY_HARNESS_SUITE_RUNTIME_ROOT="$private_suite_root"
export CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID="$run_id"
export CARTULARY_WEB_E2E_SERVER_LOG="$private_session_root/logs/server.log"
export CARTULARY_WEB_E2E_WEB_LOG="$private_session_root/logs/web.log"
export CARTULARY_WEB_E2E_API_ORIGIN=https://127.0.0.1:38080
export CARTULARY_WEB_E2E_PUBLIC_ORIGIN=https://127.0.0.1:34173
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
# Independent parent fixture; the earlier subshell does not supply its template.
# shellcheck disable=SC2031
export CARTULARY_PGTEST_TEMPLATE_DB=ct_suite_template
CARTULARY_PGTEST_SCHEMA_HASH="sha256:$(printf '2%.0s' {1..64})"
export CARTULARY_PGTEST_SCHEMA_HASH
export CARTULARY_S3TEST_ENDPOINT=127.0.0.1:39000
export CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT=127.0.0.1:39000
export CARTULARY_S3_OBJECT_PRIMARY_SECURE=true
export CARTULARY_S3_OBJECT_PRIMARY_BUCKET=ct-web-test

"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$run_root" "$private_suite_root" "$run_id" <<'JS'
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
const [root, runRoot, privateRoot, runID] = process.argv.slice(2);
const { sealFrontendArtifact } = await import(path.join(root, "tools/harness/readiness/frontend-artifact.mjs"));
const { borrowRuntime } = await import(path.join(root, "tools/harness/runtime/session-client.mjs"));
const runtime = borrowRuntime({repoRoot:root, runRoot});
const staging = runtime.privatePath("frontend-staging");
mkdirSync(staging, { mode: 0o700 });
writeFileSync(path.join(staging, "index.html"), "<main>synthetic frontend</main>");
sealFrontendArtifact({ repoRoot: root, runRoot,
  runtime,
  profile: { id: "production", producer_report_key: "build-web", producer_operation: "harness.build_web", entries: ["index.html"] }, staging,
});
JS

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

assert_json "$stack_file" 'value.schema_id === "cartulary.web_e2e_stack.v8"' "v8 schema identity"
assert_json "$stack_file" 'value.suite_id === "suite-test" && value.browser_session_id === "session-default"' "v7 suite/session identity"
assert_json "$stack_file" 'value.postgres_identity.database_name === "ct_web_test" && value.object_store_identity.bucket === "ct-web-test"' "v7 isolated resource identity"
assert_json "$stack_file" 'value.frontend.frontend_command_kind === "vite-preview"' "v7 preview identity"
if grep -Eq 'access_key|secret|postgres://' "$stack_file"; then
  fail "v8 stack must not contain credentials or DSNs"
fi

attachment_exports="$("$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file")"
eval "$attachment_exports"
[[ "$CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED" == "1" ]] ||
  fail "valid v7 attachment must set the admission marker"
[[ "$CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER" == "1" ]] ||
  fail "valid v7 attachment must select external-server lifecycle semantics"
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
  fail "Playwright environment must carry the v7 admission marker"
printf '%s\n' "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" |
  grep -Fq "CARTULARY_PLAYWRIGHT_STATE_DIR=$runtime_root/playwright-state" ||
  fail "Playwright environment must carry session-scoped shared state"

if CARTULARY_BROWSER_RUNTIME_PROFILE_ID=network_flow_claimed \
  "$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file" >/dev/null 2>&1; then
  fail "profile-mismatched v7 attachment must fail"
fi
if CARTULARY_BROWSER_STAGE=measurement "$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file" >"$tmp_dir/wrong-artifact.log" 2>&1; then
  fail "production artifact must not attach to measurement"
fi
assert_file_contains "$tmp_dir/wrong-artifact.log" 'frontend artifact does not match selected stage' "artifact purpose rejection"
cp "$stack_file" "$tmp_dir/original-stack.json"
"$NODE_BIN" - "$stack_file" <<'JS'
const fs = require("node:fs");
const file = process.argv[2];
const value = JSON.parse(fs.readFileSync(file, "utf8"));
value.frontend.build_artifact_sha256 = `sha256:${"0".repeat(64)}`;
fs.writeFileSync(file, JSON.stringify(value));
JS
if "$NODE_BIN" "$EVIDENCE_HELPER" attach "$stack_file" >"$tmp_dir/stale-build.log" 2>&1; then
  fail "stale frontend build digest must fail attachment"
fi
assert_file_contains "$tmp_dir/stale-build.log" 'frontend build digest mismatch' "stale build rejection"
cp "$tmp_dir/original-stack.json" "$stack_file"
if "$NODE_BIN" "$EVIDENCE_HELPER" stack >/dev/null 2>&1; then
  fail "v8 stack publication must be immutable"
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
  'value.schema_id === "cartulary.browser_startup_diagnostics.v3" && value.status === "failed"' \
  "failed v2 terminal diagnostic"

# Browser reset is a process-replacement lifecycle, not a live HTTP route or
# generation-renewal compatibility path.
if CARTULARY_TEST_TARGET=adhoc \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/reset-results" \
  CARTULARY_TEST_RUN_ID=missing-lease \
  NODE_BIN="$NODE_BIN" \
  "$RESET_SCRIPT" --label missing-lease >/dev/null 2>&1; then
  fail "browser reset without the scheduler-owned session lease must fail"
fi
if "$RESET_SCRIPT" --label retired-renew --renew-generation 2 >/dev/null 2>&1; then
  fail "retired functional renewal option must fail"
fi
if grep -Fq '/api/v1/test/runtime/reset' "$ROOT_DIR/tools/harness/browser/lifecycle/reset-route.sh"; then
  fail "browser reset lifecycle must not call the retired runtime reset route"
fi
grep -Fq -- '--session-reset-backend' "$ROOT_DIR/tools/harness/browser/lifecycle/reset-route.sh" ||
  fail "browser reset wrapper must delegate to backend replacement"
grep -Fq -- '--database-result-file' "$ROOT_DIR/tools/harness/browser/lifecycle/reset-route.sh" ||
  fail "browser reset wrapper must retain the database diagnostic"

