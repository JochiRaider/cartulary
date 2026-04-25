#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TEST_OUTPUT_SCRIPT="${TEST_OUTPUT_SCRIPT:-${ROOT_DIR}/scripts/lib/test-output.sh}"
MAKE_BIN="${MAKE:-make}"

label=""
summary_targets_csv=""
summary_groups_spec=""
steps=()

usage() {
  echo "usage: run-make-sequence.sh --label <name> --summary-targets <a,b> [--summary-groups <name=a,b;name=c>] [--step <target> | --parallel-step <target>:<jobs>]..." >&2
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
    --summary-groups)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      summary_groups_spec="$2"
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
target_count="${#summary_targets[@]}"
max_jobs=1

for step in "${steps[@]}"; do
  step_kind="${step%%:*}"
  step_rest="${step#*:}"
  if [[ "${step_kind}" != "parallel" ]]; then
    continue
  fi
  step_jobs="${step_rest#*:}"
  if [[ "${step_jobs}" =~ ^[0-9]+$ ]] && (( step_jobs > max_jobs )); then
    max_jobs="${step_jobs}"
  fi
done

emit_run_start() {
  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" run-start \
    "${label}" --steps "${total}" --targets "${target_count}" --jobs "${max_jobs}"
}

emit_step_start() {
  local encoded="$1"
  local index="$2"
  local kind="${encoded%%:*}"
  local rest="${encoded#*:}"
  local target
  local jobs=1
  local mode=serial

  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  case "${kind}" in
    step)
      target="${rest}"
      ;;
    parallel)
      target="${rest%%:*}"
      jobs="${rest#*:}"
      mode=parallel
      ;;
    *)
      target="${rest}"
      ;;
  esac

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" step-start \
    "${label}" "${index}" "${total}" "${target}" --mode "${mode}" --jobs "${jobs}"
}

run_summary() {
  local status="$1"
  local aborted_after="$2"

  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  local summary_args=()
  if [[ -n "${summary_groups_spec}" ]]; then
    summary_args+=(--summary-groups "${summary_groups_spec}")
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" run-summary \
    "${label}" "${status}" "${completed}" "${total}" "${aborted_after}" \
    "${summary_args[@]}" "${summary_targets[@]}"
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

emit_run_start

for index in "${!steps[@]}"; do
  step="${steps[$index]}"
  step_target="${step#*:}"
  step_target="${step_target%%:*}"
  step_number=$((index + 1))

  emit_step_start "${step}" "${step_number}"

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
