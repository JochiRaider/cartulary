#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-phase.sh"
GO_HELPER="$ROOT_DIR/scripts/lib/run-go-phase.sh"
GO_MANIFEST_HELPER="$ROOT_DIR/scripts/lib/run-go-manifest-phase.sh"
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
    "$HELPER" "quiet success" -- bash -lc 'echo hidden-success-output'
)"
assert_empty "$quiet_success_output" "quiet success"

mkdir -p "$ROOT_DIR/tmp"

success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
    "$HELPER" "success log replay" -- bash -lc 'echo keep-this-warning >&2' \
    2>&1
)"
assert_contains "$success_log_output" "keep-this-warning" "success log replay output"
assert_not_contains "$success_log_output" "== success log replay ==" "success log replay banner"

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
assert_equals "$(json_field "$short_failure_summary" "failure_class")" "helper" "short failure class"
assert_equals "$(json_field "$short_failure_summary" "failure_classes.helper")" "1" "short failure helper count"
assert_equals "$(json_field "$short_failure_summary" "failures.0.failure_class")" "helper" "short failure record class"
assert_matches "$(json_field "$short_failure_summary" "start_time")" '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$' "short failure millisecond start time"
assert_matches "$(json_field "$short_failure_summary" "end_time")" '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$' "short failure millisecond end time"

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
  "failure_classes": { "test": 0, "infra": 0, "timing": 0, "artifact": 0, "helper": 0 },
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
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary short-target pass >/dev/null 2>&1
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
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "missing target" fail 0 1 - test-fast-service-backed \
    2>&1
)"
missing_target_status=$?
set -e
assert_equals "$missing_target_status" "1" "missing target run summary status"
assert_contains "$missing_target_output" "non_test_failed=1" "missing target run summary output"
assert_contains "$missing_target_output" "failure_class=artifact" "missing target run summary failure class"
assert_contains "$missing_target_output" "artifact failure: missing target summary: test-fast-service-backed" "missing target run summary headline"
missing_target_summary="$missing_target_results/missing-target/run-summary.json"
assert_equals "$(json_field "$missing_target_summary" "counts.failed")" "1" "missing target failed count"
assert_equals "$(json_field "$missing_target_summary" "counts.non_test")" "1" "missing target non-test count"
assert_equals "$(json_field "$missing_target_summary" "counts.non_test_failed")" "1" "missing target non-test failed count"
assert_equals "$(json_field "$missing_target_summary" "failure_class")" "artifact" "missing target failure class"
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
  "failure_classes": { "test": 0, "infra": 0, "timing": 0, "artifact": 0, "helper": 0 },
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
    "label": "test-services start minio",
    "start_time": "2026-01-01T00:00:01Z",
    "end_time": "2026-01-01T00:00:02Z",
    "duration_ms": 1000,
    "status": "fail"
  }
}
JSON
infra_timing_output="$(
  CARTULARY_TEST_RESULTS_DIR="$infra_timing_results" \
  CARTULARY_TEST_RUN_ID="infra-timing" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary infra-target pass \
    2>&1
)"
assert_contains "$infra_timing_output" "failure_class=infra" "infra timing target failure class"
assert_contains "$infra_timing_output" "tests passed; infra timing failure: test-services start minio" "infra timing target headline"
infra_timing_summary="$infra_timing_results/infra-timing/infra-target/target-summary.json"
assert_equals "$(json_field "$infra_timing_summary" "failure_class")" "infra" "infra timing JSON failure class"
assert_equals "$(json_field "$infra_timing_summary" "failure_classes.infra")" "1" "infra timing JSON class count"
assert_equals "$(json_field "$infra_timing_summary" "failures.0.kind")" "timing" "infra timing JSON failure kind"

