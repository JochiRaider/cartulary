#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(unset CDPATH && cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NODE_CANDIDATE="${NODE_BIN:-}"

if [[ -n "${NODE_CANDIDATE}" && -x "${NODE_CANDIDATE}" ]]; then
  exec "${NODE_CANDIDATE}" "${SCRIPT_DIR}/test-output.mjs" "$@"
fi

if [[ -x "${SCRIPT_DIR}/../../tmp/node-runtime/bin/node" ]]; then
  exec "${SCRIPT_DIR}/../../tmp/node-runtime/bin/node" "${SCRIPT_DIR}/test-output.mjs" "$@"
fi

if command -v node >/dev/null 2>&1; then
  exec "$(command -v node)" "${SCRIPT_DIR}/test-output.mjs" "$@"
fi

echo "node is required to run test-output.mjs" >&2
exit 1
