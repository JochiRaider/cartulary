#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"
if [[ -z "${CARTULARY_OUTPUT_MODE+x}" ]]; then
  export CARTULARY_OUTPUT_MODE=ci
fi
exec make --no-print-directory ci
