#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/tools/harness/phase-accounting/phase-test-name-check-cli.mjs"
cleanup_paths=()
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

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
    fail "$label: expected output to contain [$needle], got [$haystack]"
  fi
}

assert_passes() {
  local label="$1"
  shift

  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label: expected success, got output: $output"
  fi
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

write_phase5_manifest() {
  local root="$1"

  mkdir -p "$root/tools"
  cat >"$root/tools/phase_registry.json" <<'JSON'
{
  "schema_id": "cartulary.phase_registry.v1",
  "phases": [
    {
      "phase": "phase5",
      "order": 5,
      "status": "active",
      "label": "Phase 5",
      "manifest_path": "tools/phase5_test_map.json",
      "ledger_path": "docs/testing/phase5_coverage_ledger.md",
      "scope": "synthetic phase5 scope.",
      "normative_owners": "Synthetic owner."
    }
  ]
}
JSON
  cat >"$root/tools/phase5_test_map.json" <<'JSON'
{
  "schema_id": "cartulary.phase_test_map.v2",
  "phase": "phase5",
  "note": "Synthetic phase test names fixture.",
  "ledger": {
    "title": "Phase 5 Coverage Ledger",
    "notes": "Synthetic phase test names fixture.",
    "authoritative_execution": "make phase-slice PHASE=phase5",
    "support_execution_extras": [],
    "sections": [],
    "shared_harness": [],
    "support_only": []
  },
  "expected_ids": ["U-5-01"],
  "support_go_targets": [],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "file": "internal/modules/future/phase5_test.go",
      "symbol": "TestPhase5_Create_U_5_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "unit",
      "evidence_class": "product_conformance",
      "layer": "backend_unit",
      "default_check_required": true,
      "default_check_kind": "primary_local_evidence",
      "default_check_reason_code": "cheapest_authoritative_layer",
      "primary_evidence_owner": "phase-test-names-fixture",
      "duplicate_of": null,
      "evidence_delta": "Synthetic phase test names fixture coverage.",
      "warm_local_cost_class": "low",
      "claim": "future unit evidence",
      "out_of_scope": "none"
    }
  ],
  "integration": [],
  "e2e": [
    {
      "id": "E-5-SMOKE-01",
      "coverage": "supplemental",
      "runner": "go_test",
      "file": "cmd/server/main_phase5_process_test.go",
      "symbol": "TestPhase5_Process_E_5_SMOKE_01_ProcessSmoke",
      "execution_dependency": "backend_process",
      "evidence_layer": "process_smoke",
      "evidence_class": "implementation_support",
      "layer": "backend_process",
      "default_check_required": false,
      "default_check_kind": "explicit_only",
      "default_check_reason_code": "implementation_support_explicit_only",
      "primary_evidence_owner": "E-5-SMOKE-01",
      "duplicate_of": null,
      "evidence_delta": "support evidence is explicit-only",
      "warm_local_cost_class": "explicit_heavy",
      "claim": "future process smoke",
      "out_of_scope": "authoritative browser evidence"
    }
  ]
}
JSON
}

write_case() {
  local case_root="$1"
  local package_dir="$2"
  local source="$3"

  mkdir -p "$case_root/$package_dir" "$case_root/cmd/server"
  printf '%s\n' "$source" >"$case_root/$package_dir/phase_test.go"
}

run_checker() {
  local case_root="$1"
  local manifest_root="$2"

  (cd "$case_root" && CARTULARY_PHASE_MANIFEST_ROOT="$manifest_root" "$NODE_BIN" "$CHECKER")
}

tmp_dir="$(cartulary_harness_mktemp_dir "phase-test-names.XXXXXX")"
cleanup_paths+=("$tmp_dir")
manifest_root="$tmp_dir/manifests"
write_phase5_manifest "$manifest_root"

valid_root="$tmp_dir/valid"
write_case "$valid_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_Create_U_5_01(t *testing.T) {}
func TestPhase5_Process_E_5_SMOKE_01_ProcessSmoke(t *testing.T) {}
'
assert_passes "phase5 manifest-owned names" run_checker "$valid_root" "$manifest_root"

support_root="$tmp_dir/support"
write_case "$support_root" "internal/modules/future" 'package future

import "testing"

func TestSupportPhase5Unit_HelperCoverage(t *testing.T) {}
'
assert_passes "phase5 support names stay outside authoritative checker" run_checker "$support_root" "$manifest_root"

unknown_root="$tmp_dir/unknown"
write_case "$unknown_root" "internal/modules/future" 'package future

import "testing"

func TestPhase6_Create_U_6_01(t *testing.T) {}
'
unknown_output="$(assert_fails "unknown phase" run_checker "$unknown_root" "$manifest_root")"
assert_contains "$unknown_output" "no active phase registry entry exists for phase6" "unknown phase output"

undeclared_root="$tmp_dir/undeclared"
write_case "$undeclared_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_Create_U_5_02(t *testing.T) {}
'
undeclared_output="$(assert_fails "undeclared ID" run_checker "$undeclared_root" "$manifest_root")"
assert_contains "$undeclared_output" "TestPhase5_Create_U_5_02" "undeclared ID output"
assert_contains "$undeclared_output" "U_5_01" "undeclared ID expected fragment output"

wrong_row_root="$tmp_dir/wrong-row"
write_case "$wrong_row_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_Create_I_5_01(t *testing.T) {}
'
wrong_row_manifest="$tmp_dir/wrong-row-manifest"
write_phase5_manifest "$wrong_row_manifest"
perl -0pi -e 's/TestPhase5_Create_U_5_01/TestPhase5_Create_I_5_01/' "$wrong_row_manifest/tools/phase5_test_map.json"
wrong_row_output="$(assert_fails "wrong row ID in authoritative manifest" run_checker "$wrong_row_root" "$wrong_row_manifest")"
assert_contains "$wrong_row_output" "TestPhase5_Create_I_5_01" "wrong row manifest output"
assert_contains "$wrong_row_output" "authoritative evidence for U-5-01 must include U-5-01 or U_5_01" "wrong row manifest reason"

legacy_root="$tmp_dir/legacy"
write_case "$legacy_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_ProcessSmoke_Create(t *testing.T) {}
'
legacy_output="$(assert_fails "legacy process smoke" run_checker "$legacy_root" "$manifest_root")"
assert_contains "$legacy_output" "TestPhase5_ProcessSmoke_Create" "legacy process smoke output"
assert_contains "$legacy_output" "E_5_SMOKE_01" "legacy process smoke expected fragment output"
