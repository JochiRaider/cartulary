#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"
source "$ROOT_DIR/scripts/lib/run-phase-common.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"

measurement_metadata_dir="$(prepare_target_support_dir ordinary-measurement)"
measurement_metadata_file="$measurement_metadata_dir/ordinary-measurement-metadata.json"
cat >"$measurement_metadata_file" <<EOF
{
  "schema_id": "cartulary.ordinary_measurement_metadata.v1",
  "evidence_kind": "ordinary_measurement",
  "claim_bearing": false,
  "target": "browser-e2e-measurement",
  "execution_dependency": "browser_measurement",
  "sample_count_per_predicate": 12,
  "warmup_samples_per_predicate": 1,
  "benchmark_manifest_schema_id": null,
  "benchmark_profile_id": null,
  "artifact_bundle_sha256": null
}
EOF

exec "$ROOT_DIR/scripts/run-browser-e2e-manifest-dependency.sh" \
  browser-e2e-measurement authoritative browser_measurement -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test
