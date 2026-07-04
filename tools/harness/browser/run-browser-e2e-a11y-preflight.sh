#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/execution/run-phase-common.sh"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

summary_dir="$(prepare_target_support_dir accessibility-preflight)"
summary_path="${summary_dir}/frontend-accessibility-preflight-summary.json"

set +e
"${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
  "$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh" \
  "browser-e2e-a11y-preflight accessibility preflight" -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  apps/web/e2e/workbook.a11y-preflight.spec.ts
status=$?
set -e

phase_dir="$(find "$(ensure_target_artifact_dir)" -maxdepth 1 -type d -name 'browser-e2e-a11y-preflight-accessibility-preflight' -print -quit)"
if [[ -z "$phase_dir" ]]; then
  phase_dir="$(ensure_target_artifact_dir)"
fi

if [[ "$status" -eq 0 ]]; then
  summary_status=pass
else
  summary_status=fail
fi

"${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/browser/accessibility-summary-cli.mjs" \
  --output "$summary_path" \
  --status "$summary_status" \
  --phase-dir "$phase_dir" \
  --mode preflight

exit "$status"
