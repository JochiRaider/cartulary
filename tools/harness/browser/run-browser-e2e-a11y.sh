#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/execution/phase-runtime.sh"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"
runtime_profile_id="${CARTULARY_BROWSER_RUNTIME_PROFILE_ID:-default}"

summary_dir="$(prepare_target_support_dir accessibility)"
summary_path="${summary_dir}/frontend-accessibility-summary.json"
contrast_dir="${summary_dir}/profiles/${runtime_profile_id}/contrast-checks"
phase_secure_mkdir "$contrast_dir"
phase_label="browser-e2e-a11y accessibility-${runtime_profile_id}"
playwright_args=(apps/web/e2e/workbook.a11y.spec.ts)
if [[ "$runtime_profile_id" != "default" ]]; then
  frontend_grep="$(
    "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "$ROOT_DIR/tools/harness/phase-accounting/frontend-phase-manifest.mjs" \
      playwright-grep browser-e2e-a11y accessibility \
      --runtime-profile-id "$runtime_profile_id"
  )"
  if [[ -z "$frontend_grep" ]]; then
    echo "browser-e2e-a11y has no scenarios for runtime profile $runtime_profile_id" >&2
    exit 2
  fi
  playwright_args+=(-g "$frontend_grep")
fi

set +e
"${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
  NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
  CARTULARY_FRONTEND_ACCESSIBILITY_SUMMARY="$summary_path" \
  CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR="$contrast_dir" \
  "$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh" \
  "$phase_label" -- \
  "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
  "${playwright_args[@]}"
status=$?
set -e

phase_dir="$(ensure_target_artifact_dir)/$(slugify_phase_label "$phase_label")"
if [[ -z "$phase_dir" ]]; then
  phase_dir="$(ensure_target_artifact_dir)"
fi

if [[ "$status" -eq 0 ]]; then
  summary_status=pass
else
  summary_status=fail
fi

"${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/browser/accessibility-summary-cli.mjs" \
  --output "$summary_path" \
  --status "$summary_status" \
  --phase-dir "$phase_dir" \
  --contrast-dir "$contrast_dir" \
  --runtime-profile-id "$runtime_profile_id"

exit "$status"
