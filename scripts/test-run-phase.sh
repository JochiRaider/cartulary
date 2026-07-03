#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/tools/harness/core/run-phase.sh"
GO_HELPER="$ROOT_DIR/tools/harness/backend/run-go-phase.sh"
GO_MANIFEST_HELPER="$ROOT_DIR/tools/harness/backend/run-go-manifest-phase.sh"
ARTIFACT_ERROR_EXIT=11
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE

json_field() {
  local file="$1"
  local path="$2"

  "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "$file" "$path"
}

assert_json_field_absent() {
  local file="$1"
  local path="$2"
  local label="$3"

  if "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
process.exit(value === undefined ? 0 : 1);
' "$file" "$path"; then
    return 0
  fi
  fail "$label: expected JSON field [$path] to be absent"
}

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_matches() {
  local value="$1"
  local pattern="$2"
  local label="$3"

  if [[ ! "$value" =~ $pattern ]]; then
    fail "$label: expected [$value] to match /$pattern/"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output to omit [$needle]"
  fi
}

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

assert_not_negative() {
  local actual="$1"
  local label="$2"

  if [[ -z "$actual" || ! "$actual" =~ ^-?[0-9]+$ || "$actual" == -* ]]; then
    fail "$label: expected a non-negative integer, got [$actual]"
  fi
}

assert_at_least() {
  local actual="$1"
  local minimum="$2"
  local label="$3"

  if [[ -z "$actual" || ! "$actual" =~ ^-?[0-9]+$ || "$actual" == -* || "$actual" -lt "$minimum" ]]; then
    fail "$label: expected an integer >= $minimum, got [$actual]"
  fi
}

write_target_summary() {
  local results_dir="$1"
  local run_id="$2"
  local target="$3"
  local duration_ms="$4"
  local wall_duration_ms="$5"
  local phases="$6"
  local tests="$7"
  local target_dir="${results_dir}/${run_id}/${target}"

  mkdir -p "$target_dir"
  cat >"$target_dir/target-summary.json" <<JSON
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "${target}",
  "kind": "leaf",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "own": {
    "target": "${target}",
    "status": "pass",
    "start_time": "2026-01-01T00:00:00Z",
    "end_time": "2026-01-01T00:00:01Z",
    "executed_duration_ms": ${duration_ms},
    "logical_duration_ms": ${duration_ms},
    "reused_duration_ms": 0,
    "derived_duration_ms": 0,
    "wall_duration_ms": ${wall_duration_ms},
    "critical_path_wall_duration_ms": ${wall_duration_ms},
    "teardown_duration_ms": 0,
    "accounting_modes": { "actual": ${phases}, "reused": 0, "derived": 0 },
    "counts": {
      "phases": ${phases},
      "tests": ${tests},
      "failed": 0,
      "authoritative": ${tests},
      "support": 0,
      "unmapped": 0,
      "non_test": 0,
      "authoritative_failed": 0,
      "support_failed": 0,
      "unmapped_failed": 0,
      "non_test_failed": 0,
      "packages": 1
    },
    "fixture": { "target": "${target}", "total_count": 0, "total_duration_ms": 0, "by_package": [], "by_test": [], "by_strategy": [], "slowest": [] },
    "artifacts": { "dir": ".cartulary/test-results/${run_id}/${target}" }
  },
  "children": {
    "target": "${target}",
    "status": "pass",
    "expected": [],
    "present": [],
    "missing": [],
    "failed_targets": [],
    "executed_duration_ms": 0,
    "logical_duration_ms": 0,
    "reused_duration_ms": 0,
    "derived_duration_ms": 0,
    "wall_duration_ms": 0,
    "critical_path_wall_duration_ms": 0,
    "teardown_duration_ms": 0,
    "accounting_modes": { "actual": 0, "reused": 0, "derived": 0 },
    "counts": { "phases": 0, "tests": 0, "failed": 0, "authoritative": 0, "support": 0, "unmapped": 0, "non_test": 0, "authoritative_failed": 0, "support_failed": 0, "unmapped_failed": 0, "non_test_failed": 0, "packages": 0 },
    "fixture": { "target": "${target}", "total_count": 0, "total_duration_ms": 0, "by_package": [], "by_test": [], "by_strategy": [], "slowest": [] }
  },
  "totals": {
    "target": "${target}",
    "status": "pass",
    "start_time": "2026-01-01T00:00:00Z",
    "end_time": "2026-01-01T00:00:01Z",
    "executed_duration_ms": ${duration_ms},
    "logical_duration_ms": ${duration_ms},
    "reused_duration_ms": 0,
    "derived_duration_ms": 0,
    "wall_duration_ms": ${wall_duration_ms},
    "critical_path_wall_duration_ms": ${wall_duration_ms},
    "teardown_duration_ms": 0,
    "accounting_modes": { "actual": ${phases}, "reused": 0, "derived": 0 },
    "counts": {
      "phases": ${phases},
      "tests": ${tests},
      "failed": 0,
      "authoritative": ${tests},
      "support": 0,
      "unmapped": 0,
      "non_test": 0,
      "authoritative_failed": 0,
      "support_failed": 0,
      "unmapped_failed": 0,
      "non_test_failed": 0,
      "packages": 1
    },
    "fixture": { "target": "${target}", "total_count": 0, "total_duration_ms": 0, "by_package": [], "by_test": [], "by_strategy": [], "slowest": [] }
  }
}
JSON
}

write_fixture_event() {
  local results_dir="$1"
  local run_id="$2"
  local suite_id="$3"
  local sequence="$4"
  local event_type="$5"
  local target="$6"
  local duration_ms="$7"
  local fixture_policy="$8"
  local reuse_scope="$9"
  local caller_package="${10}"
  local test_name="${11}"
  local event_dir="${results_dir}/${run_id}/_shared/test-services/${suite_id}/events"

  mkdir -p "$event_dir"
  cat >"${event_dir}/2026-01-01T00-00-${sequence}Z-100-${sequence}-${event_type}.json" <<JSON
{
  "type": "${event_type}",
  "timestamp": "2026-01-01T00:00:${sequence}Z",
  "pid": 100,
  "service": "postgres",
  "name": "ct_fixture_${sequence}",
  "kind": "template-clone",
  "details": {
    "target": "${target}",
    "duration_ms": ${duration_ms},
    "fixture_policy": "${fixture_policy}",
    "reuse_scope": "${reuse_scope}",
    "caller_package": "${caller_package}",
    "caller_file": "${caller_package}/fixture_test.go",
    "test_name": "${test_name}",
    "preparation_strategy": "template-clone"
  }
}
JSON
}

quiet_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
    "$HELPER" "quiet success" -- bash -lc 'echo hidden-success-output'
)"
assert_empty "$quiet_success_output" "quiet success"

mkdir -p "$ROOT_DIR/tmp"

success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
    "$HELPER" "success log replay" -- bash -lc 'echo keep-this-warning >&2' \
    2>&1
)"
assert_not_contains "$success_log_output" "keep-this-warning" "success log replay output"
assert_contains "$success_log_output" "[RESULT] target=adhoc status=pass" "success summary output"
assert_not_contains "$success_log_output" "== success log replay ==" "success log replay banner"

legacy_success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
  CARTULARY_ENABLE_LEGACY_SUCCESS_LOG=1 \
    "$HELPER" "legacy success log replay" -- bash -lc 'echo keep-this-warning >&2' \
    2>&1
)"
assert_contains "$legacy_success_log_output" "keep-this-warning" "legacy success log replay output"

short_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-results.XXXXXX")"
cleanup_paths+=("$short_failure_results")
set +e
short_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$short_failure_results" \
  CARTULARY_TEST_RUN_ID="short-failure" \
    "$HELPER" "short failure" -- bash -lc 'echo short-failure >&2; exit 7' \
    2>&1
)"
short_failure_status=$?
set -e
if [[ "$short_failure_status" -ne 7 ]]; then
  fail "short failure: expected exit status 7, got $short_failure_status"
fi
assert_contains "$short_failure_output" "failure: short failure" "short failure label"
assert_contains "$short_failure_output" "coverage=non_test" "short failure coverage"
assert_contains "$short_failure_output" "runner=shell" "short failure runner"
assert_contains "$short_failure_output" "message=short-failure" "short failure message"
assert_contains "$short_failure_output" "raw=" "short failure raw path"
assert_not_contains "$short_failure_output" "== short failure ==" "short failure banner"
short_failure_summary="$short_failure_results/short-failure/adhoc/short-failure/phase-summary.json"
assert_equals "$(json_field "$short_failure_summary" "counts.failed")" "1" "short failure failed count"
assert_equals "$(json_field "$short_failure_summary" "counts.non_test")" "1" "short failure non-test count"
assert_equals "$(json_field "$short_failure_summary" "counts.non_test_failed")" "1" "short failure non-test failed count"
assert_equals "$(json_field "$short_failure_summary" "counts.unmapped_failed")" "0" "short failure unmapped failed count"
assert_equals "$(json_field "$short_failure_summary" "failure_class")" "harness" "short failure class"
assert_equals "$(json_field "$short_failure_summary" "failure_classes.harness")" "1" "short failure helper count"
assert_equals "$(json_field "$short_failure_summary" "failures.0.failure_class")" "harness" "short failure record class"
assert_matches "$(json_field "$short_failure_summary" "start_time")" '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$' "short failure millisecond start time"
assert_matches "$(json_field "$short_failure_summary" "end_time")" '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$' "short failure millisecond end time"

browser_start_failure_results="$(mktemp -d "$ROOT_DIR/tmp/browser-start-failure.XXXXXX")"
cleanup_paths+=("$browser_start_failure_results")
set +e
browser_start_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$browser_start_failure_results" \
  CARTULARY_TEST_RUN_ID="browser-start-failure" \
  CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
    "$HELPER" "browser-e2e startup frontend ready" -- bash -lc 'echo "frontend exited before readiness" >&2; echo "Error: ENOSPC: System limit for number of file watchers reached, watch '\''/home/rook/code/cartulary/apps/web/vite.config.ts'\''" >&2; exit 1' \
    2>&1
)"
browser_start_failure_status=$?
set -e
if [[ "$browser_start_failure_status" -ne 1 ]]; then
  fail "browser start failure: expected exit status 1, got $browser_start_failure_status"
fi
assert_contains "$browser_start_failure_output" "failure_class=infra" "browser start phase failure class"
assert_contains "$browser_start_failure_output" "reason=service_start_error" "browser start phase failure reason"
browser_start_failure_phase_summary="$browser_start_failure_results/browser-start-failure/browser-e2e-webserver-backed/browser-e2e-startup-frontend-ready/phase-summary.json"
assert_equals "$(json_field "$browser_start_failure_phase_summary" "failure_class")" "infra" "browser start phase summary class"
assert_equals "$(json_field "$browser_start_failure_phase_summary" "failure_reason")" "service_start_error" "browser start phase summary reason"
set +e
browser_start_failure_target_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$browser_start_failure_results" \
  CARTULARY_TEST_RUN_ID="browser-start-failure" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary browser-e2e-webserver-backed fail \
    2>&1
)"
browser_start_failure_target_status=$?
set -e
if [[ "$browser_start_failure_target_status" -ne 0 ]]; then
  fail "browser start target summary writer: expected status 0, got $browser_start_failure_target_status"
fi
assert_contains "$browser_start_failure_target_output" "failure_class=infra" "browser start target failure class"
assert_contains "$browser_start_failure_target_output" "reason=service_start_error" "browser start target failure reason"
browser_start_failure_target_summary="$browser_start_failure_results/browser-start-failure/browser-e2e-webserver-backed/target-summary.json"
browser_start_failure_tool_summary="$browser_start_failure_results/browser-start-failure/browser-e2e-webserver-backed/tool-run-summary.json"
assert_equals "$(json_field "$browser_start_failure_tool_summary" "exit_code")" "3" "browser start tool summary exit code"
assert_equals "$(json_field "$browser_start_failure_target_summary" "failure_class")" "infra" "browser start target summary class"
assert_equals "$(json_field "$browser_start_failure_target_summary" "failure_reason")" "service_start_error" "browser start target summary reason"

browser_resource_conflict_results="$(mktemp -d "$ROOT_DIR/tmp/browser-resource-conflict.XXXXXX")"
cleanup_paths+=("$browser_resource_conflict_results")
browser_resource_conflict_owned_stack="$browser_resource_conflict_results/browser-resource-conflict/browser-e2e-webserver-backed/owned-stack"
mkdir -p "$browser_resource_conflict_owned_stack"
cat >"$browser_resource_conflict_owned_stack/startup-diagnostics.json" <<'JSON'
{
  "schema_id": "cartulary.browser_startup_diagnostics.v1",
  "generated_at": "2026-01-01T00:00:00Z",
  "target": "browser-e2e-webserver-backed",
  "status": "fail",
  "startup_phase": "frontend_readiness",
  "frontend_mode": "preview",
  "frontend_command_kind": "vite-preview",
  "api_origin": "http://127.0.0.1:39080",
  "public_origin": "http://127.0.0.1:39000",
  "backend_port": 39080,
  "frontend_port": 39000,
  "failure_class": "infra",
  "failure_reason": "resource_conflict",
  "message": "Port 39000 is already in use",
  "logs": {
    "frontend": ".cartulary/test-results/run/browser-e2e-webserver-backed/owned-stack/web.log"
  }
}
JSON
set +e
browser_resource_conflict_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$browser_resource_conflict_results" \
  CARTULARY_TEST_RUN_ID="browser-resource-conflict" \
  CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
    "$HELPER" "browser-e2e startup frontend ready" -- bash -lc 'echo "frontend exited before readiness" >&2; exit 1' \
    2>&1
)"
browser_resource_conflict_status=$?
set -e
if [[ "$browser_resource_conflict_status" -ne 1 ]]; then
  fail "browser resource conflict: expected exit status 1, got $browser_resource_conflict_status"
