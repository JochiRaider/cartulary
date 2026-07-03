#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
exec "${ROOT_DIR}/tools/harness/frontend/frontend-install.sh" "$@"
