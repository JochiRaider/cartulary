#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
GO_STEP_HELPER="$ROOT_DIR/tools/harness/backend/run-go-step.sh"
GO_TARGET_HELPER="$ROOT_DIR/tools/harness/backend/go-target-runner.mjs"
GO_TARGET_PLAN_COVERAGE_HELPER="$ROOT_DIR/tools/harness/backend/go-target-plan-coverage-cli.mjs"
cleanup_paths=()

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

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
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

assert_less_than() {
  local actual="$1"
  local expected_upper_bound="$2"
  local label="$3"

  if [[ -z "$actual" || -z "$expected_upper_bound" || ! "$actual" =~ ^[0-9]+$ || ! "$expected_upper_bound" =~ ^[0-9]+$ || "$actual" -ge "$expected_upper_bound" ]]; then
    fail "$label: expected [$actual] to be less than [$expected_upper_bound]"
  fi
}

node_bin="${NODE_BIN:-node}"

find_planned_shard_for_symbol() {
  local target="$1"
  local symbol="$2"
  local owner="${3:-}"

  "$node_bin" - "$ROOT_DIR" "$target" "$symbol" "$owner" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root, target, symbol, owner] = process.argv.slice(2);
const args = owner
  ? [path.join(root, "tools/harness/backend/go-shard-plan.mjs"), "--owner", owner, "json"]
  : [path.join(root, "tools/harness/backend/go-shard-plan-cli.mjs"), "--json", "--target", target];
const plan = JSON.parse(execFileSync(process.execPath, args, { encoding: "utf8", cwd: root }));
const shard = plan.shards.find(
  (candidate) => candidate.target === target && candidate.items.some((item) => item.symbol === symbol),
);
if (!shard) {
  process.exit(1);
}
process.stdout.write(shard.name);
EOF
}
go_bin="${GO:-go}"

verbose_go_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-verbose.XXXXXX")"
cleanup_paths+=("$verbose_go_dir")
cat >"$verbose_go_dir/run_go_target_verbose_test.go" <<'EOF'
package rungotargetverbose

import "testing"

func TestSupportSemanticIntegration_VerboseSmoke(t *testing.T) {
	t.Log("support verbose line")
}
EOF

verbose_go_rel="./${verbose_go_dir#"$ROOT_DIR"/}"
explicit_quiet_verbose_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$GO_STEP_HELPER" "run-go-target verbose smoke" '^(TestSupportSemanticIntegration_VerboseSmoke)$' -- "$go_bin" test "$verbose_go_rel" \
    2>&1
)"
assert_not_contains "$explicit_quiet_verbose_output" "== run-go-target verbose smoke ==" "explicit quiet suppresses verbose go banner"
assert_not_contains "$explicit_quiet_verbose_output" "support verbose line" "explicit quiet suppresses verbose go human output"

verbose_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  VERBOSE=1 \
    "$GO_STEP_HELPER" "run-go-target verbose smoke" '^(TestSupportSemanticIntegration_VerboseSmoke)$' -- "$go_bin" test "$verbose_go_rel" \
    2>&1
)"
assert_contains "$verbose_output" "== run-go-target verbose smoke ==" "verbose go banner"
assert_contains "$verbose_output" "support verbose line" "verbose go human output"
assert_not_contains "$verbose_output" "\"Action\":\"output\"" "verbose go raw json output"
assert_not_contains "$verbose_output" "\"Test\":\"TestSupportSemanticIntegration_VerboseSmoke\"" "verbose go raw json test field"

duration_results_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-durations.XXXXXX")"
cleanup_paths+=("$duration_results_dir")
shared_report_dir="$duration_results_dir/shared-report"
mkdir -p "$shared_report_dir"
cat >"$shared_report_dir/runner.jsonl" <<'EOF'
{"Time":"2000-01-01T00:00:00Z","Action":"run","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportSemanticIntegration_Smoke"}
{"Time":"2000-01-01T00:00:00Z","Action":"output","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportSemanticIntegration_Smoke","Output":"=== RUN   TestSupportSemanticIntegration_Smoke\n"}
{"Time":"2000-01-01T00:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportSemanticIntegration_Smoke","Elapsed":0.001}
{"Time":"2000-01-01T00:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Elapsed":0.001}
EOF
touch "$shared_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(TestSupportSemanticIntegration_Smoke)$' ./internal/modules/entities" >"$shared_report_dir/command.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$shared_report_dir/start_time.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$shared_report_dir/end_time.txt"
printf '%s\n' "1200" >"$shared_report_dir/duration_ms.txt"
printf '%s\n' "0" >"$shared_report_dir/exit_status.txt"

duration_artifacts_root="$duration_results_dir/results"
# These fixture env vars are intentionally scoped to the subshell that emits
# step summaries.
# shellcheck disable=SC2030,SC2031
(
  export CARTULARY_TEST_RESULTS_DIR="$duration_artifacts_root"
  export CARTULARY_TEST_RUN_ID="duration-smoke"
  export CARTULARY_TEST_TARGET="backend-unit-smoke"
  export NODE_BIN="$node_bin"
  emit_go_step_summary() {
    local label="$1"
    local mode="$2"
    local duration_ms="$3"
    local wall_duration_ms="$4"
    local executed_duration_ms="$duration_ms"
    local start_time="${5:-$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)}"
    local end_time="${6:-$start_time}"
    local step_slug
    local step_dir
    if [[ "$mode" != "actual" ]]; then
      executed_duration_ms=0
    fi
    step_slug="$(printf '%s' "$label" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')"
    step_dir="$duration_artifacts_root/duration-smoke/backend-unit-smoke/$step_slug"
    mkdir -p "$step_dir"
    CARTULARY_STEP_LABEL="$label" \
    CARTULARY_STEP_DIR="$step_dir" \
    CARTULARY_STEP_COMMAND="$(<"$shared_report_dir/command.txt")" \
    CARTULARY_STEP_START_TIME="$start_time" \
    CARTULARY_STEP_END_TIME="$end_time" \
    CARTULARY_STEP_LOGICAL_DURATION_MS="$duration_ms" \
    CARTULARY_STEP_EXECUTED_DURATION_MS="$executed_duration_ms" \
    CARTULARY_STEP_WALL_DURATION_MS="$wall_duration_ms" \
    CARTULARY_STEP_EXIT_STATUS="0" \
    CARTULARY_REPORT_SLICE=1 \
    CARTULARY_STEP_ACCOUNTING_MODE="$mode" \
    CARTULARY_STEP_RUNNER_LOG="$shared_report_dir/runner.jsonl" \
    CARTULARY_STEP_STDERR_LOG="$shared_report_dir/stderr.log" \
    CARTULARY_GO_TEST_REGEX='^(TestSupportSemanticIntegration_Smoke)$' \
    CARTULARY_ACCOUNTING_COVERAGE=raw \
    CARTULARY_GO_PACKAGE_PATTERNS="./internal/modules/entities" \
      "$ROOT_DIR/tools/harness/output/test-output.sh" go-step >/dev/null
  }
  emit_go_step_summary "duration actual" actual 1200 1200 "2000-01-01T00:00:00Z" "2000-01-01T00:00:00Z"
  emit_go_step_summary "duration reused" reused 1200 0
  emit_go_step_summary "duration derived" derived 0 0
  "$ROOT_DIR/tools/harness/output/test-output.sh" target-summary backend-unit-smoke pass >/dev/null
  "$ROOT_DIR/tools/harness/output/test-output.sh" run-summary "duration smoke" pass 1 1 - backend-unit-smoke >/dev/null
)