fi
assert_contains "$browser_resource_conflict_output" "failure_class=infra" "browser resource conflict phase failure class"
assert_contains "$browser_resource_conflict_output" "reason=resource_conflict" "browser resource conflict phase failure reason"
browser_resource_conflict_phase_summary="$browser_resource_conflict_results/browser-resource-conflict/browser-e2e-webserver-backed/browser-e2e-startup-frontend-ready/phase-summary.json"
assert_equals "$(json_field "$browser_resource_conflict_phase_summary" "failure_class")" "infra" "browser resource conflict phase summary class"
assert_equals "$(json_field "$browser_resource_conflict_phase_summary" "failure_reason")" "resource_conflict" "browser resource conflict phase summary reason"
set +e
browser_resource_conflict_target_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$browser_resource_conflict_results" \
  CARTULARY_TEST_RUN_ID="browser-resource-conflict" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary browser-e2e-webserver-backed fail \
    2>&1
)"
browser_resource_conflict_target_status=$?
set -e
if [[ "$browser_resource_conflict_target_status" -ne 0 ]]; then
  fail "browser resource conflict target summary writer: expected status 0, got $browser_resource_conflict_target_status"
fi
assert_contains "$browser_resource_conflict_target_output" "failure_class=infra" "browser resource conflict target failure class"
assert_contains "$browser_resource_conflict_target_output" "reason=resource_conflict" "browser resource conflict target failure reason"
browser_resource_conflict_target_summary="$browser_resource_conflict_results/browser-resource-conflict/browser-e2e-webserver-backed/target-summary.json"
browser_resource_conflict_tool_summary="$browser_resource_conflict_results/browser-resource-conflict/browser-e2e-webserver-backed/tool-run-summary.json"
assert_equals "$(json_field "$browser_resource_conflict_tool_summary" "exit_code")" "4" "browser resource conflict tool summary exit code"
assert_equals "$(json_field "$browser_resource_conflict_target_summary" "failure_class")" "infra" "browser resource conflict target summary class"
assert_equals "$(json_field "$browser_resource_conflict_target_summary" "failure_reason")" "resource_conflict" "browser resource conflict target summary reason"

listener_conflict_results="$(mktemp -d "$ROOT_DIR/tmp/listener-conflict.XXXXXX")"
cleanup_paths+=("$listener_conflict_results")
set +e
listener_conflict_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$listener_conflict_results" \
  CARTULARY_TEST_RUN_ID="listener-conflict" \
  CARTULARY_TEST_TARGET="browser-e2e-stateful" \
    "$HELPER" "browser-e2e startup services" -- bash -lc 'echo "listen tcp 127.0.0.1:8333: bind: address already in use" >&2; exit 1' \
    2>&1
)"
listener_conflict_status=$?
set -e
if [[ "$listener_conflict_status" -ne 1 ]]; then
  fail "listener conflict: expected exit status 1, got $listener_conflict_status"
fi
assert_contains "$listener_conflict_output" "failure_class=infra" "listener conflict phase failure class"
assert_contains "$listener_conflict_output" "reason=resource_conflict" "listener conflict phase failure reason"
listener_conflict_phase_summary="$listener_conflict_results/listener-conflict/browser-e2e-stateful/browser-e2e-startup-services/phase-summary.json"
assert_equals "$(json_field "$listener_conflict_phase_summary" "failure_class")" "infra" "listener conflict phase summary class"
assert_equals "$(json_field "$listener_conflict_phase_summary" "failure_reason")" "resource_conflict" "listener conflict phase summary reason"

shell_progress_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-progress-results.XXXXXX")"
cleanup_paths+=("$shell_progress_failure_results")
set +e
shell_progress_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$shell_progress_failure_results" \
  CARTULARY_TEST_RUN_ID="shell-progress-failure" \
    "$HELPER" "shell progress failure" -- bash -lc 'printf "%s\n" "[TARGET] start browser-e2e-webserver-backed service_backed=1 expected_phases=0 expected_tests=0" "real-shell-failure" >&2; exit 9' \
    2>&1
)"
shell_progress_failure_status=$?
set -e
if [[ "$shell_progress_failure_status" -ne 9 ]]; then
  fail "shell progress failure: expected exit status 9, got $shell_progress_failure_status"
fi
assert_contains "$shell_progress_failure_output" "message=real-shell-failure" "shell progress failure message"
shell_progress_failure_summary="$shell_progress_failure_results/shell-progress-failure/adhoc/shell-progress-failure/phase-summary.json"
assert_equals "$(json_field "$shell_progress_failure_summary" "dossiers.0.message")" "real-shell-failure" "shell progress failure summary message"

shellcheck_stdout_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-shellcheck-results.XXXXXX")"
cleanup_paths+=("$shellcheck_stdout_failure_results")
set +e
shellcheck_stdout_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$shellcheck_stdout_failure_results" \
  CARTULARY_TEST_RUN_ID="shellcheck-stdout-failure" \
    "$HELPER" "shellcheck stdout failure" -- bash -lc 'printf "%s\n" "tools/harness/readiness/bootstrap-node-runtime.sh" "In scripts/test-json-shapes.sh line 200:" "               ^-- SC2016 (info): Expressions don'\''t expand in single quotes, use double quotes for that."; exit 11' \
    2>&1
)"
shellcheck_stdout_failure_status=$?
set -e
if [[ "$shellcheck_stdout_failure_status" -ne 11 ]]; then
  fail "shellcheck stdout failure: expected exit status 11, got $shellcheck_stdout_failure_status"
fi
shellcheck_message="ShellCheck SC2016 at scripts/test-json-shapes.sh:200: Expressions don't expand in single quotes, use double quotes for that."
assert_contains "$shellcheck_stdout_failure_output" "message=$shellcheck_message" "shellcheck stdout failure message"
shellcheck_stdout_failure_summary="$shellcheck_stdout_failure_results/shellcheck-stdout-failure/adhoc/shellcheck-stdout-failure/phase-summary.json"
assert_equals "$(json_field "$shellcheck_stdout_failure_summary" "dossiers.0.message")" "$shellcheck_message" "shellcheck stdout failure summary message"

shellcheck_diagnostic_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-shellcheck-diagnostic-results.XXXXXX")"
cleanup_paths+=("$shellcheck_diagnostic_failure_results")
set +e
shellcheck_diagnostic_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$shellcheck_diagnostic_failure_results" \
  CARTULARY_TEST_RUN_ID="shellcheck-diagnostic-failure" \
  CARTULARY_TEST_TARGET="lint-shell" \
    "$HELPER" "lint shell" -- bash -lc 'printf "%s\n" "In scripts/example.sh line 5:" "unused_var=1" "^-- SC2034 (warning): unused_var appears unused. Verify use (or export if used externally)."; exit 1' \
    2>&1
)"
shellcheck_diagnostic_failure_status=$?
set -e
if [[ "$shellcheck_diagnostic_failure_status" -ne 1 ]]; then
  fail "shellcheck diagnostic failure: expected exit status 1, got $shellcheck_diagnostic_failure_status"
fi
shellcheck_diagnostic_message="ShellCheck SC2034 at scripts/example.sh:5: unused_var appears unused. Verify use (or export if used externally)."
assert_contains "$shellcheck_diagnostic_failure_output" "reason=tool_diagnostic_failure" "shellcheck diagnostic failure reason"
assert_contains "$shellcheck_diagnostic_failure_output" "$shellcheck_diagnostic_message" "shellcheck diagnostic failure message"
shellcheck_diagnostic_failure_summary="$shellcheck_diagnostic_failure_results/shellcheck-diagnostic-failure/lint-shell/lint-shell/phase-summary.json"
assert_equals "$(json_field "$shellcheck_diagnostic_failure_summary" "failure_reason")" "tool_diagnostic_failure" "shellcheck diagnostic summary reason"
assert_equals "$(json_field "$shellcheck_diagnostic_failure_summary" "dossiers.0.message")" "$shellcheck_diagnostic_message" "shellcheck diagnostic summary message"

biome_diagnostic_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-biome-results.XXXXXX")"
cleanup_paths+=("$biome_diagnostic_failure_results")
set +e
biome_diagnostic_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$biome_diagnostic_failure_results" \
  CARTULARY_TEST_RUN_ID="biome-diagnostic-failure" \
  CARTULARY_TEST_TARGET="lint-biome" \
    "$HELPER" "lint biome" -- bash -lc 'printf "%s\n" "apps/web/src/example.ts:12:8 lint/style/noNonNullAssertion ━━━━━━━━━━" "  ! Forbidden non-null assertion."; exit 1' \
    2>&1
)"
biome_diagnostic_failure_status=$?
set -e
if [[ "$biome_diagnostic_failure_status" -ne 1 ]]; then
  fail "biome diagnostic failure: expected exit status 1, got $biome_diagnostic_failure_status"
fi
biome_message="Biome lint/style/noNonNullAssertion at apps/web/src/example.ts:12:8"
assert_contains "$biome_diagnostic_failure_output" "reason=tool_diagnostic_failure" "biome diagnostic failure reason"
assert_contains "$biome_diagnostic_failure_output" "$biome_message" "biome diagnostic failure message"
biome_diagnostic_failure_summary="$biome_diagnostic_failure_results/biome-diagnostic-failure/lint-biome/lint-biome/phase-summary.json"
assert_equals "$(json_field "$biome_diagnostic_failure_summary" "failure_reason")" "tool_diagnostic_failure" "biome diagnostic summary reason"
assert_equals "$(json_field "$biome_diagnostic_failure_summary" "dossiers.0.message")" "$biome_message" "biome diagnostic summary message"

govulncheck_security_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-govulncheck-results.XXXXXX")"
cleanup_paths+=("$govulncheck_security_results")
govulncheck_security_script="$govulncheck_security_results/fake-govulncheck-phase.sh"
cat >"$govulncheck_security_script" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

cat >"$CARTULARY_PHASE_ARTIFACT_DIR/govulncheck-findings.json" <<'JSON'
{
  "schema_id": "cartulary.govulncheck_findings.v1",
  "tool": "govulncheck",
  "status": "fail",
  "config": null,
  "counts": {
    "raw_event_count": 2,
    "osv_count": 1,
    "finding_count": 1,
    "blocking_count": 1,
    "reachability": {
      "module": 0,
      "package": 0,
      "symbol": 1
    }
  },
  "vulnerability_ids": ["GO-2099-0001"],
  "blocking_vulnerability_ids": ["GO-2099-0001"],
  "findings": [
    {
      "id": "GO-2099-0001",
      "aliases": [],
      "summary": "synthetic reachable vulnerability",
      "fixed_version": "",
      "fixed_versions": [],
      "affected_packages": [],
      "reachability": "symbol",
      "blocking": true,
      "modules": [],
      "packages": [],
      "symbols": [],
      "trace": []
    }
  ]
}
JSON
printf '%s\n' '=== Symbol Results ==='
exit 1
EOF
chmod +x "$govulncheck_security_script"
set +e
govulncheck_security_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$govulncheck_security_results" \
  CARTULARY_TEST_RUN_ID="govulncheck-security-failure" \
  CARTULARY_TEST_TARGET="go-vulncheck" \
    "$HELPER" "go-vulncheck" -- "$govulncheck_security_script" \
    2>&1
)"
govulncheck_security_status=$?
set -e
assert_equals "$govulncheck_security_status" "1" "Govulncheck security phase status"
govulncheck_message="govulncheck found 1 symbol-reachable vulnerabilities: GO-2099-0001"
assert_contains "$govulncheck_security_output" "reason=security_finding" "Govulncheck security failure reason"
assert_contains "$govulncheck_security_output" "$govulncheck_message" "Govulncheck security failure message"
govulncheck_security_summary="$govulncheck_security_results/govulncheck-security-failure/go-vulncheck/go-vulncheck/phase-summary.json"
govulncheck_security_target_summary="$govulncheck_security_results/govulncheck-security-failure/go-vulncheck/target-summary.json"
govulncheck_security_tool_summary="$govulncheck_security_results/govulncheck-security-failure/go-vulncheck/tool-run-summary.json"
assert_equals "$(json_field "$govulncheck_security_summary" "failure_class")" "security" "Govulncheck security summary class"
assert_equals "$(json_field "$govulncheck_security_summary" "failure_reason")" "security_finding" "Govulncheck security summary reason"
assert_equals "$(json_field "$govulncheck_security_summary" "failure_classes.security")" "1" "Govulncheck security class count"
assert_equals "$(json_field "$govulncheck_security_summary" "failure_reasons.security_finding")" "1" "Govulncheck security reason count"
assert_equals "$(json_field "$govulncheck_security_summary" "dossiers.0.message")" "$govulncheck_message" "Govulncheck security dossier message"
"${NODE:-node}" - "$govulncheck_security_target_summary" "$govulncheck_security_tool_summary" <<'JS'
const fs = require("node:fs");
const [targetSummaryFile, toolSummaryFile] = process.argv.slice(2);
const targetSummary = JSON.parse(fs.readFileSync(targetSummaryFile, "utf8"));
const toolSummary = JSON.parse(fs.readFileSync(toolSummaryFile, "utf8"));
for (const [label, summary] of [
  ["target", targetSummary],
  ["tool", toolSummary],
]) {
  const govulncheck = summary.extensions?.["cartulary.security"]?.govulncheck;
  if (govulncheck?.blocking_count !== 1) {
    throw new Error(`${label} summary Govulncheck blocking_count got ${govulncheck?.blocking_count}`);
  }
}
const findingArtifact = (toolSummary.summary_artifacts ?? []).find(
  (artifact) =>
    artifact.role === "govulncheck_findings" &&
    artifact.kind === "json" &&
    artifact.path.endsWith("/go-vulncheck/go-vulncheck/govulncheck-findings.json"),
);
if (!findingArtifact) {
  throw new Error("canonical tool summary must reference govulncheck findings");
}
JS

govulncheck_malformed_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-govulncheck-malformed-results.XXXXXX")"
cleanup_paths+=("$govulncheck_malformed_results")
govulncheck_malformed_script="$govulncheck_malformed_results/fake-govulncheck-malformed-phase.sh"
cat >"$govulncheck_malformed_script" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

