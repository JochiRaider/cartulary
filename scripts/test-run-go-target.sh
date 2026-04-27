#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GO_PHASE_HELPER="$ROOT_DIR/scripts/lib/run-go-phase.sh"
GO_TARGET_HELPER="$ROOT_DIR/scripts/run-go-target.sh"
MANIFEST_HELPER="$ROOT_DIR/scripts/lib/phase-manifest.mjs"
PHASE_MAP_CHECK="$ROOT_DIR/scripts/check-phase-map.mjs"
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

assert_not_zero() {
  local actual="$1"
  local label="$2"

  if [[ -z "$actual" || "$actual" == "0" ]]; then
    fail "$label: expected a non-zero value, got [$actual]"
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

  "$node_bin" - "$ROOT_DIR" "$target" "$symbol" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root, target, symbol] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(process.execPath, [path.join(root, "scripts/print-go-shard-plan.mjs"), "--json", "--target", target], { encoding: "utf8", cwd: root }));
const shard = plan.shards.find((candidate) => candidate.items.some((item) => item.symbol === symbol));
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

func TestSupportPhase4Integration_VerboseSmoke(t *testing.T) {
	t.Log("support verbose line")
}
EOF

verbose_go_rel="./${verbose_go_dir#"$ROOT_DIR"/}"
verbose_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$GO_PHASE_HELPER" "run-go-target verbose smoke" '^(TestSupportPhase4Integration_VerboseSmoke)$' -- "$go_bin" test "$verbose_go_rel" \
    2>&1
)"
assert_contains "$verbose_output" "== run-go-target verbose smoke ==" "verbose go banner"
assert_contains "$verbose_output" "support verbose line" "verbose go human output"
assert_not_contains "$verbose_output" "\"Action\":\"output\"" "verbose go raw json output"
assert_not_contains "$verbose_output" "\"Test\":\"TestSupportPhase4Integration_VerboseSmoke\"" "verbose go raw json test field"

duration_results_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-durations.XXXXXX")"
cleanup_paths+=("$duration_results_dir")
shared_report_dir="$duration_results_dir/shared-report"
mkdir -p "$shared_report_dir"
cat >"$shared_report_dir/runner.jsonl" <<'EOF'
{"Time":"2000-01-01T00:00:00Z","Action":"run","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke"}
{"Time":"2000-01-01T00:00:00Z","Action":"output","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke","Output":"=== RUN   TestSupportPhase4Integration_Smoke\n"}
{"Time":"2000-01-01T00:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke","Elapsed":0.001}
{"Time":"2000-01-01T00:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Elapsed":0.001}
EOF
touch "$shared_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities" >"$shared_report_dir/command.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$shared_report_dir/start_time.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$shared_report_dir/end_time.txt"
printf '%s\n' "1200" >"$shared_report_dir/duration_ms.txt"
printf '%s\n' "0" >"$shared_report_dir/exit_status.txt"

duration_artifacts_root="$duration_results_dir/results"
(
  export CARTULARY_TEST_RESULTS_DIR="$duration_artifacts_root"
  export CARTULARY_TEST_RUN_ID="duration-smoke"
  export CARTULARY_TEST_TARGET="backend-unit-smoke"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  emit_go_raw_phase "duration actual" actual "$shared_report_dir" '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities
  emit_go_raw_phase "duration reused" reused "$shared_report_dir" '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities
  emit_go_raw_phase "duration derived" derived "$shared_report_dir" '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary backend-unit-smoke pass >/dev/null
  "$ROOT_DIR/scripts/lib/test-output.sh" run-summary "duration smoke" pass 1 1 - backend-unit-smoke >/dev/null
)

duration_actual_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-actual/phase-summary.json"
duration_reused_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-reused/phase-summary.json"
duration_derived_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/duration-derived/phase-summary.json"
duration_target_summary="$duration_artifacts_root/duration-smoke/backend-unit-smoke/target-summary.json"
duration_target_timing="$duration_artifacts_root/duration-smoke/backend-unit-smoke/target-timing.json"
duration_run_summary="$duration_artifacts_root/duration-smoke/run-summary.json"

