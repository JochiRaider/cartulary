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

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_not_zero() {
  local actual="$1"
  local label="$2"

  if [[ -z "$actual" || "$actual" == "0" ]]; then
    fail "$label: expected a non-zero value, got [$actual]"
  fi
}

node_bin="${NODE_BIN:-node}"
go_bin="${GO:-go}"

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