cat >"$CARTULARY_PHASE_ARTIFACT_DIR/govulncheck-findings.json" <<'JSON'
{
  "schema_id": "cartulary.govulncheck_findings.v1",
  "tool": "govulncheck",
  "status": "fail",
  "counts": {
    "finding_count": 1,
    "blocking_count": 1
  },
  "blocking_vulnerability_ids": ["GO-2099-0001"],
  "findings": []
}
JSON
printf '%s\n' '=== Symbol Results ==='
exit 1
EOF
chmod +x "$govulncheck_malformed_script"
set +e
govulncheck_malformed_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$govulncheck_malformed_results" \
  CARTULARY_TEST_RUN_ID="govulncheck-malformed-failure" \
  CARTULARY_TEST_TARGET="go-vulncheck" \
    "$HELPER" "go-vulncheck" -- "$govulncheck_malformed_script" \
    2>&1
)"
govulncheck_malformed_status=$?
set -e
assert_equals "$govulncheck_malformed_status" "1" "Govulncheck malformed phase status"
assert_contains "$govulncheck_malformed_output" "reason=artifact_error" "Govulncheck malformed failure reason"
govulncheck_malformed_summary="$govulncheck_malformed_results/govulncheck-malformed-failure/go-vulncheck/go-vulncheck/phase-summary.json"
govulncheck_malformed_target_summary="$govulncheck_malformed_results/govulncheck-malformed-failure/go-vulncheck/target-summary.json"
assert_equals "$(json_field "$govulncheck_malformed_summary" "failure_class")" "artifact" "Govulncheck malformed summary class"
assert_equals "$(json_field "$govulncheck_malformed_summary" "failure_reason")" "artifact_error" "Govulncheck malformed summary reason"
assert_equals "$(json_field "$govulncheck_malformed_target_summary" "failure_class")" "artifact" "Govulncheck malformed target class"
"${NODE:-node}" - "$govulncheck_malformed_target_summary" <<'JS'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.extensions?.["cartulary.security"]?.govulncheck !== undefined) {
  throw new Error("malformed Govulncheck findings must not produce security extension data");
}
JS

gosec_security_results="$(mktemp -d "$ROOT_DIR/tmp/run-phase-gosec-results.XXXXXX")"
cleanup_paths+=("$gosec_security_results")
set +e
gosec_security_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$gosec_security_results" \
  CARTULARY_TEST_RUN_ID="gosec-security-failure" \
  CARTULARY_TEST_TARGET="go-gosec-targeted" \
    "$HELPER" "go-gosec-targeted" -- bash -lc 'printf "%s\n" "G304: Potential file inclusion via variable"; exit 1' \
    2>&1
)"
gosec_security_status=$?
set -e
assert_equals "$gosec_security_status" "1" "Gosec security phase status"
assert_contains "$gosec_security_output" "reason=security_finding" "Gosec security failure reason"
gosec_security_summary="$gosec_security_results/gosec-security-failure/go-gosec-targeted/go-gosec-targeted/phase-summary.json"
assert_equals "$(json_field "$gosec_security_summary" "failure_class")" "security" "Gosec security summary class"
assert_equals "$(json_field "$gosec_security_summary" "failure_reason")" "security_finding" "Gosec security summary reason"

single_span_results="$(mktemp -d "$ROOT_DIR/tmp/single-span-duration.XXXXXX")"
cleanup_paths+=("$single_span_results")
single_span_phase_dir="$single_span_results/single-span/short-target/short-phase"
mkdir -p "$single_span_phase_dir"
cat >"$single_span_phase_dir/phase-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_phase_summary.v3",
  "label": "short phase",
  "target": "short-target",
  "runner": "shell",
  "status": "pass",
  "phase": "non_test",
  "command": "true",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "accounting_mode": "actual",
  "executed_duration_ms": 660,
  "logical_duration_ms": 660,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 660,
  "critical_path_wall_duration_ms": 660,
  "teardown_duration_ms": 0,
  "timing_bucket": "test_command",
  "exit_status": 0,
  "artifacts": {},
  "counts": {
    "tests": 0,
    "failed": 0,
    "authoritative": 0,
    "support": 0,
    "unmapped": 0,
    "non_test": 1,
    "authoritative_failed": 0,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0
  },
  "failure_class": null,
  "failure_classes": { "product": 0, "security": 0, "config": 0, "infra": 0, "harness": 0, "artifact": 0, "timing": 0, "interrupted": 0, "unknown": 0 },
  "failures": [],
  "failure_headline": "",
  "owners": [],
  "inventory": [],
  "dossiers": [],
  "manifest_mismatch": null
}
JSON
CARTULARY_TEST_RESULTS_DIR="$single_span_results" \
CARTULARY_TEST_RUN_ID="single-span" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary short-target pass >/dev/null 2>&1
single_span_summary="$single_span_results/single-span/short-target/target-summary.json"
single_span_timing="$single_span_results/single-span/short-target/target-timing.json"
assert_equals "$(json_field "$single_span_summary" "totals.wall_duration_ms")" "660" "single span target wall uses monotonic duration"
assert_equals "$(json_field "$single_span_summary" "totals.critical_path_wall_duration_ms")" "660" "single span target critical uses monotonic duration"
assert_equals "$(json_field "$single_span_timing" "buckets.0.duration_ms")" "660" "single span timing bucket uses monotonic duration"

missing_target_results="$(mktemp -d "$ROOT_DIR/tmp/run-summary-missing-target.XXXXXX")"
cleanup_paths+=("$missing_target_results")
set +e
missing_target_output="$(
  CARTULARY_TEST_RESULTS_DIR="$missing_target_results" \
  CARTULARY_TEST_RUN_ID="missing-target" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "missing target" fail 0 1 - test-fast-service-backed \
    2>&1
)"
missing_target_status=$?
set -e
assert_equals "$missing_target_status" "$ARTIFACT_ERROR_EXIT" "missing target run summary status"
assert_contains "$missing_target_output" "[FAIL] target=missing target" "missing target run summary output"
assert_contains "$missing_target_output" "failure_class=artifact" "missing target run summary failure class"
assert_contains "$missing_target_output" "reason=artifact_error" "missing target run summary failure reason"
assert_contains "$missing_target_output" "artifact failure: missing target summary: test-fast-service-backed" "missing target run summary headline"
missing_target_summary="$missing_target_results/missing-target/run-summary.json"
assert_equals "$(json_field "$missing_target_summary" "counts.failed")" "1" "missing target failed count"
assert_equals "$(json_field "$missing_target_summary" "counts.non_test")" "1" "missing target non-test count"
assert_equals "$(json_field "$missing_target_summary" "counts.non_test_failed")" "1" "missing target non-test failed count"
assert_equals "$(json_field "$missing_target_summary" "failure_class")" "artifact" "missing target failure class"
assert_equals "$(json_field "$missing_target_summary" "failure_reason")" "artifact_error" "missing target failure reason"
assert_equals "$(json_field "$missing_target_summary" "failure_classes.artifact")" "1" "missing target artifact count"
assert_equals "$(json_field "$missing_target_summary" "summary_targets.missing.0")" "test-fast-service-backed" "missing target summary list"

infra_timing_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-infra-timing.XXXXXX")"
cleanup_paths+=("$infra_timing_results")
infra_phase_dir="$infra_timing_results/infra-timing/infra-target/pass-phase"
infra_service_dir="$infra_timing_results/infra-timing/_shared/test-services/suite/events"
mkdir -p "$infra_phase_dir" "$infra_service_dir"
cat >"$infra_phase_dir/phase-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_phase_summary.v3",
  "label": "infra passing tests",
  "target": "infra-target",
  "runner": "shell",
  "status": "pass",
  "phase": "phase0",
  "command": "true",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "accounting_mode": "actual",
  "executed_duration_ms": 1000,
  "logical_duration_ms": 1000,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 1000,
  "critical_path_wall_duration_ms": 1000,
  "teardown_duration_ms": 0,
  "timing_bucket": "test_command",
  "exit_status": 0,
  "artifacts": {},
  "counts": {
    "tests": 1,
    "failed": 0,
    "authoritative": 1,
    "support": 0,
    "unmapped": 0,
    "non_test": 0,
    "authoritative_failed": 0,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0,
    "packages": 1
  },
  "failure_class": null,
  "failure_classes": { "product": 0, "security": 0, "config": 0, "infra": 0, "harness": 0, "artifact": 0, "timing": 0, "interrupted": 0, "unknown": 0 },
  "failures": [],
  "failure_headline": "",
  "owners": [],
  "inventory": [],
  "dossiers": [],
  "manifest_mismatch": null
}
JSON
cat >"$infra_service_dir/001.json" <<'JSON'
{
  "type": "timing-span",
  "timestamp": "2026-01-01T00:00:02Z",
  "details": {
    "target": "infra-target",
    "bucket": "service_wait",
    "label": "test-services start object-store",
    "start_time": "2026-01-01T00:00:01Z",
    "end_time": "2026-01-01T00:00:02Z",
    "duration_ms": 1000,
    "status": "fail"
  }
}
JSON
infra_timing_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$infra_timing_results" \
  CARTULARY_TEST_RUN_ID="infra-timing" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary infra-target pass \
    2>&1
)"
assert_contains "$infra_timing_output" "failure_class=infra" "infra timing target failure class"
assert_contains "$infra_timing_output" "tests passed; infra reason=preflight_error timing failure: test-services start object-store" "infra timing target headline"
infra_timing_summary="$infra_timing_results/infra-timing/infra-target/target-summary.json"
assert_equals "$(json_field "$infra_timing_summary" "failure_class")" "infra" "infra timing JSON failure class"
assert_equals "$(json_field "$infra_timing_summary" "failure_classes.infra")" "1" "infra timing JSON class count"
assert_equals "$(json_field "$infra_timing_summary" "failures.0.kind")" "timing" "infra timing JSON failure kind"

retry_timing_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-retry-timing.XXXXXX")"
cleanup_paths+=("$retry_timing_results")
retry_service_dir="$retry_timing_results/retry-timing/_shared/test-services/suite/events"
mkdir -p "$retry_service_dir"
cat >"$retry_service_dir/001.json" <<'JSON'
{
  "type": "timing-span",
  "timestamp": "2026-01-01T00:00:02Z",
  "details": {
    "target": "retry-target",
    "bucket": "service_wait",
    "label": "test-services start object-store attempt 1",
    "start_time": "2026-01-01T00:00:01Z",
    "end_time": "2026-01-01T00:00:02Z",
    "duration_ms": 1000,
    "status": "fail",
    "service": "object_store",
    "startup_attempt": true,
    "attempt": 1,
    "max_attempts": 2,
    "retryable": true,
    "retry_scheduled": true,
    "retry_blocked_by_context": false
  }
}
JSON
retry_timing_output="$(
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_TEST_RESULTS_DIR="$retry_timing_results" \
  CARTULARY_TEST_RUN_ID="retry-timing" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary retry-target pass \
    2>&1
)"
assert_contains "$retry_timing_output" "[RESULT] target=retry-target status=pass" "retry-scheduled startup target status"
retry_timing_summary="$retry_timing_results/retry-timing/retry-target/target-summary.json"
assert_equals "$(json_field "$retry_timing_summary" "status")" "pass" "retry-scheduled startup JSON status"
assert_equals "$(json_field "$retry_timing_summary" "failures.length")" "0" "retry-scheduled startup failure count"
assert_equals "$(json_field "$retry_timing_summary" "own.timing_failures.length")" "0" "retry-scheduled startup timing failure count"
assert_equals "$(json_field "$retry_timing_summary" "totals.counts.failed")" "0" "retry-scheduled startup total failed count"

skipped_after_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-summary-skipped-after-failure.XXXXXX")"
cleanup_paths+=("$skipped_after_failure_results")
CARTULARY_TEST_RESULTS_DIR="$skipped_after_failure_results" \
CARTULARY_TEST_RUN_ID="skipped-after-failure" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary failed-check fail >/dev/null 2>&1
set +e
skipped_after_failure_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$skipped_after_failure_results" \
  CARTULARY_TEST_RUN_ID="skipped-after-failure" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "skipped after failure" fail 0 1 failed-check \
      --summary-groups "harness=failed-check,skipped-check" \
      --skipped-after-failure skipped-check \
      failed-check skipped-check \
    2>&1
)"
skipped_after_failure_status=$?
set -e
assert_equals "$skipped_after_failure_status" "1" "skipped after failure run summary status"
assert_contains "$skipped_after_failure_output" "work_unit=failed-check" "skipped after failure root cause output"
assert_contains "$skipped_after_failure_output" "skipped_after_failure=skipped-check" "skipped after failure group output"
assert_not_contains "$skipped_after_failure_output" "missing_summary_targets=skipped-check" "skipped after failure missing output"
skipped_after_failure_summary="$skipped_after_failure_results/skipped-after-failure/run-summary.json"
assert_equals "$(json_field "$skipped_after_failure_summary" "counts.failed")" "1" "skipped after failure failed count"
assert_equals "$(json_field "$skipped_after_failure_summary" "counts.non_test_failed")" "1" "skipped after failure non-test failed count"
assert_equals "$(json_field "$skipped_after_failure_summary" "failure_class")" "harness" "skipped after failure class"
assert_equals "$(json_field "$skipped_after_failure_summary" "work_units.aborted_after")" "failed-check" "skipped after failure aborted target"
assert_equals "$(json_field "$skipped_after_failure_summary" "summary_targets.missing.length")" "0" "skipped after failure missing target count"
assert_equals "$(json_field "$skipped_after_failure_summary" "summary_targets.skipped_after_failure.0")" "skipped-check" "skipped after failure summary list"
assert_equals "$(json_field "$skipped_after_failure_summary" "summary_groups.0.skipped_after_failure.0")" "skipped-check" "skipped after failure group list"
assert_equals "$(json_field "$skipped_after_failure_summary" "summary_groups.0.missing_summary_targets.length")" "0" "skipped after failure group missing count"

child_summary_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-children.XXXXXX")"
cleanup_paths+=("$child_summary_results")
write_target_summary "$child_summary_results" "child-summary" "child-a" 1000 1200 2 7
write_target_summary "$child_summary_results" "child-summary" "child-b" 2000 2000 3 11
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
CARTULARY_TEST_RUN_ID="child-summary" \
CARTULARY_TEST_TARGET="parent-target" \
  "$HELPER" "parent target" -- bash -lc 'true' >/dev/null
