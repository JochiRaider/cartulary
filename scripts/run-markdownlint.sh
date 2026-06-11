#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$ROOT_DIR/tmp/node-runtime}"
PNPM_BIN="${PNPM:-$NODE_RUNTIME_DIR/bin/pnpm}"

if [[ ! -x "$PNPM_BIN" ]]; then
  echo "repo-local pnpm was not found at $PNPM_BIN; run make frontend-toolchain" >&2
  exit 127
fi

status=0
PATH="$NODE_RUNTIME_DIR/bin:$PATH" \
  COREPACK_HOME="${COREPACK_HOME:-$NODE_RUNTIME_DIR/corepack}" \
  "$PNPM_BIN" --dir "$ROOT_DIR" exec markdownlint-cli2 --config "$ROOT_DIR/.markdownlint-cli2.jsonc" "$@" || status=$?

if [[ "$status" -eq 0 && "${CARTULARY_TEST_TARGET:-}" == "lint-markdown" ]]; then
  NODE_BIN="${NODE_BIN:-$NODE_RUNTIME_DIR/bin/node}"
  TEST_OUTPUT_SCRIPT="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/scripts/lib/test-output.mjs}"
  "$NODE_BIN" "$TEST_OUTPUT_SCRIPT" target-summary lint-markdown pass --quiet-success --suppress-machine-output || status=$?
fi

exit "$status"
