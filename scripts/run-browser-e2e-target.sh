#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"
TEST_OUTPUT_HELPER="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/scripts/lib/test-output.mjs}"

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

stage_metadata="$("$node_bin" "$ROOT_DIR/scripts/lib/browser-batch-manifest.mjs" stage-target "$MANIFEST" "$stage")"
target="$(printf '%s\n' "$stage_metadata" | sed -n '1p')"
summary_children="$(printf '%s\n' "$stage_metadata" | sed -n '2p')"

status=0
env CARTULARY_TEST_TARGET="$target" \
  NODE_BIN="$node_bin" \
  "$ROOT_DIR/scripts/start-web-e2e.sh" -- "$ROOT_DIR/scripts/run-browser-e2e-batch.sh" "$stage" --defer-summary || status=$?

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
