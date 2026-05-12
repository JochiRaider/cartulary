#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-${ROOT_DIR}/tmp/node-runtime/bin/node}"

if [[ ! -x "${NODE_BIN}" ]]; then
  NODE_BIN="node"
fi

exec "${NODE_BIN}" "${ROOT_DIR}/scripts/agent-finalize.mjs"
