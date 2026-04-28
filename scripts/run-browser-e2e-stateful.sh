#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"
exec "$ROOT_DIR/scripts/run-browser-e2e-manifest-dependency.sh" \
  browser-e2e-stateful authoritative browser_stateful -- \
  "$ROOT_DIR/scripts/run-browser-e2e-owned-stack.sh" "$@"