child_target_output="$(
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-target pass --children child-a,child-b \
    2>&1
)"
assert_contains "$child_target_output" "[RESULT] target=parent-target status=pass" "child target parent output"
assert_contains "$child_target_output" "[ARTIFACTS] target=parent-target" "child target artifact output"
assert_not_contains "$child_target_output" "[CHILD] parent-target child-a" "quiet child target hides child-a detail"
assert_not_contains "$child_target_output" "[CHILD] parent-target child-b" "quiet child target hides child-b detail"
assert_not_contains "$child_target_output" " duration=" "child target ambiguous duration output"
parent_target_summary="$child_summary_results/child-summary/parent-target/target-summary.json"
parent_target_timing="$child_summary_results/child-summary/parent-target/target-timing.json"
assert_equals "$(json_field "$parent_target_summary" "schema_id")" "cartulary.test_target_summary.v4" "parent target summary schema"
assert_equals "$(json_field "$parent_target_summary" "kind")" "aggregate" "parent target summary kind"
assert_not_negative "$(json_field "$parent_target_summary" "totals.wall_duration_ms")" "parent target wall duration"
assert_not_negative "$(json_field "$parent_target_summary" "totals.critical_path_wall_duration_ms")" "parent target critical path duration"
assert_not_negative "$(json_field "$parent_target_summary" "totals.executed_duration_ms")" "parent target executed duration"
assert_not_negative "$(json_field "$parent_target_summary" "totals.logical_duration_ms")" "parent target logical duration"
assert_json_field_absent "$parent_target_summary" "duration_ms" "parent target legacy duration"
assert_contains "$(json_field "$parent_target_summary" "own.artifacts.timing_json")" "target-timing.json" "parent timing artifact path"
assert_equals "$(json_field "$parent_target_summary" "own.counts.tests")" "0" "parent target own tests"
assert_equals "$(json_field "$parent_target_summary" "children.counts.tests")" "18" "parent target child tests"
assert_equals "$(json_field "$parent_target_summary" "totals.counts.tests")" "18" "parent target total tests"
assert_equals "$(json_field "$parent_target_summary" "own.accounting_modes.actual")" "1" "parent target own accounting count"
assert_equals "$(json_field "$parent_target_summary" "children.present.0.target")" "child-a" "child target summary first child"
assert_equals "$(json_field "$parent_target_summary" "children.present.1.totals.counts.tests")" "11" "child target summary second child tests"
assert_equals "$(json_field "$parent_target_summary" "children.missing.length")" "0" "child target summary missing list"
assert_equals "$(json_field "$parent_target_timing" "schema_id")" "cartulary.test_target_timing.v1" "parent target timing schema"
assert_equals "$(json_field "$parent_target_timing" "buckets.0.name")" "test_command" "parent target timing test command bucket"
assert_equals "$(json_field "$parent_target_timing" "buckets.1.name")" "report_collation" "parent target timing report collation bucket"
assert_equals "$(json_field "$parent_target_summary" "own.slowest_lifecycle_bucket.name")" "$(json_field "$parent_target_timing" "slowest_lifecycle_bucket.name")" "parent target summary slowest bucket"

explain_run_summary="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$child_summary_results" --run-id child-summary --target parent-target \
    2>&1
)"
assert_contains "$explain_run_summary" "[RUN] tool-summary-only target=parent-target" "explain-run target tool summary"
assert_contains "$explain_run_summary" "[TARGET] parent-target status=pass kind=aggregate tests=18 failed=0" "explain-run target summary"
assert_contains "$explain_run_summary" "failed_children=none missing_children=none skipped_children=none slowest_child=child-b(2.00s)" "explain-run compact child hints"
explain_run_children="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$child_summary_results/child-summary" --target parent-target --detail children \
    2>&1
)"
assert_contains "$explain_run_children" "[CHILD] child-a status=pass tests=7 failed=0 authoritative=7 support=0 raw=0 tooling_support=0 unowned_regression=0 unmapped=0 duration=1.20s" "explain-run child-a detail"
assert_contains "$explain_run_children" "[CHILD] child-b status=pass tests=11 failed=0 authoritative=11 support=0 raw=0 tooling_support=0 unowned_regression=0 unmapped=0 duration=2.00s" "explain-run child-b detail"
set +e
explain_run_logs_output="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$child_summary_results/child-summary" --detail logs \
    2>&1
)"
explain_run_logs_status=$?
set -e
assert_equals "$explain_run_logs_status" "1" "explain-run logs requires target status"
assert_contains "$explain_run_logs_output" "DETAIL=logs requires TARGET=<target>" "explain-run logs requires target output"
set +e
explain_run_progress_output="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$child_summary_results/child-summary" --detail progress \
    2>&1
)"
explain_run_progress_status=$?
set -e
assert_equals "$explain_run_progress_status" "1" "explain-run progress requires target status"
assert_contains "$explain_run_progress_output" "DETAIL=progress requires TARGET=<target>" "explain-run progress requires target output"

helper_run_results="$(mktemp -d "$ROOT_DIR/tmp/explain-run-helper.XXXXXX")"
cleanup_paths+=("$helper_run_results")
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$helper_run_results" \
CARTULARY_TEST_RUN_ID="helper-run" \
CARTULARY_TEST_TARGET="helper-target" \
  "$HELPER" "helper-target" -- bash -lc 'printf "helper stdout\n"; printf "helper stderr\n" >&2' >/dev/null
CARTULARY_TEST_RESULTS_DIR="$helper_run_results" \
CARTULARY_TEST_RUN_ID="helper-run" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary check pass 1 1 - \
    --helper-units helper-target --completed-helper-units helper-target >/dev/null
helper_run_summary="$helper_run_results/helper-run/run-summary.json"
assert_equals "$(json_field "$helper_run_summary" "schema_id")" "cartulary.test_run_summary.v6" "helper run summary schema"
assert_equals "$(json_field "$helper_run_summary" "helper_units.artifacts.0.target")" "helper-target" "helper run summary helper artifact target"
explain_helper_summary="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$helper_run_results/helper-run" \
    2>&1
)"
assert_contains "$explain_helper_summary" "[HELPER] helper-target status=pass phases=1" "explain-run helper summary line"
assert_contains "$explain_helper_summary" "[HELPER-PHASE] helper-target label=helper-target status=pass" "explain-run helper phase line"
explain_helper_logs="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$helper_run_results/helper-run" --target helper-target --detail logs \
    2>&1
)"
assert_contains "$explain_helper_logs" "helper stdout" "explain-run helper stdout log"
assert_contains "$explain_helper_logs" "helper stderr" "explain-run helper stderr log"

tool_only_results="$(mktemp -d "$ROOT_DIR/tmp/explain-run-tool-only.XXXXXX")"
cleanup_paths+=("$tool_only_results")
mkdir -p "$tool_only_results/tool-run/agent-finalize/agent-finalize"
cat >"$tool_only_results/tool-run/agent-finalize/tool-run-summary.json" <<'JSON'
{
  "target": "agent-finalize",
  "status": "pass",
  "exit_code": 0,
  "duration_ms": 1200,
  "output_mode": "summary",
  "run_root": "tmp/explain-run-tool-only/tool-run",
  "summary_artifacts": [],
  "log_artifacts": [],
  "counts": {
    "tests": 0,
    "failed": 0
  },
  "failure_class": null,
  "failure_reason": null,
  "failures": []
}
JSON
cat >"$tool_only_results/tool-run/agent-finalize/finalize-summary.json" <<JSON
{
  "target": "agent-finalize",
  "status": "pass",
  "results_dir_status": "valid",
  "generated": {
    "status": "unchanged",
    "updated_file_count": 0
  },
  "duration": {
    "status": "skipped"
  },
  "run_checks": {
    "status": "pass"
  },
  "actions": [
    {
      "action_id": "structure_ledger_refresh",
      "status": "pass",
      "execution_state": "executed",
      "duration_ms": 1200,
      "cache": {
        "state": "miss",
        "reason_code": "cache_record_missing"
      },
      "substeps": [
        {
          "id": "phase-ledgers",
          "target": "phase-ledgers",
          "status": "pass",
          "summary_json": "none",
          "stdout_log": "$tool_only_results/tool-run/agent-finalize/agent-finalize/stdout.log",
          "stderr_log": null
        }
      ]
    }
  ],
  "failures": []
}
JSON
printf "finalize child stdout\n" >"$tool_only_results/tool-run/agent-finalize/agent-finalize/stdout.log"
explain_tool_only_summary="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$tool_only_results/tool-run" \
    2>&1
)"
assert_contains "$explain_tool_only_summary" "[RUN] tool-summary-only target=agent-finalize" "explain-run tool-only run line"
assert_contains "$explain_tool_only_summary" "[FINALIZE] agent-finalize status=pass results_dir_status=valid" "explain-run finalizer summary line"
assert_contains "$explain_tool_only_summary" "[FINALIZE-ACTION] structure_ledger_refresh status=pass execution_state=executed cache_state=miss" "explain-run finalizer action line"
explain_tool_only_children="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$tool_only_results/tool-run" --target agent-finalize --detail children \
    2>&1
)"
assert_contains "$explain_tool_only_children" "[FINALIZE-SUBSTEP] action=structure_ledger_refresh id=phase-ledgers" "explain-run finalizer child line"
explain_tool_only_logs="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$tool_only_results/tool-run" --target agent-finalize --detail logs \
    2>&1
)"
assert_contains "$explain_tool_only_logs" "finalize child stdout" "explain-run finalizer child log"

missing_finalize_results="$(mktemp -d "$ROOT_DIR/tmp/explain-run-missing-finalize.XXXXXX")"
cleanup_paths+=("$missing_finalize_results")
mkdir -p "$missing_finalize_results/tool-run/agent-finalize"
cat >"$missing_finalize_results/tool-run/agent-finalize/tool-run-summary.json" <<'JSON'
{
  "target": "agent-finalize",
  "status": "pass",
  "exit_code": 0,
  "duration_ms": 1,
  "output_mode": "summary",
  "run_root": "tmp/explain-run-missing-finalize/tool-run",
  "summary_artifacts": [],
  "log_artifacts": [],
  "counts": {
    "tests": 0,
    "failed": 0
  },
  "failure_class": null,
  "failure_reason": null,
  "failures": []
}
JSON
explain_missing_finalize="$(
  "$ROOT_DIR/tools/harness/core/explain-run-cli.mjs" --results-dir "$missing_finalize_results/tool-run" \
    2>&1
)"
assert_contains "$explain_missing_finalize" "[FINALIZE] missing" "explain-run missing finalizer summary line"

nested_artifacts_results="$(mktemp -d "$ROOT_DIR/tmp/nested-phase-artifacts.XXXXXX")"
cleanup_paths+=("$nested_artifacts_results")
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$nested_artifacts_results" \
CARTULARY_TEST_RUN_ID="nested-artifacts" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
  "$HELPER" "frontend-toolchain" -- bash -lc ':' >/dev/null &
nested_pid_a=$!
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$nested_artifacts_results" \
CARTULARY_TEST_RUN_ID="nested-artifacts" \
CARTULARY_TEST_TARGET="browser-e2e-visual" \
  "$HELPER" "frontend-toolchain" -- bash -lc ':' >/dev/null &
nested_pid_b=$!
if ! wait "$nested_pid_a"; then
  fail "nested phase artifact owner a failed"
fi
if ! wait "$nested_pid_b"; then
  fail "nested phase artifact owner b failed"
fi
[[ -f "$nested_artifacts_results/nested-artifacts/browser-e2e-webserver-backed/frontend-toolchain/phase-summary.json" ]] || fail "nested phase artifacts missing webserver-backed owner"
[[ -f "$nested_artifacts_results/nested-artifacts/browser-e2e-visual/frontend-toolchain/phase-summary.json" ]] || fail "nested phase artifacts missing visual owner"
[[ ! -d "$nested_artifacts_results/nested-artifacts/frontend-toolchain" ]] || fail "nested phase artifacts must not use child target as owner"

empty_log_race_results="$(mktemp -d "$ROOT_DIR/tmp/empty-log-race.XXXXXX")"
cleanup_paths+=("$empty_log_race_results")
shared_stdout="$empty_log_race_results/stdout.log"
shared_stderr="$empty_log_race_results/stderr.log"
: >"$shared_stdout"
: >"$shared_stderr"
empty_log_race_pids=()
for phase_owner in phase-a phase-b; do
  phase_dir="$empty_log_race_results/${phase_owner}"
  mkdir -p "$phase_dir"
  CARTULARY_TEST_TARGET="$phase_owner" \
  CARTULARY_PHASE_LABEL="empty log race ${phase_owner}" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND=":" \
  CARTULARY_PHASE_COMMAND_ARGV='[":"]' \
  CARTULARY_PHASE_START_TIME="2026-01-01T00:00:00.000Z" \
  CARTULARY_PHASE_END_TIME="2026-01-01T00:00:00.001Z" \
  CARTULARY_PHASE_DURATION_MS="1" \
  CARTULARY_PHASE_WALL_DURATION_MS="1" \
  CARTULARY_PHASE_EXIT_STATUS="0" \
  CARTULARY_PHASE_STDOUT_LOG="$shared_stdout" \
  CARTULARY_PHASE_STDERR_LOG="$shared_stderr" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" shell-phase >/dev/null &
  empty_log_race_pids+=("$!")
done
for pid in "${empty_log_race_pids[@]}"; do
  if ! wait "$pid"; then
    fail "empty log race helper failed"
  fi
done
[[ -f "$empty_log_race_results/phase-a/phase-summary.json" ]] || fail "empty log race missing phase-a summary"
[[ -f "$empty_log_race_results/phase-b/phase-summary.json" ]] || fail "empty log race missing phase-b summary"
assert_json_field_absent "$empty_log_race_results/phase-a/phase-summary.json" "artifacts.stdout_log" "empty log race phase-a stdout artifact"
assert_json_field_absent "$empty_log_race_results/phase-b/phase-summary.json" "artifacts.stdout_log" "empty log race phase-b stdout artifact"

