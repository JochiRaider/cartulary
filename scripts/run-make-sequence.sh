#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TEST_OUTPUT_SCRIPT="${TEST_OUTPUT_SCRIPT:-${ROOT_DIR}/scripts/lib/test-output.sh}"
MAKE_BIN="${MAKE:-make}"

label=""
summary_targets_csv=""
steps=()

usage() {
  echo "usage: run-make-sequence.sh --label <name> --summary-targets <a,b> [--step <target> | --parallel-step <target>:<jobs>]..." >&2
}

add_step() {
  local kind="$1"
  local spec="$2"

  if [[ -z "${spec}" ]]; then
    usage
    return 2
  fi

  case "${kind}" in
    step)
      steps+=("step:${spec}")
      ;;
    parallel)
      if [[ "${spec}" != *:* ]]; then
        echo "--parallel-step requires <target>:<jobs>, got ${spec}" >&2
        return 2
      fi
      steps+=("parallel:${spec}")
      ;;
  esac
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --label)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      label="$2"
      shift 2
      ;;
    --summary-targets)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      summary_targets_csv="$2"
      shift 2
      ;;
    --step)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      add_step step "$2"
      shift 2
      ;;
    --parallel-step)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      add_step parallel "$2"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "${label}" || -z "${summary_targets_csv}" || "${#steps[@]}" -eq 0 ]]; then
  usage
  exit 2
fi

summary_targets=()
IFS=',' read -r -a summary_targets <<<"${summary_targets_csv}"
for i in "${!summary_targets[@]}"; do
  summary_targets[$i]="${summary_targets[$i]//[[:space:]]/}"
done

dry_run=0
case " ${MAKEFLAGS:-} " in
  *" n"*|*" --just-print"*|*" --dry-run"*) dry_run=1 ;;
esac

completed=0
total="${#steps[@]}"

run_summary() {
  local status="$1"
  local aborted_after="$2"

  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" run-summary \
    "${label}" "${status}" "${completed}" "${total}" "${aborted_after}" "${summary_targets[@]}"
}

run_step() {
  local encoded="$1"
  local kind="${encoded%%:*}"
  local rest="${encoded#*:}"
  local target
  local jobs

  case "${kind}" in
    step)
      target="${rest}"
      "${MAKE_BIN}" --no-print-directory "${target}"
      ;;
    parallel)
      target="${rest%%:*}"
      jobs="${rest#*:}"
      if [[ -z "${target}" || -z "${jobs}" || "${jobs}" == "${rest}" ]]; then
        echo "invalid parallel step ${rest}; expected <target>:<jobs>" >&2
        return 2
      fi
      "${MAKE_BIN}" --no-print-directory --output-sync=target "-j${jobs}" "${target}"
      ;;
    *)
      echo "unknown step kind ${kind}" >&2
      return 2
      ;;
  esac
}

for step in "${steps[@]}"; do
  step_target="${step#*:}"
  step_target="${step_target%%:*}"

  set +e
  run_step "${step}"
  status=$?
  set -e

  if [[ "${status}" -eq 0 ]]; then
    completed=$((completed + 1))
    continue
  fi

  run_summary fail "${step_target}" || true
  exit "${status}"
done

run_summary pass -