duration_actual_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-actual/step-summary.json"
duration_reused_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-reused/step-summary.json"
duration_derived_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-derived/step-summary.json"
duration_target_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/target-summary.json"
duration_target_timing="$duration_artifacts_root/duration-smoke/backend-unit-smoke/target-timing.json"
duration_run_summary="$duration_artifacts_root/duration-smoke/run-summary.json"

assert_json_field_absent "$duration_actual_summary" "duration_ms" "duration actual legacy duration"
assert_not_negative "$(json_field "$duration_actual_summary" "wall_duration_ms")" "duration actual step wall duration"
assert_not_negative "$(json_field "$duration_actual_summary" "critical_path_wall_duration_ms")" "duration actual step critical path duration"
assert_equals "$(json_field "$duration_actual_summary" "accounting_mode")" "actual" "duration actual accounting mode"
assert_equals "$(json_field "$duration_actual_summary" "counts.raw")" "1" "duration actual raw count"
assert_equals "$(json_field "$duration_actual_summary" "counts.support")" "0" "duration actual support count"
assert_equals "$(json_field "$duration_actual_summary" "executed_duration_ms")" "1200" "duration actual executed duration"
assert_equals "$(json_field "$duration_actual_summary" "logical_duration_ms")" "1200" "duration actual logical duration"
assert_equals "$(json_field "$duration_actual_summary" "reused_duration_ms")" "0" "duration actual reused duration"
assert_equals "$(json_field "$duration_actual_summary" "derived_duration_ms")" "0" "duration actual derived duration"
assert_equals "$(json_field "$duration_actual_summary" "wall_duration_ms")" "1200" "duration actual wall duration"
assert_json_field_absent "$duration_reused_summary" "duration_ms" "duration reused legacy duration"
assert_not_negative "$(json_field "$duration_reused_summary" "wall_duration_ms")" "duration reused step wall duration"
assert_not_negative "$(json_field "$duration_reused_summary" "critical_path_wall_duration_ms")" "duration reused step critical path duration"
assert_equals "$(json_field "$duration_reused_summary" "accounting_mode")" "reused" "duration reused accounting mode"
assert_equals "$(json_field "$duration_reused_summary" "executed_duration_ms")" "0" "duration reused executed duration"
assert_equals "$(json_field "$duration_reused_summary" "logical_duration_ms")" "1200" "duration reused logical duration"
assert_equals "$(json_field "$duration_reused_summary" "reused_duration_ms")" "1200" "duration reused reused duration"
assert_equals "$(json_field "$duration_reused_summary" "derived_duration_ms")" "0" "duration reused derived duration"
assert_equals "$(json_field "$duration_reused_summary" "wall_duration_ms")" "0" "duration reused wall duration"
assert_json_field_absent "$duration_derived_summary" "duration_ms" "duration derived legacy duration"
assert_not_negative "$(json_field "$duration_derived_summary" "wall_duration_ms")" "duration derived step wall duration"
assert_not_negative "$(json_field "$duration_derived_summary" "critical_path_wall_duration_ms")" "duration derived step critical path duration"
assert_equals "$(json_field "$duration_derived_summary" "accounting_mode")" "derived" "duration derived accounting mode"
assert_equals "$(json_field "$duration_derived_summary" "executed_duration_ms")" "0" "duration derived executed duration"
assert_equals "$(json_field "$duration_derived_summary" "logical_duration_ms")" "0" "duration derived logical duration"
assert_equals "$(json_field "$duration_derived_summary" "reused_duration_ms")" "0" "duration derived reused duration"
assert_equals "$(json_field "$duration_derived_summary" "derived_duration_ms")" "0" "duration derived derived duration"
assert_equals "$(json_field "$duration_derived_summary" "wall_duration_ms")" "0" "duration derived wall duration"
assert_json_field_absent "$duration_target_summary" "duration_ms" "duration target legacy duration"
assert_not_negative "$(json_field "$duration_target_summary" "wall_duration_ms")" "duration target wall duration"
assert_not_negative "$(json_field "$duration_target_summary" "critical_path_wall_duration_ms")" "duration target critical path duration"
assert_equals "$(json_field "$duration_target_summary" "executed_duration_ms")" "1200" "duration target executed duration"
assert_equals "$(json_field "$duration_target_summary" "logical_duration_ms")" "2400" "duration target logical duration"
assert_equals "$(json_field "$duration_target_summary" "reused_duration_ms")" "1200" "duration target reused duration"
assert_equals "$(json_field "$duration_target_summary" "derived_duration_ms")" "0" "duration target derived duration"
assert_equals "$(json_field "$duration_target_summary" "wall_duration_ms")" "1200" "duration target wall duration"
assert_equals "$(json_field "$duration_target_summary" "start_time")" "2000-01-01T00:00:00Z" "duration target summary start time"
assert_equals "$(json_field "$duration_target_summary" "end_time")" "2000-01-01T00:00:00Z" "duration target summary end time"
assert_equals "$(json_field "$duration_target_summary" "accounting_modes.actual")" "1" "duration target actual accounting count"
assert_equals "$(json_field "$duration_target_summary" "accounting_modes.reused")" "1" "duration target reused accounting count"
assert_equals "$(json_field "$duration_target_summary" "accounting_modes.derived")" "1" "duration target derived accounting count"
assert_equals "$(json_field "$duration_target_summary" "own.counts.raw")" "3" "duration target raw count"
assert_equals "$(json_field "$duration_target_summary" "own.counts.unmapped")" "0" "duration target unmapped count"
assert_contains "$(json_field "$duration_target_summary" "artifacts.timing_json")" "target-timing.json" "duration target timing artifact"
assert_equals "$(json_field "$duration_target_timing" "schema_id")" "cartulary.test_target_timing.v1" "duration target timing schema"
assert_equals "$(json_field "$duration_target_timing" "start_time")" "2000-01-01T00:00:00Z" "duration target timing start time"
assert_equals "$(json_field "$duration_target_timing" "end_time")" "2000-01-01T00:00:00Z" "duration target timing end time"
assert_equals "$(json_field "$duration_target_timing" "buckets.0.name")" "test_command" "duration target test command bucket"
assert_equals "$(json_field "$duration_target_timing" "buckets.1.name")" "report_collation" "duration target report collation bucket"
if "${NODE:-node}" -e 'const fs=require("node:fs"); const buckets=JSON.parse(fs.readFileSync(process.argv[1],"utf8")).buckets.map((bucket)=>bucket.name); process.exit(buckets.includes("setup") ? 1 : 0);' "$duration_target_timing"; then
  :
