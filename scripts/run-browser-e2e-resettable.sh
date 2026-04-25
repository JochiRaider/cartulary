#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/run-phase-common.sh"

run_child_target() {
  local target="$1"
  shift

  env CARTULARY_TEST_TARGET="$target" "$@"
  NODE_BIN="${NODE_BIN:-}" "$TEST_OUTPUT_HELPER" target-summary "$target" pass
}

run_child_target browser-e2e-measurement "$ROOT_DIR/scripts/run-browser-e2e-measurement.sh"

env CARTULARY_TEST_TARGET=browser-e2e-resettable \
  "$ROOT_DIR/scripts/reset-web-e2e-stack.sh" --label measurement-to-visual

run_child_target browser-e2e-visual "$ROOT_DIR/scripts/run-browser-e2e-visual.sh"
