#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
exec "${NODE_BIN:-node}" "${ROOT_DIR}/tools/harness/execution/sequence-schedule-cli.mjs" "$@"
