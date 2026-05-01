#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/scripts/check-phase-test-names.mjs"
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
  cat >"$root/tools/phase5_test_map.json" <<'JSON'
{
  "expected_ids": ["U-5-01"],
  "unit": [
    {
      "id": "U-5-01",
      "coverage": "authoritative",
      "runner": "go_test",
      "file": "internal/modules/future/phase5_test.go",
      "symbol": "TestPhase5_Create_U_5_01",
      "execution_dependency": "backend_unit",
      "evidence_layer": "unit",
      "claim": "future unit evidence",
      "out_of_scope": "none"
    }
  ],
  "e2e": [
    {
      "id": "E-5-SMOKE-01",
      "coverage": "supplemental",
      "runner": "go_test",
      "file": "cmd/server/main_phase5_process_test.go",
      "symbol": "TestPhase5_Process_E_5_SMOKE_01_ProcessSmoke",
      "execution_dependency": "backend_process",
      "evidence_layer": "process_smoke",
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

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/phase-test-names.XXXXXX")"
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

unknown_root="$tmp_dir/unknown"
write_case "$unknown_root" "internal/modules/future" 'package future

import "testing"

func TestPhase6_Create_U_6_01(t *testing.T) {}
'
unknown_output="$(assert_fails "unknown phase" run_checker "$unknown_root" "$manifest_root")"
assert_contains "$unknown_output" "no phase6_test_map.json manifest exists" "unknown phase output"

undeclared_root="$tmp_dir/undeclared"
write_case "$undeclared_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_Create_U_5_02(t *testing.T) {}
'
undeclared_output="$(assert_fails "undeclared ID" run_checker "$undeclared_root" "$manifest_root")"
assert_contains "$undeclared_output" "TestPhase5_Create_U_5_02" "undeclared ID output"
assert_contains "$undeclared_output" "U_5_01" "undeclared ID expected fragment output"

legacy_root="$tmp_dir/legacy"
write_case "$legacy_root" "internal/modules/future" 'package future

import "testing"

func TestPhase5_ProcessSmoke_Create(t *testing.T) {}
'
legacy_output="$(assert_fails "legacy process smoke" run_checker "$legacy_root" "$manifest_root")"
assert_contains "$legacy_output" "TestPhase5_ProcessSmoke_Create" "legacy process smoke output"
assert_contains "$legacy_output" "E_5_SMOKE_01" "legacy process smoke expected fragment output"
