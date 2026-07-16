#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"

node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

stage_metadata="$($node_bin "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-runner "$MANIFEST" visual)"
declare -A coverage_by_group=()
declare -A profile_by_group=()
while IFS= read -r group_row; do
  group_row_fields="${group_row//$'\t'/$'\x1f'}"
  IFS=$'\x1f' read -r group_name _target _kind _workers _reset_before coverage _execution_dependency _stage_schedule_tags _stage_scheduler_needs _selected_phase _selected_row_ids _browser_session_group _browser_session_isolation_reason runtime_profile_id <<<"$group_row_fields"
  coverage_by_group["$group_name"]="$coverage"
  profile_by_group["$group_name"]="$runtime_profile_id"
done < <(printf '%s\n' "$stage_metadata" | tail -n +3)

status=0
mapfile -t stage_sessions < <(
  "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-sessions "$MANIFEST" visual
)
for session_row in "${stage_sessions[@]}"; do
  session_row_fields="${session_row//$'\t'/$'\x1f'}"
  IFS=$'\x1f' read -r session_group runtime_profile_id selected_groups <<<"$session_row_fields"
  session_coverage=""
  IFS=',' read -r -a selected_group_names <<<"$selected_groups"
  for group_name in "${selected_group_names[@]}"; do
    group_coverage="${coverage_by_group[$group_name]:-}"
    group_profile="${profile_by_group[$group_name]:-}"
    if [[ -z "$group_coverage" || "$group_profile" != "$runtime_profile_id" ]]; then
      echo "visual update session $session_group has invalid group $group_name" >&2
      exit 2
    fi
    if [[ -n "$session_coverage" && "$session_coverage" != "$group_coverage" ]]; then
      echo "visual update session $session_group mixes coverage classes" >&2
      exit 2
    fi
    session_coverage="$group_coverage"
  done

  session_status=0
  env CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS=1 \
    CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-browser-e2e-visual-update}" \
    CARTULARY_BROWSER_SESSION_GROUP="$session_group" \
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID="$runtime_profile_id" \
    CARTULARY_BROWSER_GROUP_COVERAGE="$session_coverage" \
    CARTULARY_BROWSER_GROUP_EXECUTION_DEPENDENCY="browser_visual" \
    CARTULARY_PLAYWRIGHT_WORKER_COUNT="1" \
    CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET="0" \
    PLAYWRIGHT_WORKERS="1" \
    NODE_BIN="$node_bin" \
    "$ROOT_DIR/tools/harness/browser/start-web-e2e.sh" -- "$ROOT_DIR/tools/harness/browser/run-browser-e2e-visual.sh" || session_status=$?
  if [[ "$session_status" -ne 0 && "$status" -eq 0 ]]; then
    status="$session_status"
  fi
done

exit "$status"