assert_json_field_absent "$duration_actual_summary" "duration_ms" "duration actual legacy duration"
assert_not_negative "$(json_field "$duration_actual_summary" "wall_duration_ms")" "duration actual phase wall duration"
assert_not_negative "$(json_field "$duration_actual_summary" "critical_path_wall_duration_ms")" "duration actual phase critical path duration"
assert_equals "$(json_field "$duration_actual_summary" "accounting_mode")" "actual" "duration actual accounting mode"
assert_equals "$(json_field "$duration_actual_summary" "executed_duration_ms")" "1200" "duration actual executed duration"
assert_equals "$(json_field "$duration_actual_summary" "logical_duration_ms")" "1200" "duration actual logical duration"
assert_equals "$(json_field "$duration_actual_summary" "reused_duration_ms")" "0" "duration actual reused duration"
assert_equals "$(json_field "$duration_actual_summary" "derived_duration_ms")" "0" "duration actual derived duration"
assert_equals "$(json_field "$duration_actual_summary" "wall_duration_ms")" "1200" "duration actual wall duration"
assert_json_field_absent "$duration_reused_summary" "duration_ms" "duration reused legacy duration"
assert_not_negative "$(json_field "$duration_reused_summary" "wall_duration_ms")" "duration reused phase wall duration"
assert_not_negative "$(json_field "$duration_reused_summary" "critical_path_wall_duration_ms")" "duration reused phase critical path duration"
assert_equals "$(json_field "$duration_reused_summary" "accounting_mode")" "reused" "duration reused accounting mode"
assert_equals "$(json_field "$duration_reused_summary" "executed_duration_ms")" "0" "duration reused executed duration"
assert_equals "$(json_field "$duration_reused_summary" "logical_duration_ms")" "1200" "duration reused logical duration"
assert_equals "$(json_field "$duration_reused_summary" "reused_duration_ms")" "1200" "duration reused reused duration"
assert_equals "$(json_field "$duration_reused_summary" "derived_duration_ms")" "0" "duration reused derived duration"
assert_equals "$(json_field "$duration_reused_summary" "wall_duration_ms")" "0" "duration reused wall duration"
assert_json_field_absent "$duration_derived_summary" "duration_ms" "duration derived legacy duration"
assert_not_negative "$(json_field "$duration_derived_summary" "wall_duration_ms")" "duration derived phase wall duration"
assert_not_negative "$(json_field "$duration_derived_summary" "critical_path_wall_duration_ms")" "duration derived phase critical path duration"
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

reused_window_results_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-reused-window.XXXXXX")"
cleanup_paths+=("$reused_window_results_dir")
reused_window_report_dir="$reused_window_results_dir/shared-report"
mkdir -p "$reused_window_report_dir"
cp "$shared_report_dir/runner.jsonl" "$reused_window_report_dir/runner.jsonl"
touch "$reused_window_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities" >"$reused_window_report_dir/command.txt"
printf '%s\n' "2000-01-01T00:00:00Z" >"$reused_window_report_dir/start_time.txt"
printf '%s\n' "2000-01-01T00:00:10Z" >"$reused_window_report_dir/end_time.txt"
printf '%s\n' "10000" >"$reused_window_report_dir/duration_ms.txt"
printf '%s\n' "0" >"$reused_window_report_dir/exit_status.txt"
(
  export CARTULARY_TEST_RESULTS_DIR="$reused_window_results_dir/results"
  export CARTULARY_TEST_RUN_ID="reused-window"
  export CARTULARY_TEST_TARGET="backend-integration-support"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  emit_go_raw_phase "reused window support" reused "$reused_window_report_dir" '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities
  emit_target_timing_span \
    test_command \
    "run-go-target backend-integration-support" \
    "2026-01-01T00:00:00Z" \
    "2026-01-01T00:00:00.400Z" \
    400 \
    pass \
    0
  emit_target_summary pass >/dev/null
)
reused_window_summary="$reused_window_results_dir/results/reused-window/backend-integration-support/target-summary.json"
assert_equals "$(json_field "$reused_window_summary" "accounting_modes.actual")" "0" "reused window actual accounting count"
assert_equals "$(json_field "$reused_window_summary" "accounting_modes.reused")" "1" "reused window reused accounting count"
assert_equals "$(json_field "$reused_window_summary" "wall_duration_ms")" "400" "reused window target wall follows invocation span"
assert_equals "$(json_field "$reused_window_summary" "start_time")" "2026-01-01T00:00:00Z" "reused window target start follows invocation span"
assert_equals "$(json_field "$reused_window_summary" "end_time")" "2026-01-01T00:00:00.400Z" "reused window target end follows invocation span"

