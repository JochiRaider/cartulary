#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
node_cmd="${NODE_BIN:-}"
if [[ -z "$node_cmd" || ! -x "$node_cmd" ]]; then
  node_cmd=node
fi

exec "$node_cmd" "$ROOT_DIR/tools/harness/core/run-make-node-tool-cli.mjs" "$@"
