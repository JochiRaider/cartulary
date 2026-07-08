#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"

resolve_playwright_owned_stack_env "$ROOT_DIR"
status=0

selected_row_ids="${CARTULARY_BROWSER_SELECTED_ROW_IDS:-}"
selected_phase="${CARTULARY_BROWSER_SELECTED_PHASE:-${CARTULARY_PHASE_SLICE_PHASE:-}}"
frontend_row_ids=()

if [[ -n "$selected_row_ids" ]]; then
  IFS=',' read -r -a selected_row_id_array <<<"$selected_row_ids"
  for raw_row_id in "${selected_row_id_array[@]}"; do
    row_id="${raw_row_id//[[:space:]]/}"
    if [[ -z "$row_id" ]]; then
      continue
    fi
    if [[ "$row_id" == FE-* ]]; then
      frontend_row_ids+=("$row_id")
    fi
  done
fi

run_base_manifest=1
if [[ -n "$selected_row_ids" && -n "$selected_phase" ]]; then
  phase_stateful_count="$(
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/phase-accounting/phase-manifest.mjs" \
        playwright-count "$selected_phase" authoritative browser_stateful
  )"
  if [[ "$phase_stateful_count" == "0" ]]; then
    run_base_manifest=0
  fi
fi

if [[ "$run_base_manifest" -eq 1 ]]; then
  base_env=(env)
  if [[ -n "$selected_phase" ]]; then
    base_env+=("CARTULARY_PHASE_SLICE_PHASE=$selected_phase")
  fi
  "${base_env[@]}" "$ROOT_DIR/tools/harness/browser/run-browser-e2e-manifest-dependency.sh" \
    browser-e2e-stateful authoritative browser_stateful -- \
    "$ROOT_DIR/tools/harness/browser/run-browser-e2e-owned-stack.sh" "$@" ||
    status=1
fi

frontend_grep=""
frontend_scope="${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}"
if [[ "$frontend_scope" != "disabled" ]]; then
  frontend_grep_args=(playwright-grep browser-e2e-stateful e2e)
  if [[ -n "$selected_row_ids" ]]; then
    if [[ "${#frontend_row_ids[@]}" -eq 0 ]]; then
      frontend_scope="disabled"
    else
      frontend_row_ids_csv="$(
        IFS=,
        printf '%s' "${frontend_row_ids[*]}"
      )"
      frontend_grep_args+=(--row-ids "$frontend_row_ids_csv")
      export CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE=selected_rows
      export CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS="$frontend_row_ids_csv"
    fi
  elif [[ "$frontend_scope" == "selected_rows" ]]; then
    frontend_grep_args+=(--row-ids "${CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:-}")
  fi
fi

if [[ "$frontend_scope" != "disabled" ]]; then
  if [[ -n "$selected_phase" ]]; then
    frontend_accounting_phase="$selected_phase"
    if [[ "$selected_phase" =~ ^phase([0-9]+)$ ]]; then
      frontend_accounting_phase="FE-P${BASH_REMATCH[1]}"
    fi
    export CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE=frontend
    export CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE="$frontend_accounting_phase"
  fi
  frontend_grep="$(
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" "$ROOT_DIR/tools/harness/phase-accounting/frontend-phase-manifest.mjs" \
        "${frontend_grep_args[@]}"
  )"
fi

if [[ -n "$frontend_grep" ]]; then
  frontend_phase_label="browser-e2e-stateful frontend-readiness"
  if [[ -n "$selected_phase" ]]; then
    frontend_phase_label="${frontend_phase_label} ${selected_phase}"
  fi
  "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}" \
    NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
    "$ROOT_DIR/tools/harness/browser/run-playwright-phase.sh" \
    "$frontend_phase_label" -- \
    "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" --dir apps/web exec playwright test \
    apps/web/e2e/*.spec.ts -g "$frontend_grep" ||
    status=1
fi

exit "$status"
