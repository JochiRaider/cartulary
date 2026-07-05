#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"

bash "$ROOT_DIR/tools/harness/smoke/target-plan-smoke-suite.sh" diagnostics