backend_unit_core_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-unit backend-unit-core
)"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase0_" "backend-unit-core phase0 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase2Unit_" "backend-unit-core phase2 selector"
assert_contains "$backend_unit_core_shared_command" "TestSupportPhase3Unit_" "backend-unit-core phase3 selector"

backend_unit_auth_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-unit backend-unit-auth
)"
assert_contains "$backend_unit_auth_shared_command" "TestSupportPhase1_" "backend-unit-auth phase1 selector"

phase0_platform_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support backend-integration-platform
)"
assert_contains "$phase0_platform_shared_command" "TestSupportPhase0_" "backend-integration phase0 platform support selector"
phase0_platform_authoritative_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-platform
)"
assert_contains "$phase0_platform_authoritative_command" "TestPhase0_SchemaBootstrap" "backend-integration phase0 platform authoritative selector"
assert_not_contains "$phase0_platform_authoritative_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 platform excludes app selector"

phase0_app_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-app
)"
assert_contains "$phase0_app_shared_command" "TestPhase0_FirstAdminBootstrap" "backend-integration phase0 app selector"
assert_not_contains "$phase0_app_shared_command" "TestSupportPhase0_" "backend-integration phase0 app excludes platform support selector"

phase2_incidents_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support backend-integration-incidents
)"
assert_contains "$phase2_incidents_shared_command" "TestSupportPhase2_" "backend-integration phase2 incidents support selector"
phase2_incidents_authoritative_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-incidents
)"
assert_contains "$phase2_incidents_authoritative_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents authoritative selector"

phase2_incidents_shard="$(find_planned_shard_for_symbol backend-integration TestPhase2_I_2_01_IncidentCreatePersistsBootstrapStateAndRollsBackAtomically)"
phase2_incidents_shard_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration "$phase2_incidents_shard"
)"
assert_contains "$phase2_incidents_shard_command" "TestPhase2_I_2_01" "backend-integration phase2 incidents planned shard selector"

phase2_incidents_support_shard="$(find_planned_shard_for_symbol backend-integration-support TestSupportPhase2_ControlBoundaryIncidentCoreDeploymentAdminWithoutMembershipDenied)"
phase2_incidents_support_shard_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support "$phase2_incidents_support_shard"
)"
assert_contains "$phase2_incidents_support_shard_command" "TestSupportPhase2_" "backend-integration support phase2 planned shard selector"

"$node_bin" - "$ROOT_DIR" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(process.execPath, [path.join(root, "scripts/lib/go-shard-plan.mjs"), "json"], { encoding: "utf8", cwd: root }));
const mixed = plan.shards.filter((shard) => shard.has_authoritative && shard.has_support);
const shared = plan.shards.filter((shard) => shard.shared_across_targets);
if (mixed.length > 0 || shared.length > 0) {
  throw new Error(`backend integration shards must keep authoritative/support ownership separate; mixed=${mixed.map((shard) => shard.name).join(",")} shared=${shared.map((shard) => shard.name).join(",")}`);
}
EOF

phase3_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support backend-integration-timeline
)"
assert_contains "$phase3_timeline_shared_command" "TestSupportPhase3Integration_" "backend-integration phase3 timeline support selector"
phase3_timeline_authoritative_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-timeline
)"
assert_contains "$phase3_timeline_authoritative_command" "TestPhase3_I_3_01" "backend-integration phase3 timeline authoritative selector"

phase4_entities_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support backend-integration-entities
)"
assert_contains "$phase4_entities_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 entities support selector"
phase4_entities_authoritative_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-entities
)"
assert_contains "$phase4_entities_authoritative_command" "TestPhase4_ResolveRoute" "backend-integration phase4 entities authoritative selector"
assert_not_contains "$phase4_entities_authoritative_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 entities excludes timeline selector"

phase4_timeline_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration backend-integration-timeline
)"
assert_contains "$phase4_timeline_shared_command" "TestPhase4_AutoResolutionEligibility" "backend-integration phase4 timeline selector"
assert_not_contains "$phase4_timeline_shared_command" "TestSupportPhase4Integration_" "backend-integration phase4 timeline excludes entities support selector"

auth_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-aggregate-command backend-integration-support backend-integration-auth
)"
assert_contains "$auth_shared_command" "TestSupportPhase1_" "backend-integration-auth phase1 selector"

