#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"

MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"
TEST_OUTPUT_HELPER="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/tools/harness/output/test-output.mjs}"

usage() {
  echo "usage: run-browser-e2e-target.sh <stage>" >&2
  exit 2
}

if [[ "$#" -ne 1 ]]; then
  usage
fi

stage="$1"
node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

emit_test_output() {
  if [[ "$TEST_OUTPUT_HELPER" == *.mjs ]]; then
    "$node_bin" "$TEST_OUTPUT_HELPER" "$@"
    return $?
  fi
  NODE_BIN="$node_bin" "$TEST_OUTPUT_HELPER" "$@"
}

stage_metadata="$("$node_bin" "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-target "$MANIFEST" "$stage")"
target="$(printf '%s\n' "$stage_metadata" | sed -n '1p')"
summary_children="$(printf '%s\n' "$stage_metadata" | sed -n '2p')"

status=0
mapfile -t stage_sessions < <(
  "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-sessions "$MANIFEST" "$stage"
)
for session_row in "${stage_sessions[@]}"; do
  session_row_fields="${session_row//$'\t'/$'\x1f'}"
  IFS=$'\x1f' read -r session_group runtime_profile_id selected_groups <<<"$session_row_fields"
  session_status=0
  env CARTULARY_TEST_TARGET="$target" \
    CARTULARY_BROWSER_SESSION_GROUP="$session_group" \
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID="$runtime_profile_id" \
    CARTULARY_BROWSER_SELECTED_GROUPS="$selected_groups" \
    NODE_BIN="$node_bin" \
    "$ROOT_DIR/tools/harness/browser/start-web-e2e.sh" -- "$ROOT_DIR/tools/harness/browser/run-browser-e2e-batch.sh" "$stage" --defer-summary || session_status=$?
  if [[ "$session_status" -ne 0 && "$status" -eq 0 ]]; then
    status="$session_status"
  fi
done

mapfile -t evidence_targets < <(
  "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-runner "$MANIFEST" "$stage" |
    tail -n +3 |
    cut -f2 |
    sort -u
)
for evidence_target in "${evidence_targets[@]}"; do
  evidence_status=0
  env CARTULARY_TEST_TARGET="$evidence_target" \
    NODE_BIN="$node_bin" \
    "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-evidence-finalize-cli.mjs" "$evidence_target" || evidence_status=$?
  if [[ "$evidence_status" -ne 0 && "$status" -eq 0 ]]; then
    status="$evidence_status"
  fi
done

if [[ "$status" -eq 0 ]]; then
  requested=pass
else
  requested=fail
fi

summary_status=0
if [[ -n "$summary_children" ]]; then
  emit_test_output target-summary "$target" "$requested" --children "$summary_children" || summary_status=$?
else
  emit_test_output target-summary "$target" "$requested" || summary_status=$?
fi

if [[ "$status" -ne 0 ]]; then
  exit "$status"
fi
exit "$summary_status"