skipped_after_failure_results="$(mktemp -d "$ROOT_DIR/tmp/run-summary-skipped-after-failure.XXXXXX")"
cleanup_paths+=("$skipped_after_failure_results")
CARTULARY_TEST_RESULTS_DIR="$skipped_after_failure_results" \
CARTULARY_TEST_RUN_ID="skipped-after-failure" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary failed-check fail >/dev/null 2>&1
set +e
skipped_after_failure_output="$(
  CARTULARY_TEST_RESULTS_DIR="$skipped_after_failure_results" \
  CARTULARY_TEST_RUN_ID="skipped-after-failure" \
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "skipped after failure" fail 0 1 failed-check \
      --summary-groups "harness=failed-check,skipped-check" \
      --skipped-after-failure skipped-check \
      failed-check skipped-check \
    2>&1
)"
skipped_after_failure_status=$?
set -e
assert_equals "$skipped_after_failure_status" "1" "skipped after failure run summary status"
assert_contains "$skipped_after_failure_output" "aborted_after=failed-check" "skipped after failure root cause output"
assert_contains "$skipped_after_failure_output" "skipped_after_failure=skipped-check" "skipped after failure group output"
assert_not_contains "$skipped_after_failure_output" "missing_summary_targets=skipped-check" "skipped after failure missing output"
skipped_after_failure_summary="$skipped_after_failure_results/skipped-after-failure/run-summary.json"
assert_equals "$(json_field "$skipped_after_failure_summary" "counts.failed")" "1" "skipped after failure failed count"
assert_equals "$(json_field "$skipped_after_failure_summary" "counts.non_test_failed")" "1" "skipped after failure non-test failed count"
assert_equals "$(json_field "$skipped_after_failure_summary" "failure_class")" "helper" "skipped after failure class"
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
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary parent-target pass --children child-a,child-b \
    2>&1
)"
assert_contains "$child_target_output" "[PASS] parent-target kind=aggregate children=2/2 child_tests=18 child_failed=0 failed_children=none slowest_child=child-b(2.00s) own_phases=1 own_tests=0 own_failed=0 total_tests=18 total_failed=0" "child target parent output"
assert_contains "$child_target_output" "failed_children=none slowest_child=child-b(2.00s)" "child target compact child hints"
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
  "$ROOT_DIR/scripts/print-explain-run.mjs" --results-dir "$child_summary_results" --run-id child-summary --target parent-target \
    2>&1
)"
assert_contains "$explain_run_summary" "[RUN] missing" "explain-run missing run summary"
assert_contains "$explain_run_summary" "[TARGET] parent-target status=pass kind=aggregate tests=18 failed=0" "explain-run target summary"
assert_contains "$explain_run_summary" "failed_children=none missing_children=none slowest_child=child-b(2.00s)" "explain-run compact child hints"
explain_run_children="$(
  "$ROOT_DIR/scripts/print-explain-run.mjs" --results-dir "$child_summary_results/child-summary" --target parent-target --detail children \
    2>&1
)"
assert_contains "$explain_run_children" "[CHILD] child-a status=pass tests=7 failed=0 duration=1.20s" "explain-run child-a detail"
assert_contains "$explain_run_children" "[CHILD] child-b status=pass tests=11 failed=0 duration=2.00s" "explain-run child-b detail"
set +e
explain_run_logs_output="$(
  "$ROOT_DIR/scripts/print-explain-run.mjs" --results-dir "$child_summary_results/child-summary" --detail logs \
    2>&1
)"
explain_run_logs_status=$?
set -e
assert_equals "$explain_run_logs_status" "1" "explain-run logs requires target status"
assert_contains "$explain_run_logs_output" "DETAIL=logs requires TARGET=<target>" "explain-run logs requires target output"

