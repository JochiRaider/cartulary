#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi
exec env \
  CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-browser-e2e-visual-update}" \
  NODE_BIN="$node_bin" \
  "$ROOT_DIR/tools/harness/browser/run-browser-e2e-target.sh" visual --mode snapshot_update
