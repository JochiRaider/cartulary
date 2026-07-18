#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

"$node_bin" "$ROOT_DIR/tools/harness/test-catalog/semantic-identity-check-cli.mjs"

printf 'semantic identity check smoke passed\n'
