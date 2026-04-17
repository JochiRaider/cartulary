#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR"
make generate

if ! git diff --quiet -- internal/gen/contracts packages/protocol-ts/src/generated; then
  echo "generated artifact drift detected after make generate" >&2
  git diff -- internal/gen/contracts packages/protocol-ts/src/generated
  exit 1
fi