else
  fail "duration target timing must omit setup bucket"
fi
assert_json_field_absent "$duration_run_summary" "duration_ms" "duration run legacy duration"
assert_not_negative "$(json_field "$duration_run_summary" "wall_duration_ms")" "duration run wall duration"
assert_not_negative "$(json_field "$duration_run_summary" "critical_path_wall_duration_ms")" "duration run critical path duration"
assert_equals "$(json_field "$duration_run_summary" "executed_duration_ms")" "1200" "duration run executed duration"
assert_equals "$(json_field "$duration_run_summary" "logical_duration_ms")" "2400" "duration run logical duration"
assert_equals "$(json_field "$duration_run_summary" "reused_duration_ms")" "1200" "duration run reused duration"
assert_equals "$(json_field "$duration_run_summary" "derived_duration_ms")" "0" "duration run derived duration"
assert_equals "$(json_field "$duration_run_summary" "accounting_modes.actual")" "1" "duration run actual accounting count"
assert_equals "$(json_field "$duration_run_summary" "accounting_modes.reused")" "1" "duration run reused accounting count"
assert_equals "$(json_field "$duration_run_summary" "accounting_modes.derived")" "1" "duration run derived accounting count"

raw_failure_results_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-raw-failure.XXXXXX")"
cleanup_paths+=("$raw_failure_results_dir")
raw_failure_report_dir="$raw_failure_results_dir/shared-report"
mkdir -p "$raw_failure_report_dir"
cat >"$raw_failure_report_dir/runner.jsonl" <<'EOF'
{"Time":"2000-01-01T00:00:00Z","Action":"output","Package":"github.com/JochiRaider/cartulary/internal/testutil/configtest","Output":"setup failed before test attribution\n"}
{"Time":"2000-01-01T00:00:00Z","Action":"fail","Package":"github.com/JochiRaider/cartulary/internal/testutil/configtest","Elapsed":0.001}
EOF
touch "$raw_failure_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(Test)$' ./internal/testutil/configtest" >"$raw_failure_report_dir/command.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$raw_failure_report_dir/start_time.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$raw_failure_report_dir/end_time.txt"
printf '%s\n' "100" >"$raw_failure_report_dir/duration_ms.txt"
printf '%s\n' "1" >"$raw_failure_report_dir/exit_status.txt"
# These fixture env vars are intentionally scoped to the subshell that emits
# step summaries.
# shellcheck disable=SC2030,SC2031
(
  export CARTULARY_TEST_RESULTS_DIR="$raw_failure_results_dir/results"
  export CARTULARY_TEST_RUN_ID="raw-failure"
  export CARTULARY_TEST_TARGET="backend-unit-configtest"
  export NODE_BIN="$node_bin"
  raw_step_dir="$raw_failure_results_dir/results/raw-failure/backend-unit-configtest/backend-unit-configtest"
  mkdir -p "$raw_step_dir"
  CARTULARY_STEP_LABEL="backend-unit configtest" \
  CARTULARY_STEP_DIR="$raw_step_dir" \
  CARTULARY_STEP_COMMAND="$(<"$raw_failure_report_dir/command.txt")" \
  CARTULARY_STEP_START_TIME="2000-01-01T00:00:00Z" \
  CARTULARY_STEP_END_TIME="2000-01-01T00:00:00Z" \
  CARTULARY_STEP_LOGICAL_DURATION_MS="100" \
  CARTULARY_STEP_EXECUTED_DURATION_MS="100" \
  CARTULARY_STEP_WALL_DURATION_MS="100" \
  CARTULARY_STEP_EXIT_STATUS="1" \
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_STEP_ACCOUNTING_MODE=actual \
  CARTULARY_STEP_RUNNER_LOG="$raw_failure_report_dir/runner.jsonl" \
  CARTULARY_STEP_STDERR_LOG="$raw_failure_report_dir/stderr.log" \
  CARTULARY_GO_TEST_REGEX='^(Test)$' \
  CARTULARY_ACCOUNTING_COVERAGE=raw \
  CARTULARY_GO_PACKAGE_PATTERNS="./internal/testutil/configtest" \
    "$ROOT_DIR/tools/harness/output/test-output.sh" go-step >/dev/null 2>&1 || true
)
raw_failure_summary="$raw_failure_results_dir/results/raw-failure/backend-unit-configtest/backend-unit-configtest/step-summary.json"
assert_equals "$(json_field "$raw_failure_summary" "counts.failed")" "1" "raw package setup failed count"
assert_equals "$(json_field "$raw_failure_summary" "counts.raw_failed")" "1" "raw package setup raw failed count"
assert_equals "$(json_field "$raw_failure_summary" "counts.unmapped_failed")" "0" "raw package setup unmapped failed count"

