#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

exec "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
  "$ROOT_DIR/scripts/lib/run-playwright-webserver-batch.sh" \
  functional \
  -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  --config playwright.webserver-backed.config.ts
