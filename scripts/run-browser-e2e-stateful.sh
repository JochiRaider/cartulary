#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"
manifest_env=(
  "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}"
)

"${manifest_env[@]}" \
  "$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh" \
  "browser-e2e-stateful phase1 authoritative" \
  phase1 authoritative browser_stateful -- \
  "$ROOT_DIR/scripts/run-browser-e2e-owned-stack.sh" "$@"