shared_mismatch_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-mismatch.XXXXXX")"
cleanup_paths+=("$shared_mismatch_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$shared_mismatch_results/results"
  export CARTULARY_TEST_RUN_ID="shared-mismatch"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  shared_dir="$(prepare_shared_artifact_dir backend-integration-app)"
  mkdir -p "$shared_dir"
  printf '%s\n' "env go test -json -run '^TestOld$' ./internal/app" >"$shared_dir/command.txt"
  touch "$shared_dir/complete"

  set +e
  mismatch_output="$(
    capture_go_report backend-integration-app '^TestCurrent$' -- ./internal/app \
      2>&1
  )"
  mismatch_status=$?
  set -e

  if [[ "$mismatch_status" -eq 0 ]]; then
    fail "shared command mismatch: expected capture_go_report to fail"
  fi
  assert_contains "$mismatch_output" "shared_go_report_command_mismatch" "shared command mismatch marker"
)

mixed_aggregate_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-mixed-aggregate.XXXXXX")"
cleanup_paths+=("$mixed_aggregate_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$mixed_aggregate_results/results"
  export CARTULARY_TEST_RUN_ID="mixed-aggregate"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  metadata_dir="$mixed_aggregate_results/metadata"
  mkdir -p "$metadata_dir/actual-shard" "$metadata_dir/reused-shard"
  for shard in actual-shard reused-shard; do
    printf '%s\n' "env go test -json -run '^Test$' ./internal/modules/entities" >"$metadata_dir/$shard/command.txt"
    touch "$metadata_dir/$shard/runner.jsonl" "$metadata_dir/$shard/stderr.log"
    printf '%s\n' "0" >"$metadata_dir/$shard/exit_status.txt"
  done
  printf '%s\n' "2026-01-01T00:00:00Z" >"$metadata_dir/reused-shard/start_time.txt"
  printf '%s\n' "2026-01-01T00:00:10Z" >"$metadata_dir/reused-shard/end_time.txt"
  printf '%s\n' "10000" >"$metadata_dir/reused-shard/duration_ms.txt"
  printf '%s\n' "2026-01-01T00:00:20Z" >"$metadata_dir/actual-shard/start_time.txt"
  printf '%s\n' "2026-01-01T00:00:22Z" >"$metadata_dir/actual-shard/end_time.txt"
  printf '%s\n' "2000" >"$metadata_dir/actual-shard/duration_ms.txt"
  printf '%s\n%s\n' "$metadata_dir/reused-shard" reused >"$metadata_dir/reused-shard.meta"
  printf '%s\n%s\n' "$metadata_dir/actual-shard" actual >"$metadata_dir/actual-shard.meta"

  create_aggregate_report aggregate_dir aggregate_usage "$metadata_dir" mixed-aggregate backend-integration reused-shard actual-shard
  assert_equals "$aggregate_usage" "actual" "mixed aggregate usage"
  assert_equals "$(<"$aggregate_dir/duration_ms.txt")" "2000" "mixed aggregate actual duration excludes reused shard"
  assert_equals "$(<"$aggregate_dir/wall_duration_ms.txt")" "2000" "mixed aggregate wall excludes reused shard"
  assert_equals "$(<"$aggregate_dir/start_time.txt")" "2026-01-01T00:00:20Z" "mixed aggregate start excludes reused shard"
  assert_equals "$(<"$aggregate_dir/end_time.txt")" "2026-01-01T00:00:22Z" "mixed aggregate end excludes reused shard"
)

shared_reuse_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-reuse.XXXXXX")"
cleanup_paths+=("$shared_reuse_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$shared_reuse_results/results"
  export CARTULARY_TEST_RUN_ID="shared-reuse"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  shared_dir="$(prepare_shared_artifact_dir backend-integration-incidents)"
  mkdir -p "$shared_dir"
  printf '%s\n' "$phase2_incidents_shared_command" >"$shared_dir/command.txt"
  touch "$shared_dir/complete"

  assign_execution_family reused_dir reused_usage backend-integration-support backend-integration-incidents
  if [[ "$reused_dir" != "$shared_dir" ]]; then
    fail "shared reuse: expected assign_execution_family to reuse the existing shared dir"
  fi
  if [[ "$reused_usage" != "reused" ]]; then
    fail "shared reuse: expected assign_execution_family to mark the report as reused"
  fi
)