fixture_results="$(mktemp -d "$ROOT_DIR/tmp/fixture-reporting.XXXXXX")"
cleanup_paths+=("$fixture_results")
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "01" "postgres-db-reset" "fixture-target" 20000 "package_reset" "package-reused" "internal/modules/auth" "TestSlowB"
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "02" "postgres-db-reset" "fixture-target" 15000 "package_reset" "package-reused" "internal/modules/auth" "TestSlowA"
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "03" "postgres-db-created" "fixture-target" 1000 "template_clone" "per-test" "internal/modules/entities" "TestClone"
below_fixture_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  FIXTURE_THRESHOLD_MS=40000 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-target pass \
    2>&1
)"
assert_not_contains "$below_fixture_output" "[FIXTURE]" "fixture output below threshold"
fixture_target_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  FIXTURE_THRESHOLD_MS=30000 \
  FIXTURE_TOP=2 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-target pass \
    2>&1
)"
assert_contains "$fixture_target_output" "[FIXTURE] fixture-target total=36.0s count=3 top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=35.0s hotspots=internal/modules/auth/postgres/database-reset/package_reset/package-reused(35.0s,count=2),internal/modules/entities/postgres/database-create/template_clone/per-test(1.00s,count=1) slowest=TestSlowB(20.0s),TestSlowA(15.0s)" "fixture target threshold output"
assert_equals "$(json_field "$fixture_results/fixture-run/fixture-target/target-summary.json" "totals.fixture.total_duration_ms")" "36000" "fixture target summary duration"

write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "01" "postgres-db-reset" "fixture-tie" 10000 "package_reset" "package-reused" "internal/modules/auth" "TestResetA"
write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "02" "postgres-db-reset" "fixture-tie" 10000 "package_reset" "package-reused" "internal/modules/auth" "TestResetB"
write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "03" "postgres-transaction" "fixture-tie" 20000 "transaction" "transaction" "internal/modules/auth" "TestTxn"
fixture_tie_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  FIXTURE_THRESHOLD_MS=1 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-tie-run" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-tie pass \
    2>&1
)"
assert_contains "$fixture_tie_output" "top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=20.0s" "fixture strategy tie prefers count"
assert_contains "$fixture_tie_output" "hotspots=internal/modules/auth/postgres/database-reset/package_reset/package-reused(20.0s,count=2),internal/modules/auth/postgres/transaction/transaction/transaction(20.0s,count=1)" "fixture hotspot tie prefers count"

write_fixture_event "$fixture_results" "fixture-hotspot-cap-run" "fixture-suite" "01" "postgres-db-reset" "fixture-hotspot-cap" 4000 "package_reset" "package-reused" "internal/modules/one" "TestOne"
write_fixture_event "$fixture_results" "fixture-hotspot-cap-run" "fixture-suite" "02" "postgres-db-reset" "fixture-hotspot-cap" 3000 "package_reset" "package-reused" "internal/modules/two" "TestTwo"
write_fixture_event "$fixture_results" "fixture-hotspot-cap-run" "fixture-suite" "03" "postgres-db-reset" "fixture-hotspot-cap" 2000 "package_reset" "package-reused" "internal/modules/three" "TestThree"
write_fixture_event "$fixture_results" "fixture-hotspot-cap-run" "fixture-suite" "04" "postgres-db-reset" "fixture-hotspot-cap" 1000 "package_reset" "package-reused" "internal/modules/four" "TestFour"
fixture_hotspot_cap_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  FIXTURE_THRESHOLD_MS=1 \
  FIXTURE_TOP=9 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-hotspot-cap-run" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-hotspot-cap pass \
    2>&1
)"
assert_contains "$fixture_hotspot_cap_output" "hotspots=internal/modules/one/postgres/database-reset/package_reset/package-reused(4.00s,count=1),internal/modules/two/postgres/database-reset/package_reset/package-reused(3.00s,count=1),internal/modules/three/postgres/database-reset/package_reset/package-reused(2.00s,count=1)" "fixture hotspots cap at three"
assert_not_contains "$fixture_hotspot_cap_output" "internal/modules/four/postgres" "fixture hotspots omit fourth entry"

fixture_run_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  FIXTURE_THRESHOLD_MS=30000 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "fixture run" pass 1 1 - fixture-target \
    2>&1
)"
assert_contains "$fixture_run_output" "[FIXTURE] fixture run total=36.0s count=3 top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=35.0s hotspots=internal/modules/auth/postgres/database-reset/package_reset/package-reused(35.0s,count=2),internal/modules/entities/postgres/database-create/template_clone/per-test(1.00s,count=1)" "fixture run summary output"

mkdir -p "$fixture_results/older-run"
touch -d '2026-01-01T00:00:00Z' "$fixture_results/older-run"
touch -d '2030-01-02T00:00:00Z' "$fixture_results/fixture-run"
fixture_report_output="$(
  "$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results" --threshold-ms 30000 --top 2 \
    2>&1
)"
assert_contains "$fixture_report_output" "[FIXTURE] fixture run total=36.0s count=3" "fixture report newest run aggregate output"
assert_contains "$fixture_report_output" "hotspots=internal/modules/auth/postgres/database-reset/package_reset/package-reused(35.0s,count=2),internal/modules/entities/postgres/database-create/template_clone/per-test(1.00s,count=1)" "fixture report newest run hotspots"
assert_contains "$fixture_report_output" "[FIXTURE] fixture-target total=36.0s count=3" "fixture report newest run target output"
fixture_report_concrete_output="$(
  "$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results/fixture-run" --threshold-ms 30000 --top 2 \
    2>&1
)"
assert_contains "$fixture_report_concrete_output" "[FIXTURE] fixture run total=36.0s count=3" "fixture report concrete run aggregate output"
assert_contains "$fixture_report_concrete_output" "hotspots=internal/modules/auth/postgres/database-reset/package_reset/package-reused(35.0s,count=2),internal/modules/entities/postgres/database-create/template_clone/per-test(1.00s,count=1)" "fixture report concrete run hotspots"
assert_contains "$fixture_report_concrete_output" "[FIXTURE] fixture-target total=36.0s count=3" "fixture report concrete run target output"
if fixture_report_mismatch_output="$(
  "$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results/fixture-run" --run-id fixture-tie-run --threshold-ms 1 \
    2>&1
)"; then
  fail "fixture report concrete run mismatch: expected failure"
fi
assert_contains "$fixture_report_mismatch_output" "RESULTS_DIR points to run fixture-run, but RUN_ID requested fixture-tie-run" "fixture report concrete run mismatch error"
fixture_report_json="$fixture_results/fixture-report.json"
"$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results" --run-id fixture-run --threshold-ms 30000 --json >"$fixture_report_json"
assert_equals "$(json_field "$fixture_report_json" "schema_id")" "cartulary.fixture_report.v1" "fixture report schema"
assert_equals "$(json_field "$fixture_report_json" "run_id")" "fixture-run" "fixture report run id"
assert_equals "$(json_field "$fixture_report_json" "run_dir")" "$fixture_results/fixture-run" "fixture report run dir"
assert_equals "$(json_field "$fixture_report_json" "aggregate.total_duration_ms")" "36000" "fixture report aggregate duration"
assert_equals "$(json_field "$fixture_report_json" "targets.0.target")" "fixture-target" "fixture report target"

write_fixture_event "$fixture_results" "fixture-aggregate-run" "fixture-suite" "01" "postgres-db-reset" "fixture-child" 32000 "package_reset" "package-reused" "internal/modules/auth" "TestAggregateChild"
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-child pass >/dev/null 2>&1
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary fixture-parent pass --children fixture-child >/dev/null 2>&1
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary check pass 1 1 - fixture-parent >/dev/null 2>&1
fixture_report_run_label_output="$(
  "$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results" --run-id fixture-aggregate-run --target check --threshold-ms 1 \
    2>&1
)"
assert_contains "$fixture_report_run_label_output" "[FIXTURE] check total=32.0s count=1" "fixture report run label target uses run summary"
fixture_report_aggregate_target_output="$(
  "$ROOT_DIR/tools/harness/core/fixture-report-cli.mjs" --results-dir "$fixture_results" --run-id fixture-aggregate-run --target fixture-parent --threshold-ms 1 \
    2>&1
)"
assert_contains "$fixture_report_aggregate_target_output" "[FIXTURE] fixture-parent total=32.0s count=1" "fixture report aggregate target uses target summary totals"

teardown_accounting_results="$(mktemp -d "$ROOT_DIR/tmp/target-timing-teardown-accounting.XXXXXX")"
cleanup_paths+=("$teardown_accounting_results")
teardown_services_dir="$teardown_accounting_results/teardown-accounting/_shared/test-services/web-fixture/events"
mkdir -p "$teardown_services_dir"
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
CARTULARY_TIMING_BUCKET="teardown" \
CARTULARY_TIMING_LABEL="browser-e2e stop owned processes" \
CARTULARY_TIMING_START_TIME="2026-01-01T00:00:00Z" \
CARTULARY_TIMING_END_TIME="2026-01-01T00:00:01.100Z" \
CARTULARY_TIMING_DURATION_MS="1100" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" timing-span
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
CARTULARY_TIMING_BUCKET="teardown" \
CARTULARY_TIMING_LABEL="browser-e2e overlapping process cleanup" \
CARTULARY_TIMING_START_TIME="2026-01-01T00:00:00.500Z" \
CARTULARY_TIMING_END_TIME="2026-01-01T00:00:01.500Z" \
CARTULARY_TIMING_DURATION_MS="1000" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" timing-span
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
CARTULARY_TIMING_BUCKET="teardown" \
CARTULARY_TIMING_LABEL="browser-e2e remove runtime root" \
CARTULARY_TIMING_START_TIME="2026-01-01T00:00:01.800Z" \
CARTULARY_TIMING_END_TIME="2026-01-01T00:00:02.100Z" \
CARTULARY_TIMING_DURATION_MS="300" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" timing-span
cat >"$teardown_services_dir/cleanup-browser-fixture.json" <<'JSON'
{
  "type": "timing-span",
  "name": "test-services cleanup browser e2e fixture",
  "status": "fail",
  "timestamp": "2026-01-01T00:00:01.800Z",
  "details": {
    "target": "browser-e2e-webserver-backed",
    "bucket": "teardown",
    "label": "test-services cleanup browser e2e fixture",
    "start_time": "2026-01-01T00:00:01.100Z",
    "end_time": "2026-01-01T00:00:01.800Z",
    "duration_ms": 700,
    "status": "fail"
  }
}
JSON
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary browser-e2e-webserver-backed pass >/dev/null 2>&1
teardown_accounting_summary="$teardown_accounting_results/teardown-accounting/browser-e2e-webserver-backed/target-summary.json"
teardown_accounting_timing="$teardown_accounting_results/teardown-accounting/browser-e2e-webserver-backed/target-timing.json"
assert_equals "$(json_field "$teardown_accounting_summary" "status")" "fail" "teardown accounting failed service span target summary status"
assert_equals "$(json_field "$teardown_accounting_summary" "kind")" "leaf" "teardown accounting target summary kind"
assert_equals "$(json_field "$teardown_accounting_timing" "status")" "fail" "teardown accounting failed service span target timing status"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.wall_duration_ms")" "2100" "teardown accounting target summary wall includes teardown"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.critical_path_wall_duration_ms")" "2100" "teardown accounting target critical path includes teardown"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.teardown_duration_ms")" "2100" "teardown accounting target teardown duration"
assert_equals "$(json_field "$teardown_accounting_summary" "children.expected.length")" "0" "teardown accounting target empty children"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.teardown_status")" "fail" "teardown accounting target teardown status"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.teardown_failures.0.label")" "test-services cleanup browser e2e fixture" "teardown accounting target teardown failure"
assert_equals "$(json_field "$teardown_accounting_summary" "totals.timing_failures.0.bucket")" "teardown" "teardown accounting target timing failure"
assert_equals "$(json_field "$teardown_accounting_summary" "own.counts.non_test_failed")" "1" "teardown accounting target non-test failed count"
assert_equals "$(json_field "$teardown_accounting_summary" "failure_class")" "artifact" "teardown accounting failure class"
assert_equals "$(json_field "$teardown_accounting_summary" "failure_classes.artifact")" "1" "teardown accounting artifact count"
assert_equals "$(json_field "$teardown_accounting_summary" "start_time")" "2026-01-01T00:00:00Z" "teardown accounting target summary start time"
assert_equals "$(json_field "$teardown_accounting_summary" "end_time")" "2026-01-01T00:00:02.100Z" "teardown accounting target summary end time"
assert_equals "$(json_field "$teardown_accounting_timing" "start_time")" "2026-01-01T00:00:00Z" "teardown accounting target timing start time"
assert_equals "$(json_field "$teardown_accounting_timing" "end_time")" "2026-01-01T00:00:02.100Z" "teardown accounting target timing end time"
assert_equals "$(json_field "$teardown_accounting_timing" "buckets.0.name")" "teardown" "teardown accounting bucket"
assert_equals "$(json_field "$teardown_accounting_timing" "buckets.0.duration_ms")" "2100" "teardown accounting disjoint duration"
assert_equals "$(json_field "$teardown_accounting_timing" "buckets.0.spans.length")" "4" "teardown accounting span count"
assert_equals "$(json_field "$teardown_accounting_timing" "slowest_lifecycle_bucket.name")" "teardown" "teardown accounting slowest bucket"