identity_reuse_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-identity-reuse.XXXXXX")"
cleanup_paths+=("$identity_reuse_results")
mkdir -p "$identity_reuse_results/results/run-a"
printf 'prepared\n' >"$identity_reuse_results/results/run-a/preflight.txt"
identity_reuse_rejected="$(
  set +e
  CARTULARY_TEST_RESULTS_DIR="$identity_reuse_results/results" \
  CARTULARY_TEST_RUN_ID="run-a" \
  CARTULARY_TEST_TARGET="backend-integration" \
    "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration module.revisions.integration 2>&1
  printf 'status=%s\n' "$?"
)"
assert_contains "$identity_reuse_rejected" "non-empty run root" "public go target rejects unprepared non-empty run root"
assert_contains "$identity_reuse_rejected" "status=1" "public go target unprepared reuse exits non-zero"
identity_reuse_allowed="$(
  CARTULARY_TEST_RESULTS_DIR="$identity_reuse_results/results" \
  CARTULARY_TEST_RUN_ID="run-a" \
  CARTULARY_TEST_TARGET="backend-integration" \
  CARTULARY_HARNESS_IDENTITY_PREPARED=1 \
    "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration module.revisions.integration
)"
assert_contains "$identity_reuse_allowed" "go test" "public go target prepared identity reuse inspects command"

reused_window_results_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-reused-window.XXXXXX")"
cleanup_paths+=("$reused_window_results_dir")
reused_window_report_dir="$reused_window_results_dir/shared-report"
mkdir -p "$reused_window_report_dir"
cp "$shared_report_dir/runner.jsonl" "$reused_window_report_dir/runner.jsonl"
touch "$reused_window_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(TestSupportSemanticIntegration_Smoke)$' ./internal/modules/entities" >"$reused_window_report_dir/command.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$reused_window_report_dir/start_time.txt"
printf '%s\n' "2000-01-01T00:00:10Z" >"$reused_window_report_dir/end_time.txt"
printf '%s\n' "10000" >"$reused_window_report_dir/duration_ms.txt"
printf '%s\n' "0" >"$reused_window_report_dir/exit_status.txt"
# These fixture env vars are intentionally scoped to the subshell that emits
# step summaries.
# shellcheck disable=SC2030,SC2031
(
  export CARTULARY_TEST_RESULTS_DIR="$reused_window_results_dir/results"
  export CARTULARY_TEST_RUN_ID="reused-window"
  export CARTULARY_TEST_TARGET="backend-integration-support"
  export NODE_BIN="$node_bin"
  reused_step_dir="$reused_window_results_dir/results/reused-window/backend-integration-support/reused-window-support"
  mkdir -p "$reused_step_dir"
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)"
  CARTULARY_STEP_LABEL="reused window support" \
  CARTULARY_STEP_DIR="$reused_step_dir" \
  CARTULARY_STEP_COMMAND="$(<"$reused_window_report_dir/command.txt")" \
  CARTULARY_STEP_START_TIME="$timestamp" \
  CARTULARY_STEP_END_TIME="$timestamp" \
  CARTULARY_STEP_LOGICAL_DURATION_MS="10000" \
  CARTULARY_STEP_EXECUTED_DURATION_MS="0" \
  CARTULARY_STEP_WALL_DURATION_MS="0" \
  CARTULARY_STEP_EXIT_STATUS="0" \
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_STEP_ACCOUNTING_MODE=reused \
  CARTULARY_STEP_RUNNER_LOG="$reused_window_report_dir/runner.jsonl" \
  CARTULARY_STEP_STDERR_LOG="$reused_window_report_dir/stderr.log" \
  CARTULARY_GO_TEST_REGEX='^(TestSupportSemanticIntegration_Smoke)$' \
  CARTULARY_ACCOUNTING_COVERAGE=support \
  CARTULARY_GO_PACKAGE_PATTERNS="./internal/modules/entities" \
    "$ROOT_DIR/tools/harness/output/test-output.sh" go-step >/dev/null
  CARTULARY_TIMING_BUCKET=test_command \
  CARTULARY_TIMING_LABEL="run-go-target backend-integration-support" \
  CARTULARY_TIMING_START_TIME="2026-01-01T00:00:00Z" \
  CARTULARY_TIMING_END_TIME="2026-01-01T00:00:00.400Z" \
  CARTULARY_TIMING_DURATION_MS=400 \
  CARTULARY_TIMING_STATUS=pass \
    "$ROOT_DIR/tools/harness/output/test-output.sh" timing-span >/dev/null
  "$ROOT_DIR/tools/harness/output/test-output.sh" target-summary backend-integration-support pass >/dev/null
)
reused_window_summary="$reused_window_results_dir/results/reused-window/backend-integration-support/target-summary.json"
assert_equals "$(json_field "$reused_window_summary" "accounting_modes.actual")" "0" "reused window actual accounting count"
assert_equals "$(json_field "$reused_window_summary" "accounting_modes.reused")" "1" "reused window reused accounting count"
assert_equals "$(json_field "$reused_window_summary" "wall_duration_ms")" "400" "reused window target wall follows invocation span"
assert_equals "$(json_field "$reused_window_summary" "start_time")" "2026-01-01T00:00:00Z" "reused window target start follows invocation span"
assert_equals "$(json_field "$reused_window_summary" "end_time")" "2026-01-01T00:00:00.400Z" "reused window target end follows invocation span"

NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_PLAN_COVERAGE_HELPER" --root "$ROOT_DIR" --commands --quiet

bootstrap_platform_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support platform.postgres.support_integration
)"
assert_contains "$bootstrap_platform_shared_command" "TestEntityAliasMigration31" "backend-integration platform support selector"
bootstrap_platform_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration platform.postgres.integration
)"
assert_contains "$bootstrap_platform_authoritative_command" "TestSchemaBootstrap_Integration" "backend-integration platform authoritative selector"
assert_not_contains "$bootstrap_platform_authoritative_command" "TestFirstAdminBootstrap_Integration" "backend-integration platform excludes app selector"