shared_lock_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-lock.XXXXXX")"
cleanup_paths+=("$shared_lock_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$shared_lock_results/results"
  export CARTULARY_TEST_RUN_ID="shared-lock"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  shared_dir="$(prepare_shared_artifact_dir backend-integration-incidents)"
  acquire_shared_report_lock "$shared_dir" backend-integration-incidents
  assert_equals "$(<"$shared_dir/capture.lock/shared_report")" "backend-integration-incidents" "shared lock report name"
  assert_equals "$(<"$shared_dir/capture.lock/pid")" "$$" "shared lock owner pid"

  set +e
  lock_timeout_output="$(
    CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS=1 \
      acquire_shared_report_lock "$shared_dir" backend-integration-incidents \
      2>&1
  )"
  lock_timeout_status=$?
  set -e
  if [[ "$lock_timeout_status" -eq 0 ]]; then
    fail "shared lock timeout: expected second lock acquisition to fail"
  fi
  assert_contains "$lock_timeout_output" "shared_go_report_lock_timeout" "shared lock timeout marker"

  release_shared_report_lock "$shared_dir"
  if [[ -e "$shared_dir/capture.lock" ]]; then
    fail "shared lock release: capture.lock must be removed"
  fi

  mkdir -p "$shared_dir/capture.lock"
  printf '%s\n' "999999" >"$shared_dir/capture.lock/pid"
  acquire_shared_report_lock "$shared_dir" backend-integration-incidents
  assert_equals "$(<"$shared_dir/capture.lock/pid")" "$$" "shared lock stale owner replacement"
  release_shared_report_lock "$shared_dir"
)

parallel_capture_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-parallel-capture.XXXXXX")"
cleanup_paths+=("$parallel_capture_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$parallel_capture_results/results"
  export CARTULARY_TEST_RUN_ID="parallel-capture"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  assign_execution_family() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local shared_name="$4"
    dir_ref="/tmp/${shared_name}"
    usage_ref="actual"
  }

  metadata_dir="$parallel_capture_results/metadata"
  capture_named_shared_reports_parallel backend-integration 2 "$metadata_dir" shard-a shard-b shard-c
  read_shared_report_metadata shard_a_dir shard_a_usage "$metadata_dir" shard-a
  read_shared_report_metadata shard_c_dir shard_c_usage "$metadata_dir" shard-c
  assert_equals "$shard_a_dir" "/tmp/shard-a" "parallel capture shard-a dir"
  assert_equals "$shard_a_usage" "actual" "parallel capture shard-a usage"
  assert_equals "$shard_c_dir" "/tmp/shard-c" "parallel capture shard-c dir"
  assert_equals "$shard_c_usage" "actual" "parallel capture shard-c usage"

  set +e
  invalid_jobs_output="$(
    capture_named_shared_reports_parallel backend-integration 0 "$metadata_dir" shard-a \
      2>&1
  )"
  invalid_jobs_status=$?
  set -e
  if [[ "$invalid_jobs_status" -eq 0 ]]; then
    fail "parallel capture invalid jobs: expected failure"
  fi
  assert_contains "$invalid_jobs_output" "invalid shard job count" "parallel capture invalid jobs marker"
)

parallel_scheduler_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-parallel-scheduler.XXXXXX")"
cleanup_paths+=("$parallel_scheduler_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$parallel_scheduler_results/results"
  export CARTULARY_TEST_RUN_ID="parallel-scheduler"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  assign_execution_family() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local shared_name="$4"
    printf '%s\n' "$(phase_now_monotonic_ms)" >"$parallel_scheduler_results/${shared_name}.start"
    case "${shared_name}" in
      shard-a) sleep 0.4 ;;
      shard-b) sleep 0.05 ;;
      *) sleep 0.01 ;;
    esac
    printf '%s\n' "$(phase_now_monotonic_ms)" >"$parallel_scheduler_results/${shared_name}.end"
    dir_ref="/tmp/${shared_name}"
    usage_ref="actual"
  }

  metadata_dir="$parallel_scheduler_results/metadata"
  capture_named_shared_reports_parallel backend-integration 2 "$metadata_dir" shard-a shard-b shard-c
  shard_a_end="$(<"$parallel_scheduler_results/shard-a.end")"
  shard_c_start="$(<"$parallel_scheduler_results/shard-c.start")"
  assert_less_than "$shard_c_start" "$shard_a_end" "parallel capture starts next shard after first completed child"
)

