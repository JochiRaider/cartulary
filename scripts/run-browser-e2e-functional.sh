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
  "browser-e2e-functional phase1 authoritative" \
  phase1 authoritative browser_functional -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test

"${manifest_env[@]}" \
  "$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh" \
  "browser-e2e-functional phase2 authoritative" \
  phase2 authoritative -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test

"${manifest_env[@]}" \
  "$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh" \
  "browser-e2e-functional phase3 authoritative" \
  phase3 authoritative browser_functional -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test

exec "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
  "$ROOT_DIR/scripts/lib/run-playwright-phase.sh" \
  "browser-e2e-functional phase4 raw" \
  -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  e2e/phase4.spec.ts