bootstrap_app_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration app.server.integration
)"
assert_contains "$bootstrap_app_shared_command" "TestFirstAdminBootstrap_Integration" "backend-integration app selector"
assert_not_contains "$bootstrap_app_shared_command" "TestEntityAliasMigration31" "backend-integration app excludes platform support selector"

incidents_incidents_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support module.incidents.support_integration
)"
assert_contains "$incidents_incidents_shared_command" "TestControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied" "backend-integration incidents support selector"
incidents_incidents_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration module.incidents.integration
)"
assert_contains "$incidents_incidents_authoritative_command" "TestIncidentCreatePersistsBootstrapStateAndRollsBackAtomically_Integration" "backend-integration incidents authoritative selector"

incidents_incidents_shard="$(find_planned_shard_for_symbol backend-integration TestIncidentCreatePersistsBootstrapStateAndRollsBackAtomically_Integration)"
incidents_incidents_shard_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration "$incidents_incidents_shard"
)"
assert_contains "$incidents_incidents_shard_command" "TestIncidentCreatePersistsBootstrapStateAndRollsBackAtomically_Integration" "backend-integration incidents planned shard selector"

incidents_incidents_support_shard="$(find_planned_shard_for_symbol backend-integration-support TestControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied)"
incidents_incidents_support_shard_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support "$incidents_incidents_support_shard"
)"
assert_contains "$incidents_incidents_support_shard_command" "TestControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied" "backend-integration support incidents planned shard selector"

recovery_operator_shard="$(find_planned_shard_for_symbol backend-process TestCanonicalOperatorRestoreVerifyLatest_Process)"
recovery_operator_shard_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-process "$recovery_operator_shard"
)"
assert_contains "$recovery_operator_shard_command" "TestCanonicalOperatorRestoreVerifyLatest_Process" "backend-process recovery operator shard selector"
assert_contains "$recovery_operator_shard_command" "TestCanonicalOperatorRestoreVerifyDue_Process" "backend-process recovery operator compatible peer batch"

runtime_binary_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-runtime-binary.XXXXXX")"
cleanup_paths+=("$runtime_binary_results")
runtime_binary_output="$(
  NODE_BIN="$node_bin" "$node_bin" --input-type=module - "$ROOT_DIR" "$runtime_binary_results" <<'EOF_NODE'
