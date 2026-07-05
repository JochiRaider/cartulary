#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

bash "$ROOT_DIR/tools/harness/diagnostics/tests/test-target-plan-smoke.sh"
bash "$ROOT_DIR/tools/harness/backend/tests/test-go-shard-plan-smoke.sh"
bash "$ROOT_DIR/tools/harness/phase-accounting/tests/test-phase-registry-target-plan-smoke.sh"
