#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

NODE_VERSION="${NODE_VERSION:-24.15.0}"
export NODE_VERSION

exec "${ROOT_DIR}/tools/harness/readiness/bootstrap-node-runtime.sh" "$@"