fixture_results="$(mktemp -d "$ROOT_DIR/tmp/fixture-reporting.XXXXXX")"
cleanup_paths+=("$fixture_results")
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "01" "postgres-db-reset" "fixture-target" 20000 "package_reset" "package-reused" "internal/modules/auth" "TestSlowB"
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "02" "postgres-db-reset" "fixture-target" 15000 "package_reset" "package-reused" "internal/modules/auth" "TestSlowA"
write_fixture_event "$fixture_results" "fixture-run" "fixture-suite" "03" "postgres-db-created" "fixture-target" 1000 "template_clone" "per-test" "internal/modules/entities" "TestClone"
below_fixture_output="$(
  FIXTURE_THRESHOLD_MS=40000 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary fixture-target pass \
    2>&1
)"
assert_not_contains "$below_fixture_output" "[FIXTURE]" "fixture output below threshold"
fixture_target_output="$(
  FIXTURE_THRESHOLD_MS=30000 \
  FIXTURE_TOP=2 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary fixture-target pass \
    2>&1
)"
assert_contains "$fixture_target_output" "[FIXTURE] fixture-target total=36.0s count=3 top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=35.0s slowest=TestSlowB(20.0s),TestSlowA(15.0s)" "fixture target threshold output"
assert_equals "$(json_field "$fixture_results/fixture-run/fixture-target/target-summary.json" "totals.fixture.total_duration_ms")" "36000" "fixture target summary duration"

write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "01" "postgres-db-reset" "fixture-tie" 10000 "package_reset" "package-reused" "internal/modules/auth" "TestResetA"
write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "02" "postgres-db-reset" "fixture-tie" 10000 "package_reset" "package-reused" "internal/modules/auth" "TestResetB"
write_fixture_event "$fixture_results" "fixture-tie-run" "fixture-suite" "03" "postgres-transaction" "fixture-tie" 20000 "transaction" "transaction" "internal/modules/auth" "TestTxn"
fixture_tie_output="$(
  FIXTURE_THRESHOLD_MS=1 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-tie-run" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary fixture-tie pass \
    2>&1
)"
assert_contains "$fixture_tie_output" "top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=20.0s" "fixture strategy tie prefers count"

fixture_run_output="$(
  FIXTURE_THRESHOLD_MS=30000 \
  CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
  CARTULARY_TEST_RUN_ID="fixture-run" \
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "fixture run" pass 1 1 - fixture-target \
    2>&1
)"
assert_contains "$fixture_run_output" "[FIXTURE] fixture run total=36.0s count=3 top_strategy=postgres/database-reset/package_reset/package-reused count=2 duration=35.0s" "fixture run summary output"

mkdir -p "$fixture_results/older-run"
touch -d '2026-01-01T00:00:00Z' "$fixture_results/older-run"
touch -d '2030-01-02T00:00:00Z' "$fixture_results/fixture-run"
fixture_report_output="$(
  "$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results" --threshold-ms 30000 --top 2 \
    2>&1
)"
assert_contains "$fixture_report_output" "[FIXTURE] fixture run total=36.0s count=3" "fixture report newest run aggregate output"
assert_contains "$fixture_report_output" "[FIXTURE] fixture-target total=36.0s count=3" "fixture report newest run target output"
fixture_report_concrete_output="$(
  "$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results/fixture-run" --threshold-ms 30000 --top 2 \
    2>&1
)"
assert_contains "$fixture_report_concrete_output" "[FIXTURE] fixture run total=36.0s count=3" "fixture report concrete run aggregate output"
assert_contains "$fixture_report_concrete_output" "[FIXTURE] fixture-target total=36.0s count=3" "fixture report concrete run target output"
if fixture_report_mismatch_output="$(
  "$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results/fixture-run" --run-id fixture-tie-run --threshold-ms 1 \
    2>&1
)"; then
  fail "fixture report concrete run mismatch: expected failure"
fi
assert_contains "$fixture_report_mismatch_output" "RESULTS_DIR points to run fixture-run, but RUN_ID requested fixture-tie-run" "fixture report concrete run mismatch error"
fixture_report_json="$fixture_results/fixture-report.json"
"$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results" --run-id fixture-run --threshold-ms 30000 --json >"$fixture_report_json"
assert_equals "$(json_field "$fixture_report_json" "schema_id")" "cartulary.fixture_report.v1" "fixture report schema"
assert_equals "$(json_field "$fixture_report_json" "run_id")" "fixture-run" "fixture report run id"
assert_equals "$(json_field "$fixture_report_json" "run_dir")" "$fixture_results/fixture-run" "fixture report run dir"
assert_equals "$(json_field "$fixture_report_json" "aggregate.total_duration_ms")" "36000" "fixture report aggregate duration"
assert_equals "$(json_field "$fixture_report_json" "targets.0.target")" "fixture-target" "fixture report target"

