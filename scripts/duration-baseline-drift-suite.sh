#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
MAKE_BIN="${MAKE:-make}"

run_drift_target() {
  local target="$1"
  env -u CARTULARY_TEST_TARGET CARTULARY_SUPPRESS_CHILD_SUCCESS=1 "${MAKE_BIN}" --no-print-directory -C "$ROOT_DIR" "$target"
}

run_drift_target go-test-duration-baseline-drift
run_drift_target browser-e2e-duration-baseline-drift
run_drift_target service-backed-make-target-duration-baseline-drift
run_drift_target harness-smoke-duration-baseline-drift
