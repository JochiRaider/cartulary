#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

status=0
playwright_update_args=()
if [[ "${CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS:-0}" == "1" ]]; then
  playwright_update_args=(--update-snapshots=all)
fi

"$ROOT_DIR/scripts/run-browser-e2e-manifest-dependency.sh" \
  browser-e2e-visual authoritative browser_visual -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test "${playwright_update_args[@]}" ||
  status=1

frontend_grep=""
frontend_scope="${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}"
if [[ "$frontend_scope" != "disabled" ]]; then
  frontend_grep_args=(playwright-grep browser-e2e-visual visual)
  if [[ "$frontend_scope" == "selected_rows" ]]; then
    frontend_grep_args+=(--row-ids "${CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:-}")
  fi
  frontend_grep="$(
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/frontend/frontend-phase-manifest.mjs" \
        "${frontend_grep_args[@]}"
  )"
fi

if [[ -n "$frontend_grep" ]]; then
  "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
    "$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh" \
    "browser-e2e-visual frontend-readiness" -- \
    "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test "${playwright_update_args[@]}" \
    apps/web/e2e/workbook.visual.spec.ts -g "$frontend_grep" ||
    status=1
fi

exit "$status"