backend_unit_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  assign_calls=0
  emit_calls=0
  assigned_names=""
  emitted_names=""
  finish_status=""
  assign_execution_family() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    local family="$4"
    assign_calls=$((assign_calls + 1))
    assigned_names="${assigned_names} ${family}"
    dir_ref="/tmp/${family}"
    usage_ref="actual"
  }
  emit_execution_family() {
    emit_calls=$((emit_calls + 1))
    emitted_names="${emitted_names} $2"
  }
  finish_target() {
    finish_status="$1"
    printf "assign_calls=%s emit_calls=%s finish_status=%s assigned=%s emitted=%s\n" "$assign_calls" "$emit_calls" "$finish_status" "$assigned_names" "$emitted_names"
  }
  run_unsharded_target backend-unit
)"
assert_contains "$backend_unit_structure" "assign_calls=3" "backend-unit aggregate capture count"
assert_contains "$backend_unit_structure" "emit_calls=3" "backend-unit aggregate emission count"
assert_contains "$backend_unit_structure" "backend-unit-core" "backend-unit core aggregate"
assert_contains "$backend_unit_structure" "backend-unit-auth" "backend-unit auth aggregate"
assert_contains "$backend_unit_structure" "backend-unit-configtest" "backend-unit configtest aggregate"
assert_contains "$backend_unit_structure" "finish_status=0" "backend-unit finish status"

backend_store_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    shift 3
    capture_calls=$((capture_calls + 1))
    capture_names="$*"
  }
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-store
)"
assert_contains "$backend_store_structure" "capture_calls=1" "backend-store shard capture count"
assert_contains "$backend_store_structure" "capture_target=backend-store" "backend-store shard capture target"
assert_contains "$backend_store_structure" "capture_jobs=4" "backend-store default shard jobs"
assert_contains "$backend_store_structure" "finalize_calls=1" "backend-store finalize count"
assert_contains "$backend_store_structure" "backend-store-shard-" "backend-store captures planned shards"
assert_contains "$backend_store_structure" "finish_status=0" "backend-store finish status"

backend_integration_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    local metadata_dir="$3"
    shift 3
    capture_calls=$((capture_calls + 1))
    mkdir -p "$metadata_dir"
    local shared_name
    for shared_name in "$@"; do
      capture_names="${capture_names} ${shared_name}"
      printf '/tmp/%s\nactual\n' "$shared_name" >"$metadata_dir/${shared_name}.meta"
    done
  }
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-integration
)"
assert_contains "$backend_integration_structure" "capture_calls=1" "backend-integration parallel capture count"
assert_contains "$backend_integration_structure" "capture_target=backend-integration" "backend-integration parallel capture target"
assert_contains "$backend_integration_structure" "capture_jobs=4" "backend-integration default shard jobs"
assert_contains "$backend_integration_structure" "finalize_calls=1" "backend-integration finalize count"
assert_contains "$backend_integration_structure" "backend-integration-entities-shard-" "backend-integration captures entity shards"
assert_contains "$backend_integration_structure" "$phase2_incidents_shard" "backend-integration captures planned phase2 incident shard"
assert_contains "$backend_integration_structure" "backend-integration-testutil-shard-01" "backend-integration captures raw testutil shard"
first_backend_integration_shard="$("$node_bin" "$ROOT_DIR/scripts/lib/go-shard-plan.mjs" list-shards backend-integration | head -n 1)"
assert_contains "$backend_integration_structure" "names= $first_backend_integration_shard" "backend-integration weighted shard order starts with heaviest shard"
assert_contains "$backend_integration_structure" "finish_status=0" "backend-integration finish status"