child_run_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "child run" pass 1 1 - \
      --summary-groups "backend-service-backed=child-a,child-b;browser=child-b" \
      parent-target \
    2>&1
)"
assert_contains "$child_run_output" "wall=" "child run wall duration output"
assert_contains "$child_run_output" "critical=" "child run critical path duration output"
assert_contains "$child_run_output" "exec=" "child run executed duration output"
assert_contains "$child_run_output" "logical=" "child run logical duration output"
assert_contains "$child_run_output" "teardown=" "child run teardown duration output"
assert_contains "$child_run_output" "slowest=parent-target:" "child run slowest target output"
assert_contains "$child_run_output" "[GROUP] child run backend-service-backed summary_targets=child-a,child-b status=pass wall=1.00s critical=1.00s exec=3.00s logical=3.00s teardown=0ms actual=5 reused=0 derived=0" "child run backend service group output"
assert_contains "$child_run_output" "[GROUP] child run browser summary_targets=child-b status=pass wall=1.00s critical=1.00s exec=2.00s logical=2.00s teardown=0ms actual=3 reused=0 derived=0" "child run browser group output"
assert_not_contains "$child_run_output" " duration=" "child run ambiguous duration output"
child_run_summary="$child_summary_results/child-summary/run-summary.json"
assert_not_negative "$(json_field "$child_run_summary" "wall_duration_ms")" "child run wall duration"
assert_not_negative "$(json_field "$child_run_summary" "critical_path_wall_duration_ms")" "child run critical path duration"
assert_not_negative "$(json_field "$child_run_summary" "executed_duration_ms")" "child run executed duration"
assert_not_negative "$(json_field "$child_run_summary" "logical_duration_ms")" "child run logical duration"
assert_json_field_absent "$child_run_summary" "duration_ms" "child run legacy duration"
assert_equals "$(json_field "$child_run_summary" "accounting_modes.actual")" "6" "child run actual accounting count"
assert_equals "$(json_field "$child_run_summary" "evidence_targets.summaries.0.target")" "parent-target" "run summary target object"
assert_equals "$(json_field "$child_run_summary" "slowest_lifecycle_bucket.target")" "parent-target" "child run slowest lifecycle bucket target"
assert_equals "$(json_field "$child_run_summary" "evidence_targets.summaries.0.children.present.1.target")" "child-b" "run summary preserved child target"
assert_equals "$(json_field "$child_run_summary" "summary_groups.0.name")" "backend-service-backed" "run summary backend group name"
assert_equals "$(json_field "$child_run_summary" "summary_groups.0.wall_duration_ms")" "1000" "run summary backend group wall duration"
assert_equals "$(json_field "$child_run_summary" "summary_groups.0.critical_path_wall_duration_ms")" "1000" "run summary backend group critical path duration"
assert_equals "$(json_field "$child_run_summary" "summary_groups.0.executed_duration_ms")" "3000" "run summary backend group executed duration"
assert_equals "$(json_field "$child_run_summary" "summary_groups.1.summary_targets.0")" "child-b" "run summary browser group target"
assert_equals "$(json_field "$child_run_summary" "summary_groups.1.status")" "pass" "run summary browser group status"

shared_execution_results="$(mktemp -d "$ROOT_DIR/tmp/run-summary-shared-execution.XXXXXX")"
cleanup_paths+=("$shared_execution_results")
write_target_summary "$shared_execution_results" "shared-execution" "target-fast" 100 100 1 1
write_target_summary "$shared_execution_results" "shared-execution" "target-slow" 2000 2000 1 1
shared_execution_dir="$shared_execution_results/shared-execution/_shared/backend-integration-incidents-shard-01"
mkdir -p "$shared_execution_dir"
CARTULARY_TEST_RESULTS_DIR="$shared_execution_results" \
CARTULARY_TEST_RUN_ID="shared-execution" \
  "$ROOT_DIR/tools/harness/core/test-output.sh" shared-execution \
    backend-integration-shards \
    backend-integration-incidents-shard-01 \
    pass \
    2026-01-01T00:00:00Z \
    2026-01-01T00:00:07Z \
    7000 \
    0 \
    "$shared_execution_dir/shared-execution.json"
shared_execution_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$shared_execution_results" \
  CARTULARY_TEST_RUN_ID="shared-execution" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "shared execution run" pass 2 2 - target-fast target-slow \
    2>&1
)"
assert_contains "$shared_execution_output" "slowest=target-slow:" "shared execution run slowest target ignores shared group"
assert_contains "$shared_execution_output" "[SHARED] shared execution run backend-integration-shards status=pass wall=7.00s exec=7.00s reports=1" "shared execution run group output"
shared_execution_summary="$shared_execution_results/shared-execution/run-summary.json"
assert_equals "$(json_field "$shared_execution_summary" "shared_execution_groups.0.schema_id")" "cartulary.test_shared_execution_group.v1" "shared execution group schema"
assert_equals "$(json_field "$shared_execution_summary" "shared_execution_groups.0.name")" "backend-integration-shards" "shared execution group name"
assert_equals "$(json_field "$shared_execution_summary" "shared_execution_groups.0.wall_duration_ms")" "7000" "shared execution group wall duration"
assert_equals "$(json_field "$shared_execution_summary" "shared_execution_groups.0.executed_duration_ms")" "7000" "shared execution group executed duration"
assert_equals "$(json_field "$shared_execution_summary" "shared_execution_groups.0.shared_reports.0")" "backend-integration-incidents-shard-01" "shared execution group report name"
assert_equals "$(json_field "$shared_execution_summary" "slowest_target.target")" "target-slow" "shared execution JSON slowest target"

set +e
missing_group_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" run-summary "child run missing group" pass 1 1 - \
      --summary-groups "browser=missing-browser" \
      parent-target \
    2>&1
)"
missing_group_status=$?
set -e
assert_equals "$missing_group_status" "$ARTIFACT_ERROR_EXIT" "missing group run summary status"
assert_contains "$missing_group_output" "reason=artifact_error" "missing group failure reason output"
assert_contains "$missing_group_output" "[GROUP] child run missing group browser summary_targets=missing-browser status=fail" "missing group output"
assert_contains "$missing_group_output" "missing_summary_targets=missing-browser" "missing group target output"
missing_group_summary="$child_summary_results/child-summary/run-summary.json"
assert_equals "$(json_field "$missing_group_summary" "failure_reason")" "artifact_error" "missing group failure reason"
assert_equals "$(json_field "$missing_group_summary" "summary_groups.0.missing_summary_targets.0")" "missing-browser" "missing group summary list"

missing_child_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-missing-child.XXXXXX")"
cleanup_paths+=("$missing_child_results")
missing_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$missing_child_results" \
  CARTULARY_TEST_RUN_ID="missing-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-missing pass --children missing-child \
    2>&1
)"
assert_contains "$missing_child_output" "[FAIL] parent-with-missing" "missing child parent output"
assert_contains "$missing_child_output" "[CHILD-MISSING] parent-with-missing missing-child" "missing child output"
missing_child_summary="$missing_child_results/missing-child/parent-with-missing/target-summary.json"
assert_equals "$(json_field "$missing_child_summary" "status")" "fail" "missing child status"
assert_equals "$(json_field "$missing_child_summary" "kind")" "aggregate" "missing child aggregate kind"
assert_equals "$(json_field "$missing_child_summary" "children.missing.0")" "missing-child" "missing child summary list"
assert_equals "$(json_field "$missing_child_summary" "own.counts.non_test_failed")" "1" "missing child wrapper failure count"
assert_equals "$(json_field "$missing_child_summary" "failure_class")" "artifact" "missing child failure class"

status_only_child_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-status-only-child.XXXXXX")"
cleanup_paths+=("$status_only_child_results")
status_only_child_run="$status_only_child_results/status-only-child"
mkdir -p "$status_only_child_run/status-only-child" "$status_only_child_run/parent-with-status-only-child"
cat >"$status_only_child_run/status-only-child/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "status-only-child",
  "kind": "leaf",
  "status": "fail",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "executed_duration_ms": 1000,
  "logical_duration_ms": 1000,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 1000,
  "critical_path_wall_duration_ms": 1000,
  "teardown_duration_ms": 0,
  "accounting_modes": { "actual": 1, "reused": 0, "derived": 0 },
  "counts": { "phases": 1, "tests": 1, "failed": 0, "authoritative": 1, "support": 0, "unmapped": 0, "non_test": 0, "authoritative_failed": 0, "support_failed": 0, "unmapped_failed": 0, "non_test_failed": 0, "packages": 1 },
  "failure_class": null,
  "failure_reason": null,
  "failure_classes": { "product": 0, "security": 0, "config": 0, "infra": 0, "harness": 0, "artifact": 0, "timing": 0, "interrupted": 0, "unknown": 0 },
  "failure_reasons": { "usage_error": 0, "configuration_error": 0, "preflight_error": 0, "service_start_error": 0, "service_readiness_timeout": 0, "fixture_error": 0, "resource_conflict": 0, "test_assertion_failure": 0, "security_finding": 0, "child_target_failure": 0, "tool_diagnostic_failure": 0, "scheduler_accounting_error": 0, "frontend_row_accounting": 0, "test_accounting_unmapped": 0, "artifact_error": 0, "cleanup_error": 0, "duration_baseline_drift": 0, "timeout_failure": 0, "cancelled_or_interrupted": 0, "unknown_failure": 0 },
  "failures": [],
  "failure_headline": "",
  "artifacts": { "dir": ".cartulary/test-results/status-only-child/status-only-child" }
}
JSON
status_only_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$status_only_child_results" \
  CARTULARY_TEST_RUN_ID="status-only-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-status-only-child fail --children status-only-child \
    2>&1
)"
assert_contains "$status_only_child_output" "[FAIL] parent-with-status-only-child" "status-only child parent output"
status_only_child_summary="$status_only_child_run/parent-with-status-only-child/target-summary.json"
assert_equals "$(json_field "$status_only_child_summary" "status")" "fail" "status-only child status"
assert_equals "$(json_field "$status_only_child_summary" "failure_class")" "harness" "status-only child failure class"
assert_equals "$(json_field "$status_only_child_summary" "failure_reason")" "child_target_failure" "status-only child failure reason"
assert_equals "$(json_field "$status_only_child_summary" "children.failures.0.child_target")" "status-only-child" "status-only child failure target"

skipped_child_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-skipped-child.XXXXXX")"
cleanup_paths+=("$skipped_child_results")
skipped_child_run="$skipped_child_results/skipped-child"
mkdir -p "$skipped_child_run/failed-backend" "$skipped_child_run/parent-with-skipped"
cat >"$skipped_child_run/failed-backend/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "failed-backend",
  "kind": "leaf",
  "status": "fail",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "executed_duration_ms": 1000,
  "logical_duration_ms": 1000,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 1000,
  "critical_path_wall_duration_ms": 1000,
  "teardown_duration_ms": 0,
  "accounting_modes": { "actual": 1, "reused": 0, "derived": 0 },
  "counts": { "phases": 1, "tests": 1, "failed": 1, "authoritative": 1, "support": 0, "unmapped": 0, "non_test": 0, "authoritative_failed": 1, "support_failed": 0, "unmapped_failed": 0, "non_test_failed": 0, "packages": 1 },
  "failure_class": "product",
  "failure_classes": { "product": 1, "security": 0, "config": 0, "infra": 0, "harness": 0, "artifact": 0, "timing": 0, "interrupted": 0, "unknown": 0 },
  "failures": [{ "failure_class": "product", "kind": "test", "target": "failed-backend", "message": "reported test failure" }],
  "failure_headline": "test: reported test failure",
  "artifacts": { "dir": ".cartulary/test-results/skipped-child/failed-backend" }
}
JSON
cat >"$skipped_child_run/parent-with-skipped/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_summary.v1",
  "target": "parent-with-skipped",
  "status": "fail",
  "failed_work_unit": "failed-backend",
  "skipped_work_units": [
    {
      "label": "skipped-browser",
      "id": "skipped-browser",
      "aggregate_target": "skipped-browser",
      "reason": "dependency_failure",
      "failed_dependency": "failed-backend"
    }
  ]
}
JSON
skipped_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$skipped_child_results" \
  CARTULARY_TEST_RUN_ID="skipped-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-skipped fail --children failed-backend,skipped-browser \
    2>&1
)"
assert_contains "$skipped_child_output" "[FAIL] parent-with-skipped" "skipped child parent output"
assert_contains "$skipped_child_output" "[CHILD] parent-with-skipped failed-backend status=fail failure_class=product" "skipped child failed dependency output"
assert_contains "$skipped_child_output" "[CHILD-SKIPPED] parent-with-skipped skipped-browser reason=dependency_failure failed_dependency=failed-backend" "skipped child cascade output"
assert_not_contains "$skipped_child_output" "[CHILD-MISSING] parent-with-skipped skipped-browser" "skipped child is not missing output"
skipped_child_summary="$skipped_child_run/parent-with-skipped/target-summary.json"
assert_equals "$(json_field "$skipped_child_summary" "status")" "fail" "skipped child status"
assert_equals "$(json_field "$skipped_child_summary" "failure_class")" "product" "skipped child preserves root failure class"
assert_equals "$(json_field "$skipped_child_summary" "children.failed_targets.0")" "failed-backend" "skipped child failed target"
assert_equals "$(json_field "$skipped_child_summary" "children.skipped.0.target")" "skipped-browser" "skipped child summary target"
assert_equals "$(json_field "$skipped_child_summary" "children.skipped.0.failed_dependency")" "failed-backend" "skipped child summary dependency"
assert_equals "$(json_field "$skipped_child_summary" "children.missing.length")" "0" "skipped child not missing"
assert_equals "$(json_field "$skipped_child_summary" "own.counts.non_test_failed")" "0" "skipped child does not create artifact failure"

mkdir -p "$skipped_child_run/scheduler-skipped-only"
cat >"$skipped_child_run/scheduler-skipped-only/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_summary.v1",
  "target": "scheduler-skipped-only",
  "status": "fail",
  "failed_work_unit": "lint-shell",
  "failed_work_unit_detail": {
    "label": "lint-shell",
    "id": "lint-shell",
    "aggregate_target": "lint-shell",
    "kind": "work_unit",
    "status": 1,
    "duration_ms": 12,
    "log_file": ".cartulary/test-results/skipped-child/scheduler-skipped-only/scheduler-logs/01-lint-shell.log"
  },
  "skipped_work_units": [
    {
      "label": "check-go-test-duration-baseline-drift",
      "id": "check-go-test-duration-baseline-drift",
      "aggregate_target": "check-go-test-duration-baseline-drift",
      "reason": "schedule_stopped_after_failure",
      "failed_dependency": "lint-shell"
    }
  ]
}
JSON
scheduler_skipped_only_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$skipped_child_results" \
  CARTULARY_TEST_RUN_ID="skipped-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary scheduler-skipped-only fail \
    2>&1
)"
assert_contains "$scheduler_skipped_only_output" "work_unit=lint-shell" "scheduler skipped-only work unit"
assert_contains "$scheduler_skipped_only_output" "child_target=lint-shell" "scheduler skipped-only failed aggregate target"
assert_not_contains "$scheduler_skipped_only_output" "child_target=check-go-test-duration-baseline-drift" "scheduler skipped-only skipped child is not failure target"

