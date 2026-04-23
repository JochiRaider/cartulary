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

node_bin="${NODE_BIN:-node}"
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
{"Time":"2026-04-22T12:00:00Z","Action":"run","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke"}
{"Time":"2026-04-22T12:00:00Z","Action":"output","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke","Output":"=== RUN   TestSupportPhase4Integration_Smoke\n"}
{"Time":"2026-04-22T12:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Test":"TestSupportPhase4Integration_Smoke","Elapsed":0.001}
{"Time":"2026-04-22T12:00:00Z","Action":"pass","Package":"github.com/JochiRaider/cartulary/internal/modules/entities","Elapsed":0.001}
EOF
touch "$shared_report_dir/stderr.log"
printf '%s\n' "env go test -json -run '^(TestSupportPhase4Integration_Smoke)$' ./internal/modules/entities" >"$shared_report_dir/command.txt"
printf '%s\n' "2026-04-22T12:00:00Z" >"$shared_report_dir/start_time.txt"
printf '%s\n' "2026-04-22T12:00:00Z" >"$shared_report_dir/end_time.txt"
printf '%s\n' "-17" >"$shared_report_dir/duration_ms.txt"
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
duration_run_summary="$duration_artifacts_root/duration-smoke/run-summary.json"

assert_not_negative "$(json_field "$duration_actual_summary" "duration_ms")" "duration actual phase duration"
assert_not_negative "$(json_field "$duration_actual_summary" "wall_duration_ms")" "duration actual phase wall duration"
assert_not_negative "$(json_field "$duration_reused_summary" "duration_ms")" "duration reused phase duration"
assert_not_negative "$(json_field "$duration_reused_summary" "wall_duration_ms")" "duration reused phase wall duration"
assert_not_negative "$(json_field "$duration_derived_summary" "duration_ms")" "duration derived phase duration"
assert_not_negative "$(json_field "$duration_derived_summary" "wall_duration_ms")" "duration derived phase wall duration"
assert_not_negative "$(json_field "$duration_target_summary" "duration_ms")" "duration target duration"
assert_not_negative "$(json_field "$duration_target_summary" "wall_duration_ms")" "duration target wall duration"
assert_not_negative "$(json_field "$duration_run_summary" "duration_ms")" "duration run duration"
assert_not_negative "$(json_field "$duration_run_summary" "wall_duration_ms")" "duration run wall duration"

core_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-core
)"
assert_contains "$core_shared_command" "TestSupportPhase2_" "backend-integration-core phase2 selector"
assert_contains "$core_shared_command" "TestSupportPhase3Integration_" "backend-integration-core phase3 selector"
assert_contains "$core_shared_command" "TestSupportPhase4Integration_" "backend-integration-core phase4 selector"

auth_shared_command="$(
  NODE_BIN="$node_bin" "$GO_TARGET_HELPER" inspect-shared-command backend-integration-support backend-integration-auth
)"
assert_contains "$auth_shared_command" "TestSupportPhase1_" "backend-integration-auth phase1 selector"

shared_mismatch_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-mismatch.XXXXXX")"
cleanup_paths+=("$shared_mismatch_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$shared_mismatch_results/results"
  export CARTULARY_TEST_RUN_ID="shared-mismatch"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  shared_dir="$(prepare_shared_artifact_dir backend-integration-core)"
  mkdir -p "$shared_dir"
  printf '%s\n' "env go test -json -run '^TestOld$' ./internal/app" >"$shared_dir/command.txt"
  touch "$shared_dir/complete"

  set +e
  mismatch_output="$(
    capture_go_report backend-integration-core '^TestCurrent$' -- ./internal/app \
      2>&1
  )"
  mismatch_status=$?
  set -e

  if [[ "$mismatch_status" -eq 0 ]]; then
    fail "shared command mismatch: expected capture_go_report to fail"
  fi
  assert_contains "$mismatch_output" "shared_go_report_command_mismatch" "shared command mismatch marker"
)

shared_reuse_results="$(mktemp -d "$ROOT_DIR/tmp/run-go-target-shared-reuse.XXXXXX")"
cleanup_paths+=("$shared_reuse_results")
(
  export CARTULARY_TEST_RESULTS_DIR="$shared_reuse_results/results"
  export CARTULARY_TEST_RUN_ID="shared-reuse"
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"

  shared_dir="$(prepare_shared_artifact_dir backend-integration-core)"
  mkdir -p "$shared_dir"
  printf '%s\n' "$core_shared_command" >"$shared_dir/command.txt"
  touch "$shared_dir/complete"

  assign_named_shared_report reused_dir reused_usage backend-integration-support backend-integration-core
  if [[ "$reused_dir" != "$shared_dir" ]]; then
    fail "shared reuse: expected assign_named_shared_report to reuse the existing shared dir"
  fi
  if [[ "$reused_usage" != "reused" ]]; then
    fail "shared reuse: expected assign_named_shared_report to mark the report as reused"
  fi
)

