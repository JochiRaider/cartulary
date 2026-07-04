#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

(cd "$ROOT_DIR" && "$NODE_BIN" "$ROOT_DIR/tools/harness/phase-accounting/phase-registry.mjs" validate)
(cd "$ROOT_DIR" && "$NODE_BIN" "$ROOT_DIR/tools/harness/phase-accounting/frontend-phase-manifest.mjs" validate)

mapfile -t phases < <(cd "$ROOT_DIR" && "$NODE_BIN" "$ROOT_DIR/tools/harness/phase-accounting/phase-manifest.mjs" list-registered-manifest-phases)

(cd "$ROOT_DIR" && "$NODE_BIN" "$ROOT_DIR/tools/harness/phase-accounting/phase-manifest.mjs" phase-policy-exceptions-validate)

for phase in "${phases[@]}"; do
  manifest="$ROOT_DIR/tools/${phase}_test_map.json"
  if [[ -z "${CARTULARY_PHASE_MANIFEST_ROOT:-}" && ! -f "$manifest" ]]; then
    echo "required phase test map manifest missing: $manifest" >&2
    exit 1
  fi
  "$NODE_BIN" "$ROOT_DIR/tools/harness/phase-accounting/phase-map-check-cli.mjs" "$phase"
done
