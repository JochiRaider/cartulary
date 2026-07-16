#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"
source "$ROOT_DIR/tools/harness/execution/phase-runtime.sh"

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
  "sample_count_per_predicate": 25,
  "warmup_samples_per_predicate": 1,
  "benchmark_manifest_schema_id": null,
  "benchmark_profile_id": null,
  "artifact_bundle_sha256": null
}
EOF

status=0
group_coverage="${CARTULARY_BROWSER_GROUP_COVERAGE:-authoritative}"
runtime_profile_id="${CARTULARY_BROWSER_RUNTIME_PROFILE_ID:-default}"

if [[ "$runtime_profile_id" == "default" ]]; then
  "$ROOT_DIR/tools/harness/browser/run-browser-e2e-manifest-dependency.sh" \
    browser-e2e-measurement "$group_coverage" browser_measurement -- \
    "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test ||
    status=1
fi

frontend_grep=""
frontend_scope="${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}"
if [[ "$frontend_scope" != "disabled" ]]; then
  frontend_grep_args=(
    playwright-grep browser-e2e-measurement browser_integration
    --runtime-profile-id "$runtime_profile_id"
  )
  if [[ "$frontend_scope" == "selected_rows" ]]; then
    frontend_grep_args+=(--row-ids "${CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:-}")
  fi
  frontend_grep="$(
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/phase-accounting/frontend-phase-manifest.mjs" \
        "${frontend_grep_args[@]}"
  )"
fi

if [[ -n "$frontend_grep" ]]; then
  "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
    "$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh" \
    "browser-e2e-measurement frontend-readiness-${runtime_profile_id}" -- \
    "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
    apps/web/e2e/measurement -g "$frontend_grep" ||
    status=1
fi

exit "$status"