backend_store_structure="$(
  export NODE_BIN="$node_bin"
  source "$GO_TARGET_HELPER"
  assign_calls=0
  raw_calls=0
  manifest_calls=0
  finish_status=""
  assign_captured_report() {
    local -n dir_ref="$1"
    local -n usage_ref="$2"
    assign_calls=$((assign_calls + 1))
    dir_ref="/tmp/backend-store-shared"
    usage_ref="actual"
  }
  emit_go_raw_phase() {
    raw_calls=$((raw_calls + 1))
  }
  emit_go_manifest_phase() {
    manifest_calls=$((manifest_calls + 1))
  }
  clear_go_selection_env() {
    :
  }
  finish_target() {
    finish_status="$1"
    printf "assign_calls=%s raw_calls=%s manifest_calls=%s finish_status=%s\n" "$assign_calls" "$raw_calls" "$manifest_calls" "$finish_status"
  }
  run_backend_store
)"
assert_contains "$backend_store_structure" "assign_calls=1" "backend-store single shared capture"
assert_contains "$backend_store_structure" "raw_calls=1" "backend-store raw phase count"
assert_contains "$backend_store_structure" "manifest_calls=2" "backend-store derived phase count"
assert_contains "$backend_store_structure" "finish_status=0" "backend-store finish status"

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
cleanup_paths+=(
  "$manifest_smoke_dir"
  "$ROOT_DIR/tools/phase20_test_map.json"
  "$ROOT_DIR/tools/phase21_test_map.json"
  "$ROOT_DIR/tools/phase22_test_map.json"
  "$ROOT_DIR/tools/phase23_test_map.json"
  "$ROOT_DIR/tools/phase24_test_map.json"
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

cat >"$ROOT_DIR/tools/phase20_test_map.json" <<EOF
{
  "expected_ids": ["U-20-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase20Unit_Registered",
      "selection_pattern": "TestSupportPhase20Unit_"
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase20 >/dev/null

cat >"$ROOT_DIR/tools/phase21_test_map.json" <<EOF
{
  "expected_ids": ["U-21-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase21Unit_Missing",
      "selection_pattern": "TestSupportPhase21Unit_"
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase21_output="$(
  NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase21 \
    2>&1
)"
phase21_status=$?
set -e
if [[ "$phase21_status" -eq 0 ]]; then
  fail "phase21 support manifest: expected validation failure"
fi
assert_contains "$phase21_output" "not found in" "phase21 missing support symbol"

cat >"$ROOT_DIR/tools/phase22_test_map.json" <<EOF
{
  "expected_ids": ["U-22-01"],
  "support_go_targets": [
    {
      "target": "backend_process_support",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase22Unit_Registered",
      "selection_pattern": "TestSupportPhase22Unit_"
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase22_output="$(
  NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase22 \
    2>&1
)"
phase22_status=$?
set -e
if [[ "$phase22_status" -eq 0 ]]; then
  fail "phase22 support manifest: expected validation failure"
fi
assert_contains "$phase22_output" "must declare target=backend_unit|backend_integration_support" "phase22 invalid support target"

cat >"$ROOT_DIR/tools/phase23_test_map.json" <<EOF
{
  "expected_ids": ["U-23-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "$manifest_smoke_rel",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase23Unit_Registered",
      "selection_pattern": "TestSupportPhase23Integration_"
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase23_output="$(
  NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase23 \
    2>&1
)"
phase23_status=$?
set -e
if [[ "$phase23_status" -eq 0 ]]; then
  fail "phase23 support manifest: expected validation failure"
fi
assert_contains "$phase23_output" "selection_pattern does not match symbol" "phase23 selection pattern mismatch"

cat >"$ROOT_DIR/tools/phase24_test_map.json" <<EOF
{
  "expected_ids": ["U-24-01"],
  "support_go_targets": [
    {
      "target": "backend_unit",
      "section": "unit",
      "package": "./internal/modules/auth",
      "file": "$manifest_smoke_file",
      "symbol": "TestSupportPhase24Unit_Registered",
      "selection_pattern": "TestSupportPhase24Unit_"
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
set +e
phase24_output="$(
  NODE_BIN="$node_bin" "$node_bin" "$PHASE_MAP_CHECK" phase24 \
    2>&1
)"
phase24_status=$?
set -e
if [[ "$phase24_status" -eq 0 ]]; then
  fail "phase24 support manifest: expected validation failure"
fi
assert_contains "$phase24_output" "does not belong to package" "phase24 package mismatch"
