#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
visual_smoke_phase="${CARTULARY_BROWSER_VISUAL_SMOKE_PHASE:-phase3}"

CARTULARY_PHASE_SLICE_PHASE="$visual_smoke_phase" \
  exec "$ROOT_DIR/scripts/run-browser-e2e-visual.sh"