backend_integration_support_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  capture_calls=0
  capture_jobs=""
  capture_target=""
  capture_names=""
  finalize_calls=0
  finalize_target=""
  finish_status=""
  capture_named_shared_reports_parallel() {
    capture_target="$1"
    capture_jobs="$2"
    local metadata_dir="$3"
    shift 3
    capture_calls=$((capture_calls + 1))
    mkdir -p "$metadata_dir"
    local shared_name
    for shared_name in "$@"; do
      capture_names="${capture_names} ${shared_name}"
      printf '/tmp/%s\nactual\n' "$shared_name" >"$metadata_dir/${shared_name}.meta"
    done
  }
  finalize_scheduled_shards() {
    finalize_calls=$((finalize_calls + 1))
    finalize_target="$1"
    finish_target 0
  }
  finish_target() {
    finish_status="$1"
    printf "capture_calls=%s capture_target=%s capture_jobs=%s finalize_calls=%s finalize_target=%s finish_status=%s names=%s\n" "$capture_calls" "$capture_target" "$capture_jobs" "$finalize_calls" "$finalize_target" "$finish_status" "$capture_names"
  }
  run_sharded_target backend-integration-support
)"
assert_contains "$backend_integration_support_structure" "capture_calls=1" "backend-integration-support parallel capture count"
assert_contains "$backend_integration_support_structure" "capture_target=backend-integration-support" "backend-integration-support parallel capture target"
assert_contains "$backend_integration_support_structure" "capture_jobs=4" "backend-integration-support default shard jobs"
assert_contains "$backend_integration_support_structure" "finalize_calls=1" "backend-integration-support finalize count"
assert_contains "$backend_integration_support_structure" "backend-integration-entities-shard-" "backend-integration-support captures entities shards"
assert_not_contains "$backend_integration_support_structure" "backend-integration-testutil" "backend-integration-support skips testutil shard"
first_backend_integration_support_shard="$("$node_bin" "$ROOT_DIR/scripts/lib/go-shard-plan.mjs" list-shards backend-integration-support | head -n 1)"
assert_contains "$backend_integration_support_structure" "names= $first_backend_integration_support_shard" "backend-integration-support weighted shard order starts with heaviest support shard"
assert_contains "$backend_integration_support_structure" "finish_status=0" "backend-integration-support finish status"

phase0_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase0 backend_unit ./internal/platform/... ./internal/app)"
assert_not_zero "$phase0_backend_unit_support_count" "phase0 backend-unit support-go-count"
phase1_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase1 backend_unit ./internal/modules/auth)"
assert_not_zero "$phase1_backend_unit_support_count" "phase1 backend-unit support-go-count"
phase2_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase2 backend_unit ./internal/modules/incidents)"
assert_not_zero "$phase2_backend_unit_support_count" "phase2 backend-unit support-go-count"
phase3_backend_unit_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase3 backend_unit ./internal/modules/timeline)"
assert_not_zero "$phase3_backend_unit_support_count" "phase3 backend-unit support-go-count"

phase1_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase1 backend_integration_support ./internal/modules/auth)"
assert_not_zero "$phase1_support_count" "phase1 support-go-count"
phase2_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase2 backend_integration_support ./internal/modules/incidents)"
assert_not_zero "$phase2_support_count" "phase2 support-go-count"
phase3_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase3 backend_integration_support ./internal/modules/timeline)"
assert_not_zero "$phase3_support_count" "phase3 support-go-count"
phase4_support_count="$("$node_bin" "$MANIFEST_HELPER" support-go-count phase4 backend_integration_support ./internal/modules/entities ./internal/modules/timeline)"
assert_not_zero "$phase4_support_count" "phase4 support-go-count"

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
    "$GO_PHASE_HELPER" "backend-integration support phase99" '^(TestSupportPhase99Integration_)' -- "$go_bin" test "$support_zero_rel" \
    2>&1
)"
support_zero_status=$?
set -e
if [[ "$support_zero_status" -eq 0 ]]; then
  fail "support zero-match: expected non-zero exit status"
fi
assert_contains "$support_zero_output" "failure: backend-integration support phase99" "support zero-match label"
assert_contains "$support_zero_output" "coverage=support" "support zero-match coverage"
assert_contains "$support_zero_output" "message=support phase matched zero tests" "support zero-match message"

manifest_smoke_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-support-manifest.XXXXXX")"
manifest_smoke_root="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-support-manifests.XXXXXX")"
manifest_smoke_tools="$manifest_smoke_root/tools"
mkdir -p "$manifest_smoke_tools"
cp "$ROOT_DIR"/tools/phase*_test_map.json "$manifest_smoke_tools"/
cleanup_paths+=(
  "$manifest_smoke_dir"
  "$manifest_smoke_root"
)
cat >"$manifest_smoke_dir/support_manifest_smoke_test.go" <<'EOF'
package rungotargetsupportmanifest

import "testing"

func TestPhase20_SupportManifest_U_20_01(t *testing.T) {}
func TestPhase21_SupportManifest_U_21_01(t *testing.T) {}
func TestPhase22_SupportManifest_U_22_01(t *testing.T) {}
func TestPhase23_SupportManifest_U_23_01(t *testing.T) {}
func TestPhase24_SupportManifest_U_24_01(t *testing.T) {}

func TestSupportPhase20Unit_Registered(t *testing.T) {}
func TestSupportPhase22Unit_Registered(t *testing.T) {}
func TestSupportPhase23Unit_Registered(t *testing.T) {}
func TestSupportPhase24Unit_Registered(t *testing.T) {}
EOF