write_fixture_event "$fixture_results" "fixture-aggregate-run" "fixture-suite" "01" "postgres-db-reset" "fixture-child" 32000 "package_reset" "package-reused" "internal/modules/auth" "TestAggregateChild"
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary fixture-child pass >/dev/null 2>&1
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary fixture-parent pass --children fixture-child >/dev/null 2>&1
CARTULARY_TEST_RESULTS_DIR="$fixture_results" \
CARTULARY_TEST_RUN_ID="fixture-aggregate-run" \
  "$ROOT_DIR/scripts/lib/test-output.sh" run-summary check pass 1 1 - fixture-parent >/dev/null 2>&1
fixture_report_run_label_output="$(
  "$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results" --run-id fixture-aggregate-run --target check --threshold-ms 1 \
    2>&1
)"
assert_contains "$fixture_report_run_label_output" "[FIXTURE] check total=32.0s count=1" "fixture report run label target uses run summary"
fixture_report_aggregate_target_output="$(
  "$ROOT_DIR/scripts/print-fixture-report.mjs" --results-dir "$fixture_results" --run-id fixture-aggregate-run --target fixture-parent --threshold-ms 1 \
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
  "$ROOT_DIR/scripts/lib/test-output.sh" timing-span
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
CARTULARY_TIMING_BUCKET="teardown" \
CARTULARY_TIMING_LABEL="browser-e2e overlapping process cleanup" \
CARTULARY_TIMING_START_TIME="2026-01-01T00:00:00.500Z" \
CARTULARY_TIMING_END_TIME="2026-01-01T00:00:01.500Z" \
CARTULARY_TIMING_DURATION_MS="1000" \
  "$ROOT_DIR/scripts/lib/test-output.sh" timing-span
CARTULARY_TEST_RESULTS_DIR="$teardown_accounting_results" \
CARTULARY_TEST_RUN_ID="teardown-accounting" \
CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
CARTULARY_TIMING_BUCKET="teardown" \
CARTULARY_TIMING_LABEL="browser-e2e remove runtime root" \
CARTULARY_TIMING_START_TIME="2026-01-01T00:00:01.800Z" \
CARTULARY_TIMING_END_TIME="2026-01-01T00:00:02.100Z" \
CARTULARY_TIMING_DURATION_MS="300" \
  "$ROOT_DIR/scripts/lib/test-output.sh" timing-span
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
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary browser-e2e-webserver-backed pass >/dev/null 2>&1
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
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "child run" pass 1 1 - \
      --summary-groups "backend-service-backed=child-a,child-b;browser=child-b" \
      parent-target \
    2>&1
)"
assert_contains "$child_run_output" "wall=" "child run wall duration output"
assert_contains "$child_run_output" "critical=" "child run critical path duration output"
assert_contains "$child_run_output" "exec=" "child run executed duration output"
assert_contains "$child_run_output" "logical=" "child run logical duration output"
assert_contains "$child_run_output" "teardown=" "child run teardown duration output"
assert_contains "$child_run_output" "slowest_target=parent-target(" "child run slowest target output"
assert_contains "$child_run_output" "slowest_lifecycle_bucket=parent-target:" "child run slowest lifecycle bucket output"
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
  "$ROOT_DIR/scripts/lib/test-output.sh" shared-execution \
    backend-integration-shards \
    backend-integration-incidents-shard-01 \
    pass \
    2026-01-01T00:00:00Z \
    2026-01-01T00:00:07Z \
    7000 \
    0 \
    "$shared_execution_dir/shared-execution.json"
