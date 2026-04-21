#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-phase.sh"
GO_HELPER="$ROOT_DIR/scripts/lib/run-go-phase.sh"
GO_MANIFEST_HELPER="$ROOT_DIR/scripts/lib/run-go-manifest-phase.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output to omit [$needle]"
  fi
}

extract_log_path() {
  printf '%s\n' "$1" | sed -n 's/^phase log: //p' | tail -n 1
}

quiet_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "quiet success" -- bash -lc 'echo hidden-success-output'
)"
assert_contains "$quiet_success_output" "== quiet success ==" "quiet success banner"
assert_not_contains "$quiet_success_output" "hidden-success-output" "quiet success suppresses routine output"

success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
    "$HELPER" "success log replay" -- bash -lc 'echo keep-this-warning >&2'
)"
assert_contains "$success_log_output" "== success log replay ==" "success log replay banner"
assert_contains "$success_log_output" "keep-this-warning" "success log replay output"

set +e
short_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "short failure" -- bash -lc 'echo short-failure >&2; exit 7' \
    2>&1
)"
short_failure_status=$?
set -e
if [[ "$short_failure_status" -ne 7 ]]; then
  fail "short failure: expected exit status 7, got $short_failure_status"
fi
assert_contains "$short_failure_output" "== short failure ==" "short failure banner"
assert_contains "$short_failure_output" "phase failed: short failure" "short failure label"
assert_contains "$short_failure_output" "failing command: bash -lc" "short failure command"
assert_contains "$short_failure_output" "short-failure" "short failure output"
short_failure_log="$(extract_log_path "$short_failure_output")"
if [[ -z "$short_failure_log" || ! -f "$short_failure_log" ]]; then
  fail "short failure: expected a preserved phase log path"
fi
cleanup_paths+=("$short_failure_log")

set +e
long_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "long failure" -- bash -lc 'for i in $(seq 1 230); do echo "line-$i"; done; exit 9' \
    2>&1
)"
long_failure_status=$?
set -e
if [[ "$long_failure_status" -ne 9 ]]; then
  fail "long failure: expected exit status 9, got $long_failure_status"
fi
assert_contains "$long_failure_output" "== long failure ==" "long failure banner"
assert_contains "$long_failure_output" "line-1" "long failure first excerpt"
assert_contains "$long_failure_output" "line-40" "long failure first excerpt boundary"
assert_not_contains "$long_failure_output" "line-70" "long failure middle omission"
assert_contains "$long_failure_output" "line-71" "long failure last excerpt start"
assert_contains "$long_failure_output" "line-230" "long failure last excerpt end"
long_failure_log="$(extract_log_path "$long_failure_output")"
if [[ -z "$long_failure_log" || ! -f "$long_failure_log" ]]; then
  fail "long failure: expected a preserved phase log path"
fi
cleanup_paths+=("$long_failure_log")

verbose_override_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$HELPER" "verbose override" -- bash -lc 'echo verbose-stream'
)"
assert_contains "$verbose_override_output" "== verbose override ==" "verbose override banner"
assert_contains "$verbose_override_output" "verbose-stream" "verbose override output"

mkdir -p "$ROOT_DIR/tmp"
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
assert_contains "$go_success_output" "== run-go-phase smoke ==" "run-go-phase success banner"
assert_contains "$go_success_output" "matched go tests: 2 across 1 packages" "run-go-phase matched count"
assert_contains "$go_success_output" "TestPhase0_RunGoPhase_E_0_01" "run-go-phase matched first numeric suffix test"
assert_contains "$go_success_output" "TestPhase0_RunGoPhase_E_0_02" "run-go-phase matched second numeric suffix test"

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
assert_contains "$go_zero_output" "phase matched zero tests" "run-go-phase zero-match output"

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
assert_contains "$go_skip_output" "go test inventory requires top-level pass" "run-go-phase skip inventory failure"
assert_contains "$go_skip_output" "TestPhase0_RunGoPhaseSkip_E_0_01" "run-go-phase skip test name"

go_manifest_dir="$(mktemp -d "$ROOT_DIR/tmp/run-go-manifest-phase-smoke.XXXXXX")"
cleanup_paths+=("$go_manifest_dir" "$ROOT_DIR/tools/phase9_test_map.json" "$ROOT_DIR/tools/phase10_test_map.json")
cat >"$go_manifest_dir/run_go_manifest_phase_smoke_test.go" <<'EOF'
package rungomanifestphasesmoke

import "testing"

func TestPhase9_RunGoManifest_U_9_01(t *testing.T) {}

func TestPhase10_RunGoManifest_U_10_01(t *testing.T) {
	t.Skip("matched manifest skip")
}
EOF

go_manifest_rel="./${go_manifest_dir#"$ROOT_DIR"/}"
cat >"$ROOT_DIR/tools/phase9_test_map.json" <<EOF
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF
cat >"$ROOT_DIR/tools/phase10_test_map.json" <<EOF
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF

node_bin="${NODE_BIN:-node}"
go_manifest_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase smoke" phase9 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_rel"
)"
assert_contains "$go_manifest_success_output" "== run-go-manifest-phase smoke ==" "run-go-manifest-phase success banner"
assert_contains "$go_manifest_success_output" "matched go manifest tests: 1" "run-go-manifest-phase matched count"
assert_contains "$go_manifest_success_output" "TestPhase9_RunGoManifest_U_9_01" "run-go-manifest-phase matched test"

set +e
go_manifest_skip_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="$node_bin" \
    "$GO_MANIFEST_HELPER" "run-go-manifest-phase skip" phase10 unit authoritative backend_unit -- "$go_bin" test "$go_manifest_rel" \
    2>&1
)"
go_manifest_skip_status=$?
set -e
if [[ "$go_manifest_skip_status" -eq 0 ]]; then
  fail "run-go-manifest-phase skip: expected non-zero exit status"
fi
assert_contains "$go_manifest_skip_output" "manifest-go execution mismatch" "run-go-manifest-phase skip verification failure"
assert_contains "$go_manifest_skip_output" "skipped=TestPhase10_RunGoManifest_U_10_01" "run-go-manifest-phase skip test name"
