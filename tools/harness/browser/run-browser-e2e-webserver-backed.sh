#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

exec "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
  "$ROOT_DIR/tools/harness/browser/run-playwright-webserver-batch.sh" \
  webserver-backed \
  -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  --config playwright.webserver-backed.config.ts
