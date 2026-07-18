#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"
node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

status=0
while IFS=$'\t' read -r group_name _target _kind _workers _reset_before _coverage _dependency _tags _needs _selected_rows session_group _isolation_reason runtime_profile_id _specs; do
  [[ -n "$group_name" ]] || continue
  group_status=0
  env \
    CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS=1 \
    CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-browser-e2e-visual-update}" \
    CARTULARY_BROWSER_STAGE=visual \
    CARTULARY_BROWSER_GROUP_NAME="$group_name" \
    CARTULARY_BROWSER_SESSION_GROUP="$session_group" \
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID="$runtime_profile_id" \
    NODE_BIN="$node_bin" \
    "$ROOT_DIR/tools/harness/browser/start-web-e2e.sh" -- \
      "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-catalog-group-cli.mjs" \
        --manifest "$MANIFEST" --stage visual --group "$group_name" || group_status=$?
  if [[ "$group_status" -ne 0 && "$status" -eq 0 ]]; then
    status="$group_status"
  fi
done < <(
  "$node_bin" "$ROOT_DIR/tools/harness/browser/browser-batch-manifest.mjs" stage-runner "$MANIFEST" visual |
    tail -n +3
)

exit "$status"