shared_execution_output="$(
  CARTULARY_TEST_RESULTS_DIR="$shared_execution_results" \
  CARTULARY_TEST_RUN_ID="shared-execution" \
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "shared execution run" pass 2 2 - target-fast target-slow \
    2>&1
)"
assert_contains "$shared_execution_output" "slowest_target=target-slow(2.00s)" "shared execution run slowest target ignores shared group"
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
  CARTULARY_TEST_RESULTS_DIR="$child_summary_results" \
  CARTULARY_TEST_RUN_ID="child-summary" \
    "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "child run missing group" pass 1 1 - \
      --summary-groups "browser=missing-browser" \
      parent-target \
    2>&1
)"
missing_group_status=$?
set -e
assert_equals "$missing_group_status" "1" "missing group run summary status"
assert_contains "$missing_group_output" "[GROUP] child run missing group browser summary_targets=missing-browser status=fail" "missing group output"
assert_contains "$missing_group_output" "missing_summary_targets=missing-browser" "missing group target output"
missing_group_summary="$child_summary_results/child-summary/run-summary.json"
assert_equals "$(json_field "$missing_group_summary" "summary_groups.0.missing_summary_targets.0")" "missing-browser" "missing group summary list"

missing_child_results="$(mktemp -d "$ROOT_DIR/tmp/target-summary-missing-child.XXXXXX")"
cleanup_paths+=("$missing_child_results")
missing_child_output="$(
  CARTULARY_TEST_RESULTS_DIR="$missing_child_results" \
  CARTULARY_TEST_RUN_ID="missing-child" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary parent-with-missing pass --children missing-child \
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

verbose_override_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$HELPER" "verbose override" -- bash -lc 'echo verbose-stream'
)"
assert_contains "$verbose_override_output" "== verbose override ==" "verbose override banner"
assert_contains "$verbose_override_output" "verbose-stream" "verbose override output"

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
    "$GO_HELPER" "run-go-phase smoke" '^(TestPhase0_.*_E_0_[0-9]+)$' -- "$go_bin" test "$go_smoke_rel"
)"
assert_empty "$go_success_output" "run-go-phase success"

set +e
go_zero_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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

go_manifest_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-smoke.XXXXXX")"
go_manifest_root="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-manifests.XXXXXX")"
go_manifest_tools="$go_manifest_root/tools"
mkdir -p "$go_manifest_tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$go_manifest_tools"/
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
  "expected_ids": ["U-9-01"],
  "unit": [
    {
      "id": "U-9-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$go_manifest_rel",
      "file": "${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go",
      "symbol": "TestPhase9_RunGoManifest_U_9_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest smoke",
      "out_of_scope": "synthetic run-go manifest smoke"
    }
  ]
}
EOF
cat >"$go_manifest_tools/phase10_test_map.json" <<EOF
{
  "expected_ids": ["U-10-01"],
  "unit": [
    {
      "id": "U-10-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$go_manifest_rel",
      "file": "${go_manifest_rel#./}/run_go_manifest_phase_smoke_test.go",
      "symbol": "TestPhase10_RunGoManifest_U_10_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest skip smoke",
      "out_of_scope": "synthetic run-go manifest skip smoke"
    }
  ]
}
EOF

node_bin="${NODE_BIN:-node}"
go_manifest_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_PHASE_MANIFEST_ROOT="$go_manifest_root" \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase smoke" phase9 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_rel"
)"
assert_empty "$go_manifest_success_output" "run-go-manifest-phase success"

set +e
go_manifest_empty_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  CARTULARY_OUTPUT_MODE=quiet \
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
  "expected_ids": ["U-11-01"],
  "unit": [
    {
      "id": "U-11-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$go_manifest_pkg_setup_rel",
      "file": "${go_manifest_pkg_setup_rel#./}/run_go_manifest_phase_package_setup_test.go",
      "symbol": "TestPhase11_RunGoManifestPackageSetup_U_11_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "smoke",
      "claim": "synthetic run-go manifest package setup smoke",
      "out_of_scope": "synthetic run-go manifest package setup smoke"
    }
  ]
}
EOF

set +e
go_manifest_pkg_setup_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
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
  if [[ -e "$synthetic_manifest" ]]; then
    fail "run-go-manifest-phase smoke must not write synthetic manifests into repo tools/: $synthetic_manifest"
  fi
done
