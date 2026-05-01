#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"
exec make --no-print-directory ci