manifest_smoke_rel="./${manifest_smoke_dir#"$ROOT_DIR"/}"
manifest_smoke_file="${manifest_smoke_rel#./}/support_manifest_smoke_test.go"

cat >"$manifest_smoke_tools/phase20_test_map.json" <<EOF
{
  "expected_ids": ["U-20-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase20Unit_Registered",
      "selection_pattern": "TestSupportPhase20Unit_",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core"
    }
  ],
  "unit": [
    {
      "id": "U-20-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestPhase20_SupportManifest_U_20_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "smoke"
    }
  ]
}
EOF
CARTULARY_PHASE_MANIFEST_ROOT="$manifest_smoke_root" NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase20 >/dev/null

cat >"$manifest_smoke_tools/phase21_test_map.json" <<EOF
{
  "expected_ids": ["U-21-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase21Unit_Missing",
      "selection_pattern": "TestSupportPhase21Unit_",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core"
    }
  ],
  "unit": [
    {
      "id": "U-21-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestPhase21_SupportManifest_U_21_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase21_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$manifest_smoke_root" NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase21 \
    2>&1
)"
phase21_status=$?
set -e
if [[ "$phase21_status" -eq 0 ]]; then
  fail "phase21 support manifest: expected validation failure"
fi
assert_contains "$phase21_output" "not found in" "phase21 missing support symbol"

cat >"$manifest_smoke_tools/phase22_test_map.json" <<EOF
{
  "expected_ids": ["U-22-01"],
  "support_go_targets": [
    {
      "target": "backend_process_support",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase22Unit_Registered",
      "selection_pattern": "TestSupportPhase22Unit_",
      "execution_family": "backend-process",
      "execution_label": "Backend process"
    }
  ],
  "unit": [
    {
      "id": "U-22-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestPhase22_SupportManifest_U_22_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase22_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$manifest_smoke_root" NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase22 \
    2>&1
)"
phase22_status=$?
set -e
if [[ "$phase22_status" -eq 0 ]]; then
  fail "phase22 support manifest: expected validation failure"
fi
assert_contains "$phase22_output" "must declare target=backend_unit|backend_integration_support" "phase22 invalid support target"

cat >"$manifest_smoke_tools/phase23_test_map.json" <<EOF
{
  "expected_ids": ["U-23-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase23Unit_Registered",
      "selection_pattern": "TestSupportPhase23Integration_",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core"
    }
  ],
  "unit": [
    {
      "id": "U-23-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestPhase23_SupportManifest_U_23_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase23_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$manifest_smoke_root" NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase23 \
    2>&1
)"
phase23_status=$?
set -e
if [[ "$phase23_status" -eq 0 ]]; then
  fail "phase23 support manifest: expected validation failure"
fi
assert_contains "$phase23_output" "selection_pattern does not match symbol" "phase23 selection pattern mismatch"

cat >"$manifest_smoke_tools/phase24_test_map.json" <<EOF
{
  "expected_ids": ["U-24-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "./internal/modules/auth",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase24Unit_Registered",
      "selection_pattern": "TestSupportPhase24Unit_",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core"
    }
  ],
  "unit": [
    {
      "id": "U-24-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestPhase24_SupportManifest_U_24_01",
      "execution_dependency": "backend_unit",
      "execution_family": "backend-unit-core",
      "execution_label": "Backend unit core",
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase24_output="$(
  CARTULARY_PHASE_MANIFEST_ROOT="$manifest_smoke_root" NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase24 \
    2>&1
)"
phase24_status=$?
set -e
if [[ "$phase24_status" -eq 0 ]]; then
  fail "phase24 support manifest: expected validation failure"
fi
assert_contains "$phase24_output" "does not belong to package" "phase24 package mismatch"

for synthetic_manifest in \
  "$ROOT_DIR/tools/phase20_test_map.json" \
  "$ROOT_DIR/tools/phase21_test_map.json" \
  "$ROOT_DIR/tools/phase22_test_map.json" \
  "$ROOT_DIR/tools/phase23_test_map.json" \
  "$ROOT_DIR/tools/phase24_test_map.json"
do
  if [[ -e "$synthetic_manifest" ]]; then
    fail "run-go-target smoke must not write synthetic manifests into repo tools/: $synthetic_manifest"
  fi
done
