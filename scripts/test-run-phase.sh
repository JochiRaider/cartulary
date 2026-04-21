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

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

quiet_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "quiet success" -- bash -lc 'echo hidden-success-output'
)"
assert_empty "$quiet_success_output" "quiet success"

success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
    "$HELPER" "success log replay" -- bash -lc 'echo keep-this-warning >&2' \
    2>&1
)"
assert_contains "$success_log_output" "keep-this-warning" "success log replay output"
assert_not_contains "$success_log_output" "== success log replay ==" "success log replay banner"

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
assert_contains "$short_failure_output" "failure: short failure" "short failure label"
assert_contains "$short_failure_output" "runner=shell" "short failure runner"
assert_contains "$short_failure_output" "message=short-failure" "short failure message"
assert_contains "$short_failure_output" "raw=" "short failure raw path"
assert_not_contains "$short_failure_output" "== short failure ==" "short failure banner"

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
cleanup_paths+=("$go_manifest_dir" "$ROOT_DIR/tools/phase9_test_map.json" "$ROOT_DIR/tools/phase10_test_map.json" "$ROOT_DIR/tools/phase11_test_map.json")
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
assert_empty "$go_manifest_success_output" "run-go-manifest-phase success"

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
cat >"$ROOT_DIR/tools/phase11_test_map.json" <<EOF
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
      "evidence_layer": "smoke"
    }
  ]
}
EOF

set +e
go_manifest_pkg_setup_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
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
