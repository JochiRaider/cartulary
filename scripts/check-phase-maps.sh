#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

for phase in phase0 phase1 phase2 phase3 phase4; do
  manifest="$ROOT_DIR/tools/${phase}_test_map.json"
  if [[ ! -f "$manifest" ]]; then
    echo "required phase test map manifest missing: $manifest" >&2
    exit 1
  fi
  "$NODE_BIN" "$ROOT_DIR/scripts/check-phase-map.mjs" "$phase"
done