explicit_skipped_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$skipped_child_results" \
  CARTULARY_TEST_RUN_ID="skipped-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-explicit-skipped fail \
      --children failed-backend,explicit-skipped \
      --skipped-after-failure explicit-skipped \
      --failed-dependency failed-backend \
    2>&1
)"
assert_contains "$explicit_skipped_child_output" "[FAIL] parent-with-explicit-skipped" "explicit skipped child parent output"
assert_contains "$explicit_skipped_child_output" "[CHILD] parent-with-explicit-skipped failed-backend status=fail failure_class=product" "explicit skipped child failed dependency output"
assert_contains "$explicit_skipped_child_output" "[CHILD-SKIPPED] parent-with-explicit-skipped explicit-skipped reason=schedule_stopped_after_failure failed_dependency=failed-backend" "explicit skipped child cascade output"
assert_not_contains "$explicit_skipped_child_output" "[CHILD-MISSING] parent-with-explicit-skipped explicit-skipped" "explicit skipped child is not missing output"
explicit_skipped_child_summary="$skipped_child_run/parent-with-explicit-skipped/target-summary.json"
assert_equals "$(json_field "$explicit_skipped_child_summary" "children.failed_targets.0")" "failed-backend" "explicit skipped child failed target"
assert_equals "$(json_field "$explicit_skipped_child_summary" "children.skipped.0.target")" "explicit-skipped" "explicit skipped child summary target"
assert_equals "$(json_field "$explicit_skipped_child_summary" "children.skipped.0.reason")" "schedule_stopped_after_failure" "explicit skipped child reason"
assert_equals "$(json_field "$explicit_skipped_child_summary" "children.skipped.0.failed_dependency")" "failed-backend" "explicit skipped child dependency"
assert_equals "$(json_field "$explicit_skipped_child_summary" "children.missing.length")" "0" "explicit skipped child not missing"
assert_equals "$(json_field "$explicit_skipped_child_summary" "own.counts.non_test_failed")" "0" "explicit skipped child does not create artifact failure"

mkdir -p "$skipped_child_run/projected-child-source"
cat >"$skipped_child_run/projected-child-source/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "projected-child-source",
  "kind": "aggregate",
  "status": "fail",
  "children": {
    "skipped": [
      {
        "target": "imported-skipped",
        "work_unit": "imported-work-unit",
        "reason": "schedule_stopped_after_failure",
        "failed_dependency": "failed-backend"
      }
    ]
  }
}
JSON
imported_skipped_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$skipped_child_results" \
  CARTULARY_TEST_RUN_ID="skipped-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-imported-skipped fail \
      --children failed-backend,imported-skipped \
      --skipped-from-child projected-child-source \
    2>&1
)"
assert_contains "$imported_skipped_child_output" "[FAIL] parent-with-imported-skipped" "imported skipped child parent output"
assert_contains "$imported_skipped_child_output" "[CHILD] parent-with-imported-skipped failed-backend status=fail failure_class=product" "imported skipped child failed dependency output"
assert_contains "$imported_skipped_child_output" "[CHILD-SKIPPED] parent-with-imported-skipped imported-skipped reason=schedule_stopped_after_failure failed_dependency=failed-backend work_unit=imported-work-unit" "imported skipped child output"
assert_not_contains "$imported_skipped_child_output" "[CHILD-MISSING] parent-with-imported-skipped imported-skipped" "imported skipped child is not missing output"
assert_not_contains "$imported_skipped_child_output" "missing child target summary: imported-skipped" "imported skipped child avoids artifact failure"
imported_skipped_child_summary="$skipped_child_run/parent-with-imported-skipped/target-summary.json"
assert_equals "$(json_field "$imported_skipped_child_summary" "children.failed_targets.0")" "failed-backend" "imported skipped child failed target"
assert_equals "$(json_field "$imported_skipped_child_summary" "children.skipped.0.target")" "imported-skipped" "imported skipped child summary target"
assert_equals "$(json_field "$imported_skipped_child_summary" "children.skipped.0.work_unit")" "imported-work-unit" "imported skipped child work unit"
assert_equals "$(json_field "$imported_skipped_child_summary" "children.skipped.0.failed_dependency")" "failed-backend" "imported skipped child dependency"
assert_equals "$(json_field "$imported_skipped_child_summary" "children.missing.length")" "0" "imported skipped child not missing"
assert_equals "$(json_field "$imported_skipped_child_summary" "own.counts.non_test_failed")" "0" "imported skipped child does not create artifact failure"

mkdir -p "$skipped_child_run/external-scheduler-source"
cat >"$skipped_child_run/external-scheduler-source/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.scheduler_summary.v1",
  "target": "external-scheduler-source",
  "status": "fail",
  "failed_work_unit": "failed-backend",
  "skipped_work_units": [
    {
      "label": "scheduler-imported-work",
      "id": "scheduler-imported-work-id",
      "aggregate_target": "scheduler-imported-skipped",
      "reason": "dependency_failure",
      "failed_dependency": "failed-backend"
    }
  ]
}
JSON
scheduler_imported_skipped_child_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$skipped_child_results" \
  CARTULARY_TEST_RUN_ID="skipped-child" \
    "$ROOT_DIR/tools/harness/core/test-output.sh" target-summary parent-with-scheduler-imported-skipped fail \
      --children failed-backend,scheduler-imported-skipped \
      --skipped-from-scheduler external-scheduler-source \
    2>&1
)"
assert_contains "$scheduler_imported_skipped_child_output" "[CHILD-SKIPPED] parent-with-scheduler-imported-skipped scheduler-imported-skipped reason=dependency_failure failed_dependency=failed-backend work_unit=scheduler-imported-work" "scheduler imported skipped child output"
assert_not_contains "$scheduler_imported_skipped_child_output" "[CHILD-MISSING] parent-with-scheduler-imported-skipped scheduler-imported-skipped" "scheduler imported skipped child is not missing output"
assert_not_contains "$scheduler_imported_skipped_child_output" "missing child target summary: scheduler-imported-skipped" "scheduler imported skipped child avoids artifact failure"
scheduler_imported_skipped_child_summary="$skipped_child_run/parent-with-scheduler-imported-skipped/target-summary.json"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.failed_targets.0")" "failed-backend" "scheduler imported skipped child failed target"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.skipped.0.target")" "scheduler-imported-skipped" "scheduler imported skipped child summary target"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.skipped.0.reason")" "dependency_failure" "scheduler imported skipped child reason"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.skipped.0.work_unit")" "scheduler-imported-work" "scheduler imported skipped child work unit"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.skipped.0.failed_dependency")" "failed-backend" "scheduler imported skipped child dependency"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "children.missing.length")" "0" "scheduler imported skipped child not missing"
assert_equals "$(json_field "$scheduler_imported_skipped_child_summary" "own.counts.non_test_failed")" "0" "scheduler imported skipped child does not create artifact failure"

explicit_quiet_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$HELPER" "verbose override" -- bash -lc 'echo verbose-stream'
)"
assert_not_contains "$explicit_quiet_output" "== verbose override ==" "explicit quiet suppresses verbose override banner"
assert_not_contains "$explicit_quiet_output" "verbose-stream" "explicit quiet suppresses verbose override output"

verbose_default_output="$(
  VERBOSE=1 \
    "$HELPER" "verbose default" -- bash -lc 'echo verbose-stream'
)"
assert_contains "$verbose_default_output" "== verbose default ==" "verbose default banner"
assert_contains "$verbose_default_output" "verbose-stream" "verbose default output"

go_smoke_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-phase-smoke.XXXXXX")"
cleanup_paths+=("$go_smoke_dir")
cat >"$go_smoke_dir/run_go_phase_smoke_test.go" <<'EOF'
package rungophasesmoke

import "testing"

func TestPhase0_RunGoPhase_E_0_01(t *testing.T) {}
func TestPhase0_RunGoPhase_E_0_02(t *testing.T) {}
func TestUnrelatedRunGoPhase(t *testing.T)    {}
EOF

go_smoke_rel="./${go_smoke_dir#"$ROOT_DIR"/}"
go_bin="${GO:-go}"
go_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
    "$GO_HELPER" "run-go-phase smoke" '^(TestPhase0_.*_E_0_[0-9]+)$' -- "$go_bin" test "$go_smoke_rel"
)"
assert_empty "$go_success_output" "run-go-phase success"

set +e
go_zero_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
    "$GO_HELPER" "run-go-phase zero-match" '^(TestPhase0_.*_E_0_)$' -- "$go_bin" test "$go_smoke_rel" \
    2>&1
)"
go_zero_status=$?
set -e
if [[ "$go_zero_status" -eq 0 ]]; then
  fail "run-go-phase zero-match: expected non-zero exit status"
fi
assert_contains "$go_zero_output" "failure: run-go-phase zero-match" "run-go-phase zero-match label"
assert_contains "$go_zero_output" "message=phase matched zero tests" "run-go-phase zero-match message"

go_skip_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-phase-skip.XXXXXX")"
cleanup_paths+=("$go_skip_dir")
cat >"$go_skip_dir/run_go_phase_skip_test.go" <<'EOF'
package rungophaseskip

import "testing"

func TestPhase0_RunGoPhaseSkip_E_0_01(t *testing.T) {
	t.Skip("matched skip")
}
EOF

go_skip_rel="./${go_skip_dir#"$ROOT_DIR"/}"
set +e
go_skip_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
    "$GO_HELPER" "run-go-phase skip" '^(TestPhase0_RunGoPhaseSkip_E_0_01)$' -- "$go_bin" test "$go_skip_rel" \
    2>&1
)"
go_skip_status=$?
set -e
if [[ "$go_skip_status" -eq 0 ]]; then
  fail "run-go-phase skip: expected non-zero exit status"
fi
assert_contains "$go_skip_output" "go test inventory requires top-level pass" "run-go-phase skip message"
assert_contains "$go_skip_output" "runner=go_test" "run-go-phase skip runner"

go_pause_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-phase-pause.XXXXXX")"
cleanup_paths+=("$go_pause_dir")
cat >"$go_pause_dir/run_go_phase_pause_test.go" <<'EOF'
package rungophasepause

import "testing"

func TestPhase1_RunGoPhasePause_ProcessSmoke(t *testing.T) {
	t.Parallel()
	t.Fatalf("actual fatal line")
}
EOF

go_pause_rel="./${go_pause_dir#"$ROOT_DIR"/}"
set +e
go_pause_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
    "$GO_HELPER" "run-go-phase pause-filter smoke" '^(TestPhase1_.*_ProcessSmoke)$' -- "$go_bin" test "$go_pause_rel" -parallel 2 \
    2>&1
)"
go_pause_status=$?
set -e
if [[ "$go_pause_status" -eq 0 ]]; then
  fail "run-go-phase pause-filter: expected non-zero exit status"
fi
assert_contains "$go_pause_output" "failure: run-go-phase pause-filter smoke" "run-go-phase pause-filter label"
assert_contains "$go_pause_output" "actual fatal line" "run-go-phase pause-filter message"
assert_not_contains "$go_pause_output" "message==== PAUSE" "run-go-phase pause-filter pause message"
assert_not_contains "$go_pause_output" "message==== CONT" "run-go-phase pause-filter cont message"

go_pkg_setup_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-phase-package-setup.XXXXXX")"
cleanup_paths+=("$go_pkg_setup_dir")
cat >"$go_pkg_setup_dir/run_go_phase_package_setup_test.go" <<'EOF'
package rungophasepackagesetup

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Fprintln(os.Stderr, "start shared process harnesses: package setup failed")
	os.Exit(1)
}

func TestPhase1_RunGoPhasePackageSetup_ProcessSmoke(t *testing.T) {}
EOF

go_pkg_setup_rel="./${go_pkg_setup_dir#"$ROOT_DIR"/}"
set +e
go_pkg_setup_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
    "$GO_HELPER" "run-go-phase phase1 package setup smoke" '^(TestPhase1_.*_ProcessSmoke)$' -- "$go_bin" test "$go_pkg_setup_rel" \
    2>&1
)"
go_pkg_setup_status=$?
set -e
if [[ "$go_pkg_setup_status" -eq 0 ]]; then
  fail "run-go-phase package setup: expected non-zero exit status"
fi
assert_contains "$go_pkg_setup_output" "failure: run-go-phase phase1 package setup smoke" "run-go-phase package setup label"
assert_contains "$go_pkg_setup_output" "coverage=support" "run-go-phase package setup coverage"
assert_contains "$go_pkg_setup_output" "phase=phase1" "run-go-phase package setup phase"
assert_contains "$go_pkg_setup_output" "symbol_or_title=(package setup)" "run-go-phase package setup title"
assert_contains "$go_pkg_setup_output" "message=start shared process harnesses: package setup failed" "run-go-phase package setup message"

declare -A go_manifest_repo_map_digests=()
for synthetic_manifest in \
  "$ROOT_DIR/tools/phase9_test_map.json" \
  "$ROOT_DIR/tools/phase10_test_map.json" \
  "$ROOT_DIR/tools/phase11_test_map.json"
do
  if [[ -e "$synthetic_manifest" ]]; then
    go_manifest_repo_map_digests["$synthetic_manifest"]="$(sha256sum "$synthetic_manifest" | awk '{print $1}')"
  else
    go_manifest_repo_map_digests["$synthetic_manifest"]="__absent__"
  fi
done

go_manifest_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-smoke.XXXXXX")"
go_manifest_root="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-manifests.XXXXXX")"
go_manifest_tools="$go_manifest_root/tools"
mkdir -p "$go_manifest_tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$go_manifest_tools"/
cat >"$go_manifest_tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase9",
      "order": 9,
      "status": "active",
      "label": "Phase 9",
      "manifest_path": "tools/phase9_test_map.json",
      "ledger_path": "docs/testing/phase9_coverage_ledger.md",
      "scope": "synthetic phase9 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase10",
      "order": 10,
      "status": "active",
      "label": "Phase 10",
      "manifest_path": "tools/phase10_test_map.json",
      "ledger_path": "docs/testing/phase10_coverage_ledger.md",
      "scope": "synthetic phase10 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
