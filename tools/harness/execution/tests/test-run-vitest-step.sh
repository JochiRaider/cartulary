#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
# One real runner contract covers both selected cohorts and the watchdog/output
# wrapper. There is no synthetic subprocess implementation of the old CLI.
exec "${NODE_BIN:-node}" "$ROOT_DIR/tools/harness/execution/tests/test-vitest-diagnostics.mjs"
