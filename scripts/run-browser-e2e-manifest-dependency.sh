#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

usage() {
  echo "usage: run-browser-e2e-manifest-dependency.sh <target> <coverage> <execution_dependency> -- <playwright command...>" >&2
  exit 2
}

if [[ "$#" -lt 5 ]]; then
  usage
fi

target="$1"
coverage="$2"
execution_dependency="$3"
shift 3

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

resolve_playwright_owned_stack_env "$ROOT_DIR"

mapfile -t phases < <(
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
    "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/scripts/lib/phase-manifest.mjs" \
      playwright-phases "$coverage" "$execution_dependency"
)

if [[ "${#phases[@]}" -eq 0 ]]; then
  echo "no $coverage Playwright phases found for $execution_dependency" >&2
  exit 1
fi

status=0
for phase in "${phases[@]}"; do
  if ! "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
    "$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh" \
    "$target $phase $coverage" \
    "$phase" "$coverage" "$execution_dependency" -- \
    "$@"; then
    status=1
  fi
done

exit "$status"
