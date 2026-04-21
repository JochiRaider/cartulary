#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/run-phase-common.sh"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

"$ROOT_DIR/scripts/run-browser-e2e-functional.sh"

run_phase_command \
  "browser-e2e-support phase2" \
  "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  e2e/phase2.support.spec.ts
