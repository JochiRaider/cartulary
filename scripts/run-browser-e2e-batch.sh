#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"
TEST_OUTPUT_HELPER="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/scripts/lib/test-output.mjs}"

usage() {
  echo "usage: run-browser-e2e-batch.sh <stage> [--defer-summary]" >&2
  exit 2
}

if [[ "$#" -lt 1 || "$#" -gt 2 ]]; then
  usage
fi

stage="$1"
shift
defer_summary=0
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --defer-summary)
      defer_summary=1
      ;;
    *)
      usage
      ;;
  esac
  shift
done

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

resolve_playwright_owned_stack_env "$ROOT_DIR"

stage_metadata="$("$node_bin" "$ROOT_DIR/scripts/lib/browser-batch-manifest.mjs" stage-runner "$MANIFEST" "$stage")"
stage_target="$(printf '%s\n' "$stage_metadata" | sed -n '1p')"
stage_summary_children="$(printf '%s\n' "$stage_metadata" | sed -n '2p')"
mapfile -t stage_groups < <(printf '%s\n' "$stage_metadata" | tail -n +3)

run_target_summary() {
  local target="$1"
  local status="$2"
  local children="${3:-}"

  if [[ -n "$children" ]]; then
    emit_test_output target-summary "$target" "$status" --children "$children"
    return $?
  fi

  emit_test_output target-summary "$target" "$status"
}

run_group() {
  local target="$1"
  local kind="$2"
  local workers="$3"
  local coverage="$4"
  local execution_dependency="$5"
  shift 5

  local -a group_env=(
    env
    "CARTULARY_TEST_TARGET=$target"
    "NODE_BIN=$PLAYWRIGHT_OWNED_STACK_NODE_BIN"
  )

  if [[ "$workers" != "default" ]]; then
    group_env+=("PLAYWRIGHT_WORKERS=$workers")
  fi

  case "$kind" in
    webserver-backed)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-webserver-backed.sh"
      ;;
    duration_balanced_specs)
      if [[ "$target" == "browser-e2e-webserver-backed" ]]; then
        "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-webserver-backed.sh"
      else
        "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-functional.sh"
      fi
      ;;
    functional)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-functional.sh"
      ;;
    support)
      local -a support_env=(
        "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"
        "CARTULARY_TEST_TARGET=$target"
        "NODE_BIN=$PLAYWRIGHT_OWNED_STACK_NODE_BIN"
      )
      if [[ "$workers" != "default" ]]; then
        support_env+=("PLAYWRIGHT_WORKERS=$workers")
      fi
      "${support_env[@]}" "$ROOT_DIR/scripts/lib/run-playwright-webserver-batch.sh" \
        support \
        -- \
        "$PLAYWRIGHT_OWNED_STACK_PNPM_BIN" --dir apps/web exec playwright test \
        --config playwright.webserver-backed.config.ts
      ;;
    stateful)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-stateful.sh"
      ;;
    measurement)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-measurement.sh"
      ;;
    visual)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-visual.sh"
      ;;
    a11y)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-a11y.sh"
      ;;
    a11y_preflight)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-a11y-preflight.sh"
      ;;
    *)
      echo "unsupported browser E2E batch group kind ${kind}" >&2
      return 2
      ;;
  esac
}

if [[ -n "${CARTULARY_TEST_TARGET:-}" && "${CARTULARY_TEST_TARGET}" == "$stage_target" ]]; then
  if [[ -n "$stage_summary_children" ]]; then
    emit_test_output target-start "$stage_target" --children "$stage_summary_children" || true
  else
    emit_test_output target-start "$stage_target" || true
  fi
fi

overall_status=0
declare -A child_target_status=()
declare -A summary_child_targets=()
if [[ -n "$stage_summary_children" ]]; then
  IFS=',' read -r -a summary_children_array <<<"$stage_summary_children"
  for child_target in "${summary_children_array[@]}"; do
    summary_child_targets["$child_target"]=1
  done
else
  summary_children_array=()
fi

for group_row in "${stage_groups[@]}"; do
  group_row_fields="${group_row//$'\t'/$'\x1f'}"
  IFS=$'\x1f' read -r _group_name target kind workers reset_before coverage execution_dependency _stage_schedule_tags _stage_scheduler_needs <<<"$group_row_fields"

  if [[ -n "$reset_before" ]]; then
    env CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-$stage_target}" \
      NODE_BIN="$node_bin" \
      "$ROOT_DIR/scripts/reset-web-e2e-stack.sh" --label "$reset_before"
  fi

  set +e
  run_group "$target" "$kind" "$workers" "$coverage" "$execution_dependency"
  group_status=$?
  set -e

  if [[ -n "${summary_child_targets[$target]:-}" ]]; then
    if [[ -z "${child_target_status[$target]:-}" || "${child_target_status[$target]}" -eq 0 ]]; then
      child_target_status["$target"]="$group_status"
    fi
  fi

  if [[ "$group_status" -ne 0 && "$overall_status" -eq 0 ]]; then
    overall_status="$group_status"
  fi
done

for child_target in "${summary_children_array[@]}"; do
  child_status="${child_target_status[$child_target]:-0}"
  if [[ "$child_status" -eq 0 ]]; then
    run_target_summary "$child_target" pass || child_status=$?
  else
    run_target_summary "$child_target" fail || true
  fi
  if [[ "$child_status" -ne 0 && "$overall_status" -eq 0 ]]; then
    overall_status="$child_status"
  fi
done

if [[ "$defer_summary" -ne 1 && -n "${CARTULARY_TEST_TARGET:-}" && "${CARTULARY_TEST_TARGET}" == "$stage_target" ]]; then
  if [[ "$overall_status" -eq 0 ]]; then
    run_target_summary "$stage_target" pass "$stage_summary_children" || overall_status=$?
  else
    run_target_summary "$stage_target" fail "$stage_summary_children" || true
  fi
fi

exit "$overall_status"