cleanup_paths+=("$go_manifest_dir" "$go_manifest_root")
cat >"$go_manifest_dir/run_go_manifest_phase_smoke_test.go" <<'EOF'
package rungomanifestphasesmoke

import "testing"

func TestPhase9_RunGoManifest_U_9_01(t *testing.T) {}

func TestPhase10_RunGoManifest_U_10_01(t *testing.T) {
	t.Skip("matched manifest skip")
}
EOF

go_manifest_rel="./${go_manifest_dir#"$ROOT_DIR"/}"
cat >"$go_manifest_tools/phase9_test_map.json" <<EOF
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase9",
  "note": "Synthetic run-go manifest phase fixture.",
  "ledger": {
    "title": "Phase 9 Coverage Ledger",
    "notes": "Synthetic run-go manifest phase fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase9",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-9-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-9-01",
      "coverage": "authoritative",
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_unit::${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go::TestPhase9_RunGoManifest_U_9_01",
      "duplicate_of": "none",
      "evidence_delta": "synthetic run-go manifest smoke evidence",
      "warm_local_cost_class": "low",
      "runner": "go_test",
      "package": "$go_manifest_rel",
      "file": "${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go",
      "symbol": "TestPhase9_RunGoManifest_U_9_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest smoke",
      "out_of_scope": "synthetic run-go manifest smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
EOF
cat >"$go_manifest_tools/phase10_test_map.json" <<EOF
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase10",
  "note": "Synthetic run-go manifest phase fixture.",
  "ledger": {
    "title": "Phase 10 Coverage Ledger",
    "notes": "Synthetic run-go manifest phase fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase10",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-10-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-10-01",
      "coverage": "authoritative",
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_unit::${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go::TestPhase10_RunGoManifest_U_10_01",
      "duplicate_of": "none",
      "evidence_delta": "synthetic run-go manifest skip smoke evidence",
      "warm_local_cost_class": "low",
      "runner": "go_test",
      "package": "$go_manifest_rel",
      "file": "${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go",
      "symbol": "TestPhase10_RunGoManifest_U_10_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest skip smoke",
      "out_of_scope": "synthetic run-go manifest skip smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
EOF

node_bin="${NODE_BIN:-node}"
go_manifest_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase smoke" phase9 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_rel"
)"
assert_empty "$go_manifest_success_output" "run-go-manifest-phase success"

set +e
go_manifest_empty_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase empty" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_empty_status=$?
set -e
if [[ "$go_manifest_empty_status" -eq 0 ]]; then
  fail "run-go-manifest-phase empty: expected non-zero exit status"
fi
assert_contains "$go_manifest_empty_output" "no authoritative go tests found for phase9 unit in ./internal/platform/..." "run-go-manifest-phase empty output"

set +e
go_manifest_raw_allow_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION="phase9:unit:authoritative:backend_unit:./internal/platform/..." \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase raw allow" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_raw_allow_status=$?
set -e
if [[ "$go_manifest_raw_allow_status" -eq 0 ]]; then
  fail "run-go-manifest-phase raw allow: expected non-zero exit status"
fi
assert_contains "$go_manifest_raw_allow_output" "CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION is retired" "run-go-manifest-phase raw allow output"

go_manifest_valid_exception="$go_manifest_root/phase-policy-valid.json"
cat >"$go_manifest_valid_exception" <<'JSON'
{
  "schema_id": "cartulary.phase_policy_exceptions.v1",
  "exceptions": [
    {
      "id": "allow-empty-platform-phase9",
      "type": "allowed_empty_go_manifest_selection",
      "owner": "task-surface",
      "reason": "synthetic test-only empty selection allowance",
      "expires_before_phase": "phase99",
      "selection": {
        "phase": "phase9",
        "section": "unit",
        "coverage": "authoritative",
        "execution_dependency": "backend_unit",
        "package_patterns": ["./internal/platform/..."]
      }
    }
  ]
}
JSON
go_manifest_exception_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_PHASE_POLICY_EXCEPTIONS="$go_manifest_valid_exception" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase exception" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/...
)"
assert_empty "$go_manifest_exception_output" "run-go-manifest-phase exception"

go_manifest_missing_owner_exception="$go_manifest_root/phase-policy-missing-owner.json"
cat >"$go_manifest_missing_owner_exception" <<'JSON'
{
  "schema_id": "cartulary.phase_policy_exceptions.v1",
  "exceptions": [
    {
      "id": "missing-owner",
      "type": "allowed_empty_go_manifest_selection",
      "reason": "synthetic invalid exception",
      "expires_before_phase": "phase99",
      "selection": {
        "phase": "phase9",
        "section": "unit",
        "coverage": "authoritative",
        "execution_dependency": "backend_unit",
        "package_patterns": ["./internal/platform/..."]
      }
    }
  ]
}
JSON
set +e
go_manifest_missing_owner_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_PHASE_POLICY_EXCEPTIONS="$go_manifest_missing_owner_exception" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase missing owner" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_missing_owner_status=$?
set -e
if [[ "$go_manifest_missing_owner_status" -eq 0 ]]; then
  fail "run-go-manifest-phase missing owner: expected non-zero exit status"
fi
assert_contains "$go_manifest_missing_owner_output" ".owner must be a non-empty string" "run-go-manifest-phase missing owner output"

go_manifest_missing_reason_exception="$go_manifest_root/phase-policy-missing-reason.json"
cat >"$go_manifest_missing_reason_exception" <<'JSON'
{
  "schema_id": "cartulary.phase_policy_exceptions.v1",
  "exceptions": [
    {
      "id": "missing-reason",
      "type": "allowed_empty_go_manifest_selection",
      "owner": "task-surface",
      "expires_before_phase": "phase99",
      "selection": {
        "phase": "phase9",
        "section": "unit",
        "coverage": "authoritative",
        "execution_dependency": "backend_unit",
        "package_patterns": ["./internal/platform/..."]
      }
    }
  ]
}
JSON
set +e
go_manifest_missing_reason_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_PHASE_POLICY_EXCEPTIONS="$go_manifest_missing_reason_exception" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase missing reason" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_missing_reason_status=$?
set -e
if [[ "$go_manifest_missing_reason_status" -eq 0 ]]; then
  fail "run-go-manifest-phase missing reason: expected non-zero exit status"
fi
assert_contains "$go_manifest_missing_reason_output" ".reason must be a non-empty string" "run-go-manifest-phase missing reason output"

go_manifest_missing_expiration_exception="$go_manifest_root/phase-policy-missing-expiration.json"
cat >"$go_manifest_missing_expiration_exception" <<'JSON'
{
  "schema_id": "cartulary.phase_policy_exceptions.v1",
  "exceptions": [
    {
      "id": "missing-expiration",
      "type": "allowed_empty_go_manifest_selection",
      "owner": "task-surface",
      "reason": "synthetic invalid exception",
      "selection": {
        "phase": "phase9",
        "section": "unit",
        "coverage": "authoritative",
        "execution_dependency": "backend_unit",
        "package_patterns": ["./internal/platform/..."]
      }
    }
  ]
}
JSON
set +e
go_manifest_missing_expiration_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_PHASE_POLICY_EXCEPTIONS="$go_manifest_missing_expiration_exception" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase missing expiration" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_missing_expiration_status=$?
set -e
if [[ "$go_manifest_missing_expiration_status" -eq 0 ]]; then
  fail "run-go-manifest-phase missing expiration: expected non-zero exit status"
fi
assert_contains "$go_manifest_missing_expiration_output" "must declare exactly one of expires_before_phase or expires_on" "run-go-manifest-phase missing expiration output"

go_manifest_expired_exception="$go_manifest_root/phase-policy-expired.json"
cat >"$go_manifest_expired_exception" <<'JSON'
{
  "schema_id": "cartulary.phase_policy_exceptions.v1",
  "exceptions": [
    {
      "id": "expired-exception",
      "type": "allowed_empty_go_manifest_selection",
      "owner": "task-surface",
      "reason": "synthetic expired exception",
      "expires_on": "2000-01-01",
      "selection": {
        "phase": "phase9",
        "section": "unit",
        "coverage": "authoritative",
        "execution_dependency": "backend_unit",
        "package_patterns": ["./internal/platform/..."]
      }
    }
  ]
}
JSON
set +e
go_manifest_expired_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  CARTULARY_PHASE_POLICY_EXCEPTIONS="$go_manifest_expired_exception" \
  CARTULARY_PHASE_POLICY_TODAY="2026-01-01" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase expired" phase9 unit authoritative backend_unit -- "$go_bin" test ./internal/platform/... \
    2>&1
)"
go_manifest_expired_status=$?
set -e
if [[ "$go_manifest_expired_status" -eq 0 ]]; then
  fail "run-go-manifest-phase expired: expected non-zero exit status"
fi
assert_contains "$go_manifest_expired_output" "expired on 2000-01-01" "run-go-manifest-phase expired output"

set +e
go_manifest_skip_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase skip" phase10 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_rel" \
    2>&1
)"
go_manifest_skip_status=$?
set -e
if [[ "$go_manifest_skip_status" -eq 0 ]]; then
  fail "run-go-manifest-phase skip: expected non-zero exit status"
fi
assert_contains "$go_manifest_skip_output" "failure: run-go-manifest-phase skip" "run-go-manifest-phase skip label"
assert_contains "$go_manifest_skip_output" "go test inventory requires top-level pass" "run-go-manifest-phase skip message"

go_manifest_pkg_setup_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-package-setup.XXXXXX")"
cleanup_paths+=("$go_manifest_pkg_setup_dir")
cat >"$go_manifest_pkg_setup_dir/run_go_manifest_phase_package_setup_test.go" <<'EOF'
package rungomanifestphasepackagesetup

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Fprintln(os.Stderr, "manifest package setup failed")
	os.Exit(1)
}

func TestPhase11_RunGoManifestPackageSetup_U_11_01(t *testing.T) {}
EOF

go_manifest_pkg_setup_rel="./${go_manifest_pkg_setup_dir#"$ROOT_DIR"/}"
cat >"$go_manifest_tools/phase11_test_map.json" <<EOF
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase11",
  "note": "Synthetic run-go manifest package setup fixture.",
  "ledger": {
    "title": "Phase 11 Coverage Ledger",
    "notes": "Synthetic run-go manifest package setup fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase11",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-11-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-11-01",
      "coverage": "authoritative",
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "backend_unit::${go_manifest_pkg_setup_rel#./}/run_go_manifest_phase_package_setup_test.go::TestPhase11_RunGoManifestPackageSetup_U_11_01",
      "duplicate_of": "none",
      "evidence_delta": "synthetic run-go manifest package setup smoke evidence",
      "warm_local_cost_class": "low",
      "runner": "go_test",
      "package": "$go_manifest_pkg_setup_rel",
      "file": "${go_manifest_pkg_setup_rel#./}/run_go_manifest_phase_package_setup_test.go",
      "symbol": "TestPhase11_RunGoManifestPackageSetup_U_11_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest package setup smoke",
      "out_of_scope": "synthetic run-go manifest package setup smoke"
    }
  ],
  "integration": [],
  "e2e": []
}
EOF
cat >"$go_manifest_tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase9",
      "order": 9,
      "status": "active",
      "label": "Phase 9",
      "manifest_path": "tools/phase9_test_map.json",
      "ledger_path": "docs/testing/phase9_coverage_ledger.md",
      "scope": "synthetic phase9 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase10",
      "order": 10,
      "status": "active",
      "label": "Phase 10",
      "manifest_path": "tools/phase10_test_map.json",
      "ledger_path": "docs/testing/phase10_coverage_ledger.md",
      "scope": "synthetic phase10 scope.",
      "normative_owners": "Synthetic owner."
    },
    {
      "phase": "phase11",
      "order": 11,
      "status": "active",
      "label": "Phase 11",
      "manifest_path": "tools/phase11_test_map.json",
      "ledger_path": "docs/testing/phase11_coverage_ledger.md",
      "scope": "synthetic phase11 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON

set +e
go_manifest_pkg_setup_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase package setup" phase11 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_pkg_setup_rel" \
    2>&1
)"
go_manifest_pkg_setup_status=$?
set -e
if [[ "$go_manifest_pkg_setup_status" -eq 0 ]]; then
  fail "run-go-manifest-phase package setup: expected non-zero exit status"
fi
assert_contains "$go_manifest_pkg_setup_output" "failure: run-go-manifest-phase package setup" "run-go-manifest-phase package setup label"
assert_contains "$go_manifest_pkg_setup_output" "coverage=authoritative" "run-go-manifest-phase package setup coverage"
assert_contains "$go_manifest_pkg_setup_output" "phase=phase11" "run-go-manifest-phase package setup phase"
assert_contains "$go_manifest_pkg_setup_output" "symbol_or_title=(package setup)" "run-go-manifest-phase package setup title"
assert_contains "$go_manifest_pkg_setup_output" "message=manifest package setup failed" "run-go-manifest-phase package setup message"

for synthetic_manifest in \
  "$ROOT_DIR/tools/phase9_test_map.json" \
  "$ROOT_DIR/tools/phase10_test_map.json" \
  "$ROOT_DIR/tools/phase11_test_map.json"
do
  original_digest="${go_manifest_repo_map_digests["$synthetic_manifest"]}"
  if [[ "$original_digest" == "__absent__" ]]; then
    if [[ -e "$synthetic_manifest" ]]; then
      fail "run-go-manifest-phase smoke must not create synthetic manifests in repo tools/: $synthetic_manifest"
    fi
    continue
  fi
  if [[ ! -e "$synthetic_manifest" ]]; then
    fail "run-go-manifest-phase smoke must not remove repo manifest: $synthetic_manifest"
  fi
  current_digest="$(sha256sum "$synthetic_manifest" | awk '{print $1}')"
  if [[ "$current_digest" != "$original_digest" ]]; then
    fail "run-go-manifest-phase smoke must not mutate repo manifest: $synthetic_manifest"
  fi
done
