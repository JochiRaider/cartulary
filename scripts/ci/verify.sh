#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"
CI=1 make --no-print-directory check

echo "provider-neutral CI contract passed"
