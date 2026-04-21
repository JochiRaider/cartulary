#!/usr/bin/env bash

RUN_PHASE_COMMON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_PHASE_REPO_ROOT="$(cd "${RUN_PHASE_COMMON_DIR}/../.." && pwd)"
TEST_OUTPUT_HELPER="${RUN_PHASE_COMMON_DIR}/test-output.sh"

render_command() {
  local rendered=""
  local arg
  for arg in "$@"; do
    printf -v rendered '%s%q ' "$rendered" "$arg"
  done
  printf '%s' "${rendered% }"
}

resolve_output_mode() {
  local output_mode="${CARTULARY_OUTPUT_MODE:-quiet}"
  if [[ "${VERBOSE:-0}" == "1" || "${CI_VERBOSE:-0}" == "1" ]]; then
    output_mode="normal"
  fi
  printf '%s\n' "$output_mode"
}

slugify_phase_label() {
  local label="$1"
  printf '%s' "$label" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/--+/-/g'
}

resolve_results_root() {
  if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" ]]; then
    if [[ "${CARTULARY_TEST_RESULTS_DIR}" = /* ]]; then
      printf '%s\n' "${CARTULARY_TEST_RESULTS_DIR}"
    else
      printf '%s\n' "${RUN_PHASE_REPO_ROOT}/${CARTULARY_TEST_RESULTS_DIR}"
    fi
    return
  fi
  printf '%s\n' "${RUN_PHASE_REPO_ROOT}/.cartulary/test-results"
}

resolve_test_run_id() {
  if [[ -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    CARTULARY_TEST_RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-p$$"
    export CARTULARY_TEST_RUN_ID
  fi
  printf '%s\n' "${CARTULARY_TEST_RUN_ID}"
}

resolve_test_target() {
  if [[ -n "${CARTULARY_TEST_TARGET:-}" ]]; then
    printf '%s\n' "${CARTULARY_TEST_TARGET}"
    return
  fi
  printf '%s\n' "adhoc"
}

ensure_target_artifact_dir() {
  local results_root
  local run_id
  local target
  results_root="$(resolve_results_root)"
  run_id="$(resolve_test_run_id)"
  target="$(resolve_test_target)"
  mkdir -p "${results_root}/${run_id}/${target}"
  printf '%s\n' "${results_root}/${run_id}/${target}"
}

prepare_phase_artifact_dir() {
  local phase="$1"
  local target_dir
  local slug
  target_dir="$(ensure_target_artifact_dir)"
  slug="$(slugify_phase_label "$phase")"
  mkdir -p "${target_dir}/${slug}"
  printf '%s\n' "${target_dir}/${slug}"
}

emit_target_summary() {
  local status="${1:-pass}"
  if [[ -z "${CARTULARY_TEST_TARGET:-}" ]]; then
    return 0
  fi
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" target-summary "${CARTULARY_TEST_TARGET}" "${status}"
}

run_phase_command() {
  if [[ "$#" -lt 2 ]]; then
    echo "run_phase_command requires <label> <command...>" >&2
    return 2
  fi

  local phase="$1"
  shift

  local output_mode
  local phase_dir
  local stdout_log
  local stderr_log
  local command_text
  local start_time
  local end_time
  local start_ms
  local end_ms
  local duration_ms
  local status
  local helper_status

  output_mode="$(resolve_output_mode)"
  phase_dir="$(prepare_phase_artifact_dir "$phase")"
  stdout_log="${phase_dir}/stdout.log"
  stderr_log="${phase_dir}/stderr.log"
  command_text="$(render_command "$@")"

  if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
    echo "== ${phase} =="
  fi

  start_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  start_ms="$(date +%s%3N)"

  set +e
  if [[ "$output_mode" != "quiet" ]]; then
    "$@" > >(tee "$stdout_log") 2> >(tee "$stderr_log" >&2)
    status=$?
  else
    "$@" >"$stdout_log" 2>"$stderr_log"
    status=$?
  fi
  set -e

  end_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  end_ms="$(date +%s%3N)"
  duration_ms="$((end_ms - start_ms))"

  if [[ "$status" -eq 0 && "$output_mode" == "quiet" && "${CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG:-0}" == "1" ]]; then
    if [[ -s "$stdout_log" ]]; then
      cat "$stdout_log"
    fi
    if [[ -s "$stderr_log" ]]; then
      cat "$stderr_log" >&2
    fi
  fi

  set +e
  CARTULARY_PHASE_LABEL="$phase" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$duration_ms" \
  CARTULARY_PHASE_EXIT_STATUS="$status" \
  CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" shell-phase
  helper_status=$?
  set -e

  if [[ "$status" -eq 0 && "$helper_status" -eq 0 ]]; then
    return 0
  fi

  emit_target_summary fail || true

  if [[ "$status" -ne 0 ]]; then
    return "$status"
  fi
  return "$helper_status"
}
