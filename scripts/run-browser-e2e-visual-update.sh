#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

env CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS=1 \
  CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-browser-e2e-visual-update}" \
  NODE_BIN="$node_bin" \
  "$ROOT_DIR/scripts/start-web-e2e.sh" -- "$ROOT_DIR/scripts/run-browser-e2e-visual.sh"
