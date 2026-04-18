#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"

cd "$ROOT_DIR"
make --no-print-directory generate
make --no-print-directory test
make --no-print-directory lint
make --no-print-directory check

echo "provider-neutral CI contract passed"
