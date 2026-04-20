#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

mapfile -t manifests < <(find "$ROOT_DIR/tools" -maxdepth 1 -type f -name '*_test_map.json' | sort)
if [[ ${#manifests[@]} -eq 0 ]]; then
  echo "no phase test map manifests found under $ROOT_DIR/tools" >&2
  exit 1
fi

for manifest in "${manifests[@]}"; do
  phase="$(basename "$manifest" _test_map.json)"
  "$NODE_BIN" "$ROOT_DIR/scripts/check-phase-map.mjs" "$phase"
done
