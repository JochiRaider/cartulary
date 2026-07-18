#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/tools/harness/phase-accounting/phase-test-name-check-cli.mjs"
cleanup_paths=()
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

cleanup() {
  local scratch_path
  for scratch_path in "${cleanup_paths[@]}"; do
    rm -rf "$scratch_path"
  done
}
trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

run_checker() {
  local case_root="$1"
  (cd "$case_root" && "$NODE_BIN" "$CHECKER")
}

assert_passes() {
  local label="$1"
  local case_root="$2"
  local output
  if ! output="$(run_checker "$case_root" 2>&1)"; then
    fail "$label: expected success, got: $output"
  fi
}

assert_fails_with() {
  local label="$1"
  local case_root="$2"
  local expected="$3"
  local output
  local status
  set +e
  output="$(run_checker "$case_root" 2>&1)"
  status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  if [[ "$output" != *"$expected"* ]]; then
    fail "$label: expected [$expected], got: $output"
  fi
}

write_case() {
  local case_root="$1"
  local filename="$2"
  local source="$3"
  mkdir -p "$case_root/internal/modules/future"
  printf '%s\n' "$source" >"$case_root/internal/modules/future/$filename"
}

scratch_root="$(cartulary_harness_mktemp_dir "semantic-go-identities.XXXXXX")"
cleanup_paths+=("$scratch_root")

semantic_root="$scratch_root/semantic"
write_case "$semantic_root" "create_test.go" 'package future

import "testing"

type createFixture struct{}

func TestCreateRecord_Unit(t *testing.T) {}
func newCreateFixture() createFixture { return createFixture{} }
'
assert_passes "semantic Go identities" "$semantic_root"

filename_root="$scratch_root/filename"
write_case "$filename_root" "phase5_create_test.go" 'package future

import "testing"

func TestCreateRecord_Unit(t *testing.T) {}
'
assert_fails_with "delivery-shaped filename" "$filename_root" "phase5_create_test.go"

test_symbol_root="$scratch_root/test-symbol"
write_case "$test_symbol_root" "create_test.go" 'package future

import "testing"

func TestPhase5_Create_U_5_01(t *testing.T) {}
'
assert_fails_with "delivery-shaped test symbol" "$test_symbol_root" "TestPhase5_Create_U_5_01"

helper_root="$scratch_root/helper"
write_case "$helper_root" "create_test.go" 'package future

type phase5Fixture struct{}

func newPhase5Fixture() phase5Fixture { return phase5Fixture{} }
'
assert_fails_with "delivery-shaped helper symbol" "$helper_root" "phase5Fixture"

catalog_root="$scratch_root/catalog"
write_case "$catalog_root" "create_test.go" 'package future

import "testing"

func TestCreateRecord_Unit(t *testing.T) {}
'
mkdir -p "$catalog_root/tools/test_families"
cat >"$catalog_root/tools/test_families/module.future.json" <<'JSON'
{
  "schema_id": "cartulary.test_family_manifest.v1",
  "owner_id": "module.future",
  "rows": [
    {
      "row_id": "module.future.unit.create_record",
      "runner": "go",
      "selector": {
        "package": "./internal/modules/future",
        "tests": ["TestPhase5_Create_U_5_01"]
      }
    }
  ]
}
JSON
assert_fails_with "delivery-shaped catalog selector" "$catalog_root" "Go selector for module.future.unit.create_record"