import { createHash } from "node:crypto";
import {
  chmodSync,
  mkdirSync,
  readFileSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import {
  runtimeBinaryIDsForRows,
  validateRuntimeBinaries,
} from "./tools/harness/backend/target-execution/runtime-binaries.mjs";

const [root, tmp] = process.argv.slice(2);
const resultsRoot = path.join(tmp, "results");
const runId = "runtime-binary";
const reportDir = path.join(tmp, "report");
const binary = path.join(tmp, "operator-bin");
const symlink = path.join(tmp, "operator-link");
const ctxBase = { repoRoot: root, resultsRoot, runId };

function digest(value) {
  return `sha256:${createHash("sha256").update(value).digest("hex")}`;
}

function fileDigest(file) {
  return digest(readFileSync(file));
}

function relToRepo(file) {
  return path.relative(root, file).replaceAll(path.sep, "/");
}

function buildArtifactOutputDigest(file) {
  return digest(`output\t${relToRepo(file)}\t${fileDigest(file)}\n`);
}

function writeBuildArtifact(outputDigest) {
  const dir = path.join(resultsRoot, runId, "build-operator");
  mkdirSync(dir, { recursive: true });
  writeFileSync(
    path.join(dir, "build-artifact-cache-build-operator.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.cache.build_artifact.v1",
        output_digest_sha256: outputDigest,
      },
      null,
      2,
    )}\n`,
  );
}

function expectRuntimeError(label, env, expectedExitCode, expectedReason) {
  try {
    validateRuntimeBinaries({ ...ctxBase, env }, [{ runtime_binaries: ["operator"] }], reportDir);
  } catch (error) {
    console.log(`${label}=${error.exitCode}:${error.reason}`);
    if (error.exitCode !== expectedExitCode || error.reason !== expectedReason) {
      throw error;
    }
    return;
  }
  throw new Error(`${label} did not fail`);
}

mkdirSync(reportDir, { recursive: true });
writeFileSync(binary, "#!/usr/bin/env sh\nexit 0\n");
chmodSync(binary, 0o755);

console.log("ids=" + runtimeBinaryIDsForRows([
  { runtime_binaries: ["operator"] },
  { runtime_binaries: ["operator"] },
]).join(","));
expectRuntimeError("missing_env", {}, 2, "configuration_error");
expectRuntimeError("missing_artifact", { CARTULARY_OPERATOR_BIN: binary }, 11, "artifact_error");
writeBuildArtifact("sha256:mismatch");
expectRuntimeError("digest_mismatch", { CARTULARY_OPERATOR_BIN: binary }, 11, "artifact_error");
chmodSync(binary, 0o600);
expectRuntimeError("non_executable", { CARTULARY_OPERATOR_BIN: binary }, 2, "configuration_error");
chmodSync(binary, 0o755);
symlinkSync(binary, symlink);
expectRuntimeError("symlink", { CARTULARY_OPERATOR_BIN: symlink }, 2, "configuration_error");
writeBuildArtifact(buildArtifactOutputDigest(binary));
const records = validateRuntimeBinaries(
  { ...ctxBase, env: { CARTULARY_OPERATOR_BIN: binary } },
  [{ runtime_binaries: ["operator"] }],
  reportDir,
);
console.log("valid_count=" + records.length);
console.log("valid_source=" + records[0].source);
console.log("valid_ref=" + records[0].build_artifact_ref);
EOF_NODE
)"
assert_contains "$runtime_binary_output" "ids=operator" "runtime binary ids are deduplicated"
assert_contains "$runtime_binary_output" "missing_env=2:configuration_error" "runtime binary missing env classification"
assert_contains "$runtime_binary_output" "missing_artifact=11:artifact_error" "runtime binary missing build artifact classification"
assert_contains "$runtime_binary_output" "digest_mismatch=11:artifact_error" "runtime binary digest mismatch classification"
assert_contains "$runtime_binary_output" "non_executable=2:configuration_error" "runtime binary executable classification"
assert_contains "$runtime_binary_output" "symlink=2:configuration_error" "runtime binary symlink classification"
assert_contains "$runtime_binary_output" "valid_count=1" "runtime binary valid record count"
assert_contains "$runtime_binary_output" "valid_source=scheduler-produced" "runtime binary valid record source"
assert_contains "$runtime_binary_output" "valid_ref=tmp/run-go-target-runtime-binary." "runtime binary valid artifact ref"
runtime_binary_json="$runtime_binary_results/report/runtime-binaries.json"
assert_equals "$(json_field "$runtime_binary_json" "runtime_binaries.0.id")" "operator" "runtime binary JSON id"
assert_equals "$(json_field "$runtime_binary_json" "runtime_binaries.0.producer_target")" "build-operator" "runtime binary JSON producer"
assert_equals "$(json_field "$runtime_binary_json" "runtime_binaries.0.consumer_env")" "CARTULARY_OPERATOR_BIN" "runtime binary JSON consumer env"
assert_equals "$(json_field "$runtime_binary_json" "runtime_binaries.0.source")" "scheduler-produced" "runtime binary JSON source"
assert_contains "$(json_field "$runtime_binary_json" "runtime_binaries.0.path")" "tmp/run-go-target-runtime-binary." "runtime binary JSON path"
assert_contains "$(json_field "$runtime_binary_json" "runtime_binaries.0.sha256")" "sha256:" "runtime binary JSON digest"
assert_contains "$(json_field "$runtime_binary_json" "runtime_binaries.0.build_artifact_output_digest")" "sha256:" "runtime binary JSON build artifact digest"

"$node_bin" - "$ROOT_DIR" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(process.execPath, [path.join(root, "tools/harness/backend/go-shard-plan.mjs"), "json"], { encoding: "utf8", cwd: root }));
const mixed = plan.shards.filter((shard) => shard.has_authoritative && shard.has_support);
const shared = plan.shards.filter((shard) => shard.shared_across_targets);
if (mixed.length > 0 || shared.length > 0) {
  throw new Error(`backend integration shards must keep authoritative/support ownership separate; mixed=${mixed.map((shard) => shard.name).join(",")} shared=${shared.map((shard) => shard.name).join(",")}`);
}
EOF

entities_entities_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support module.entities.support_integration
)"
assert_contains "$entities_entities_shared_command" "TestAuthorizationReDerivation" "backend-integration entities support selector"
entities_entities_authoritative_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration module.entities.integration
)"
assert_contains "$entities_entities_authoritative_command" "TestResolveRoute_Integration" "backend-integration entities authoritative selector"
assert_not_contains "$entities_entities_authoritative_command" "TestAutoResolutionEligibility_Integration" "backend-integration entities excludes timeline selector"

entities_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration module.timeline.integration
)"
assert_contains "$entities_timeline_shared_command" "TestAutoResolutionEligibility_Integration" "backend-integration timeline selector"
assert_not_contains "$entities_timeline_shared_command" "TestAuthorizationReDerivation" "backend-integration timeline excludes entities support selector"

shared_mismatch_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-mismatch.XXXXXX")"
cleanup_paths+=("$shared_mismatch_results")
set +e
mismatch_output="$(
  CARTULARY_TEST_RESULTS_DIR="$shared_mismatch_results/results" \
  CARTULARY_TEST_RUN_ID="shared-mismatch" \
  NODE_BIN="$node_bin" \
  "$node_bin" --input-type=module - "$ROOT_DIR" <<'EOF_NODE' 2>&1
import path from "node:path";
import { mkdirSync, writeFileSync } from "node:fs";
import { captureGoReport, createGoTargetContext, prepareSharedArtifactDir } from "./tools/harness/backend/backend-target-execution.mjs";

const root = process.argv[2];
const ctx = createGoTargetContext({ repoRoot: root });
const sharedDir = prepareSharedArtifactDir(ctx, "backend-integration-app");
mkdirSync(sharedDir, { recursive: true });
writeFileSync(path.join(sharedDir, "command.txt"), "env go test -json -run '^TestOld$' ./internal/app\n");
writeFileSync(path.join(sharedDir, "complete"), "");
await captureGoReport(ctx, "backend-integration-app", "^TestCurrent$", ["./internal/app"]);
EOF_NODE
)"
mismatch_status=$?
set -e
if [[ "$mismatch_status" -eq 0 ]]; then
  fail "shared command mismatch: expected captureGoReport to fail"
fi
assert_contains "$mismatch_output" "shared_go_report_command_mismatch" "shared command mismatch marker"

mixed_aggregate_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-mixed-aggregate.XXXXXX")"
cleanup_paths+=("$mixed_aggregate_results")
mixed_aggregate_output="$(
  CARTULARY_TEST_RESULTS_DIR="$mixed_aggregate_results/results" \
  CARTULARY_TEST_RUN_ID="mixed-aggregate" \
  NODE_BIN="$node_bin" \
  "$node_bin" --input-type=module - "$ROOT_DIR" "$mixed_aggregate_results" <<'EOF_NODE'
import path from "node:path";
import { mkdirSync, writeFileSync, readFileSync } from "node:fs";
import { createAggregateReport, createGoTargetContext } from "./tools/harness/backend/backend-target-execution.mjs";

const [root, tmp] = process.argv.slice(2);
const ctx = createGoTargetContext({ repoRoot: root });
const metadataDir = path.join(tmp, "metadata");
for (const shard of ["actual-shard", "reused-shard"]) {
  const dir = path.join(metadataDir, shard);
  mkdirSync(dir, { recursive: true });
  writeFileSync(path.join(dir, "command.txt"), "env go test -json -run '^Test$' ./internal/modules/entities\n");
  writeFileSync(path.join(dir, "runner.jsonl"), "");
  writeFileSync(path.join(dir, "stderr.log"), "");
  writeFileSync(path.join(dir, "exit_status.txt"), "0\n");
}
writeFileSync(path.join(metadataDir, "reused-shard", "start_time.txt"), "2026-01-01T00:00:00Z\n");
writeFileSync(path.join(metadataDir, "reused-shard", "end_time.txt"), "2026-01-01T00:00:10Z\n");
writeFileSync(path.join(metadataDir, "reused-shard", "duration_ms.txt"), "10000\n");
writeFileSync(path.join(metadataDir, "actual-shard", "start_time.txt"), "2026-01-01T00:00:20Z\n");
writeFileSync(path.join(metadataDir, "actual-shard", "end_time.txt"), "2026-01-01T00:00:22Z\n");
writeFileSync(path.join(metadataDir, "actual-shard", "duration_ms.txt"), "2000\n");
writeFileSync(path.join(metadataDir, "reused-shard.meta"), path.join(metadataDir, "reused-shard") + "\nreused\n");
writeFileSync(path.join(metadataDir, "actual-shard.meta"), path.join(metadataDir, "actual-shard") + "\nactual\n");
const report = createAggregateReport(ctx, metadataDir, "mixed-aggregate", "backend-integration", ["reused-shard", "actual-shard"]);
const read = (name) => readFileSync(path.join(report.reportDir, name), "utf8").trim();
console.log("usage=" + report.usage);
console.log("duration=" + read("duration_ms.txt"));
console.log("wall=" + read("wall_duration_ms.txt"));
console.log("start=" + read("start_time.txt"));
console.log("end=" + read("end_time.txt"));
EOF_NODE
)"
assert_contains "$mixed_aggregate_output" "usage=actual" "mixed aggregate usage"
assert_contains "$mixed_aggregate_output" "duration=2000" "mixed aggregate actual duration excludes reused shard"
assert_contains "$mixed_aggregate_output" "wall=2000" "mixed aggregate wall excludes reused shard"
assert_contains "$mixed_aggregate_output" "start=2026-01-01T00:00:20Z" "mixed aggregate start excludes reused shard"
assert_contains "$mixed_aggregate_output" "end=2026-01-01T00:00:22Z" "mixed aggregate end excludes reused shard"

shared_reuse_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-reuse.XXXXXX")"
cleanup_paths+=("$shared_reuse_results")
shared_reuse_output="$(
  CARTULARY_TEST_RESULTS_DIR="$shared_reuse_results/results" \
  CARTULARY_TEST_RUN_ID="shared-reuse" \
  NODE_BIN="$node_bin" \
  "$node_bin" --input-type=module - "$ROOT_DIR" "$incidents_incidents_shared_command" <<'EOF_NODE'
import path from "node:path";
import { mkdirSync, writeFileSync } from "node:fs";
import { assignExecutionFamily, createGoTargetContext, prepareSharedArtifactDir } from "./tools/harness/backend/backend-target-execution.mjs";

const [root, command] = process.argv.slice(2);
const ctx = createGoTargetContext({ repoRoot: root });
const sharedDir = prepareSharedArtifactDir(ctx, "module.incidents.support_integration");
mkdirSync(sharedDir, { recursive: true });
writeFileSync(path.join(sharedDir, "command.txt"), command + "\n");
writeFileSync(path.join(sharedDir, "complete"), "");
const result = await assignExecutionFamily(ctx, "backend-integration-support", "module.incidents.support_integration");
console.log("dir=" + result.reportDir);
console.log("usage=" + result.usage);
EOF_NODE
)"
assert_contains "$shared_reuse_output" "$shared_reuse_results/results/shared-reuse/_shared/module.incidents.support_integration" "shared reuse dir"
assert_contains "$shared_reuse_output" "usage=reused" "shared reuse usage"

shared_lock_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-lock.XXXXXX")"
cleanup_paths+=("$shared_lock_results")
shared_lock_output="$(
  CARTULARY_TEST_RESULTS_DIR="$shared_lock_results/results" \
  CARTULARY_TEST_RUN_ID="shared-lock" \
  NODE_BIN="$node_bin" \
  "$node_bin" --input-type=module - "$ROOT_DIR" <<'EOF_NODE'
import path from "node:path";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { acquireSharedReportLock, createGoTargetContext, prepareSharedArtifactDir, releaseSharedReportLock } from "./tools/harness/backend/backend-target-execution.mjs";

const root = process.argv[2];
const ctx = createGoTargetContext({ repoRoot: root });
const sharedDir = prepareSharedArtifactDir(ctx, "module.incidents.support_integration");
await acquireSharedReportLock(ctx, sharedDir, "module.incidents.support_integration");
console.log("shared=" + readFileSync(path.join(sharedDir, "capture.lock", "shared_report"), "utf8").trim());
console.log("pid=" + readFileSync(path.join(sharedDir, "capture.lock", "pid"), "utf8").trim());
releaseSharedReportLock(sharedDir);
console.log("released=" + (existsSync(path.join(sharedDir, "capture.lock")) ? "no" : "yes"));
mkdirSync(path.join(sharedDir, "capture.lock"), { recursive: true });
writeFileSync(path.join(sharedDir, "capture.lock", "pid"), "999999\n");
await acquireSharedReportLock(ctx, sharedDir, "module.incidents.support_integration");
console.log("stale_pid=" + readFileSync(path.join(sharedDir, "capture.lock", "pid"), "utf8").trim());
releaseSharedReportLock(sharedDir);
EOF_NODE
)"
assert_contains "$shared_lock_output" "shared=module.incidents.support_integration" "shared lock report name"
assert_contains "$shared_lock_output" "released=yes" "shared lock release"
assert_contains "$shared_lock_output" "stale_pid=" "shared lock stale owner replacement"

parallel_capture_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-parallel-capture.XXXXXX")"
cleanup_paths+=("$parallel_capture_results")
parallel_capture_output="$(
  CARTULARY_TEST_RESULTS_DIR="$parallel_capture_results/results" \
  CARTULARY_TEST_RUN_ID="parallel-capture" \
  NODE_BIN="$node_bin" \
  "$node_bin" --input-type=module - "$ROOT_DIR" "$parallel_capture_results" <<'EOF_NODE'
import path from "node:path";
import { mkdirSync, writeFileSync, readFileSync } from "node:fs";
import { captureNamedSharedReportsParallel, createGoTargetContext, inspectAggregateCommand, prepareSharedArtifactDir } from "./tools/harness/backend/backend-target-execution.mjs";

const [root, tmp] = process.argv.slice(2);
const ctx = createGoTargetContext({ repoRoot: root });
const shardNames = ["backend-integration-testutil-shard-01"];
for (const shard of shardNames) {
  const sharedDir = prepareSharedArtifactDir(ctx, shard);
  mkdirSync(sharedDir, { recursive: true });
  writeFileSync(path.join(sharedDir, "command.txt"), inspectAggregateCommand(ctx, "backend-integration", shard) + "\n");
  writeFileSync(path.join(sharedDir, "complete"), "");
}
const metadataDir = path.join(tmp, "metadata");
const status = await captureNamedSharedReportsParallel(ctx, "backend-integration", 2, metadataDir, shardNames);
const metadata = readFileSync(path.join(metadataDir, shardNames[0] + ".meta"), "utf8").trim().split(/\r?\n/u);
console.log("status=" + status);
console.log("usage=" + metadata[1]);
try {
  await captureNamedSharedReportsParallel(ctx, "backend-integration", 0, metadataDir, shardNames);
} catch (error) {
  console.log(error.message);
}
EOF_NODE
)"
assert_contains "$parallel_capture_output" "status=0" "parallel capture status"
assert_contains "$parallel_capture_output" "usage=reused" "parallel capture reused complete report"
assert_contains "$parallel_capture_output" "invalid shard job count" "parallel capture invalid jobs marker"

missing_metadata_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-missing-metadata.XXXXXX")"
cleanup_paths+=("$missing_metadata_results")
missing_metadata_dir="$missing_metadata_results/metadata"
mkdir -p "$missing_metadata_dir"
missing_metadata_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-store | head -n 1)"
set +e
missing_metadata_output="$(
  CARTULARY_OUTPUT_MODE=verbose \
  CARTULARY_TEST_RESULTS_DIR="$missing_metadata_results/results" \
  CARTULARY_TEST_RUN_ID="missing-metadata" \
  CARTULARY_TEST_TARGET="backend-store" \
  NODE_BIN="$node_bin" \
    "$node_bin" "$GO_TARGET_HELPER" finalize-shards backend-store "$missing_metadata_dir" "$missing_metadata_shard" \
    2>&1
)"
missing_metadata_status=$?
set -e
assert_equals "$missing_metadata_status" "1" "missing metadata finalizer status"
assert_contains "$missing_metadata_output" "missing shared report metadata for ${missing_metadata_shard}" "missing metadata diagnostic"
assert_contains "$missing_metadata_output" "failure_class=artifact" "missing metadata target failure class output"
missing_metadata_summary="$missing_metadata_results/results/missing-metadata/backend-store/target-summary.json"
assert_equals "$(json_field "$missing_metadata_summary" "failure_class")" "artifact" "missing metadata target failure class"
assert_equals "$(json_field "$missing_metadata_summary" "failure_reason")" "artifact_error" "missing metadata target failure reason"
assert_equals "$(json_field "$missing_metadata_summary" "failure_classes.artifact")" "1" "missing metadata artifact count"
assert_equals "$(json_field "$missing_metadata_summary" "own.timing_failures.length")" "0" "missing metadata does not infer timing failure"
assert_equals "$(json_field "$missing_metadata_summary" "failures.0.source")" "go-shard-finalizer" "missing metadata structured failure source"
assert_contains "$(json_field "$missing_metadata_summary" "failures.0.message")" "missing shared report metadata" "missing metadata structured failure message"

backend_unit_aggregates="$("$node_bin" "$ROOT_DIR/tools/harness/backend/target-plan.mjs" list-aggregates backend-unit)"
assert_contains "$backend_unit_aggregates" "app.server.unit" "backend-unit application aggregate"
assert_contains "$backend_unit_aggregates" "module.auth.unit" "backend-unit auth aggregate"
assert_contains "$backend_unit_aggregates" "backend-unit-configtest" "backend-unit configtest aggregate"

backend_store_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-store)"
assert_contains "$backend_store_shards" "module-" "backend-store captures semantic family shards"

backend_integration_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration)"
assert_contains "$backend_integration_shards" "module-entities-integration-shard-" "backend-integration captures semantic entity shards"
assert_contains "$backend_integration_shards" "$incidents_incidents_shard" "backend-integration captures planned incident shard"
assert_contains "$backend_integration_shards" "backend-integration-testutil-shard-01" "backend-integration captures raw testutil shard"
first_backend_integration_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration | head -n 1)"
assert_contains "$backend_integration_shards" "$first_backend_integration_shard" "backend-integration weighted shard order starts with heaviest shard"

backend_integration_support_shards="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration-support)"
assert_contains "$backend_integration_support_shards" "module-entities-support-integration-shard-" "backend-integration-support captures semantic entities shards"
assert_not_contains "$backend_integration_support_shards" "backend-integration-testutil" "backend-integration-support skips testutil shard"
first_backend_integration_support_shard="$("$node_bin" "$ROOT_DIR/tools/harness/backend/go-shard-plan.mjs" list-shards backend-integration-support | head -n 1)"
assert_contains "$backend_integration_support_shards" "$first_backend_integration_support_shard" "backend-integration-support weighted shard order starts with heaviest support shard"

support_zero_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-support-zero.XXXXXX")"
cleanup_paths+=("$support_zero_dir")
cat >"$support_zero_dir/support_zero_test.go" <<'EOF'
package rungotargetsupportzero

import "testing"

func TestUnrelatedSupportSmoke(t *testing.T) {}
EOF

support_zero_rel="./${support_zero_dir#"$ROOT_DIR"/}"
set +e
support_zero_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$GO_STEP_HELPER" "backend-integration support unknown-owner" '^(TestSupportUnknownOwnerIntegration_)' -- "$go_bin" test "$support_zero_rel" \
    2>&1
)"
support_zero_status=$?
set -e
if [[ "$support_zero_status" -eq 0 ]]; then
  fail "support zero-match: expected non-zero exit status"
fi
assert_contains "$support_zero_output" "failure: backend-integration support unknown-owner" "support zero-match label"
assert_contains "$support_zero_output" "coverage=support" "support zero-match coverage"
assert_contains "$support_zero_output" "message=support step matched zero tests" "support zero-match message"


printf "run-go-target smoke passed\n"
