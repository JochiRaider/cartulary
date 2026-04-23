#!/usr/bin/env bash

RUN_PHASE_COMMON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_PHASE_REPO_ROOT="$(cd "${RUN_PHASE_COMMON_DIR}/../.." && pwd)"
TEST_OUTPUT_HELPER="${RUN_PHASE_COMMON_DIR}/test-output.sh"

phase_now_utc() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

phase_now_monotonic_ms() {
  if [[ -r /proc/uptime ]]; then
    LC_ALL=C awk '{printf "%.0f\n", $1 * 1000}' /proc/uptime
    return
  fi

  local now_ms
  now_ms="$(date +%s%3N 2>/dev/null || true)"
  if [[ "${now_ms}" =~ ^[0-9]+$ ]]; then
    printf '%s\n' "${now_ms}"
    return
  fi

  printf '%s000\n' "$(date +%s)"
}

phase_clamp_duration_ms() {
  local value="${1:-0}"
  if [[ ! "${value}" =~ ^-?[0-9]+$ ]] || (( value < 0 )); then
    printf '0\n'
    return
  fi
  printf '%s\n' "${value}"
}

phase_elapsed_ms() {
  local start_ms="${1:-0}"
  local end_ms="${2:-0}"
  if [[ ! "${start_ms}" =~ ^-?[0-9]+$ ]] || [[ ! "${end_ms}" =~ ^-?[0-9]+$ ]]; then
    printf '0\n'
    return
  fi
  phase_clamp_duration_ms "$((end_ms - start_ms))"
}

phase_capture_start() {
  if [[ "$#" -ne 1 ]]; then
    echo "phase_capture_start requires <prefix>" >&2
    return 2
  fi

  local prefix="$1"
  printf -v "${prefix}_START_TIME" '%s' "$(phase_now_utc)"
  printf -v "${prefix}_START_MONOTONIC_MS" '%s' "$(phase_now_monotonic_ms)"
}

phase_capture_finish() {
  if [[ "$#" -ne 1 ]]; then
    echo "phase_capture_finish requires <prefix>" >&2
    return 2
  fi

  local prefix="$1"
  local start_monotonic_var="${prefix}_START_MONOTONIC_MS"
  local end_time
  local end_monotonic
  local duration_ms

  end_time="$(phase_now_utc)"
  end_monotonic="$(phase_now_monotonic_ms)"
  duration_ms="$(phase_elapsed_ms "${!start_monotonic_var:-0}" "${end_monotonic}")"

  printf -v "${prefix}_END_TIME" '%s' "${end_time}"
  printf -v "${prefix}_DURATION_MS" '%s' "${duration_ms}"
  printf -v "${prefix}_WALL_DURATION_MS" '%s' "${duration_ms}"
}

stream_go_json_output() {
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" go-json-stream
}

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

prepare_target_support_dir() {
  local name="${1:-support}"
  local target_dir
  target_dir="$(ensure_target_artifact_dir)"
  mkdir -p "${target_dir}/${name}"
  printf '%s\n' "${target_dir}/${name}"
}

prepare_shared_artifact_dir() {
  local name="$1"
  local results_root
  local run_id

  if [[ -z "${name}" ]]; then
    echo "prepare_shared_artifact_dir requires <name>" >&2
    return 2
  fi

  results_root="$(resolve_results_root)"
  run_id="$(resolve_test_run_id)"
  mkdir -p "${results_root}/${run_id}/_shared/${name}"
  printf '%s\n' "${results_root}/${run_id}/_shared/${name}"
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

emit_report_phase_summary() {
  if [[ "$#" -ne 8 ]]; then
    echo "emit_report_phase_summary requires <helper-command> <label> <command-text> <start-time> <end-time> <duration-ms> <wall-duration-ms> <exit-status>" >&2
    return 2
  fi

  local helper_command="$1"
  local phase="$2"
  local command_text="$3"
  local start_time="$4"
  local end_time="$5"
  local duration_ms
  local wall_duration_ms
  local exit_status="$8"
  local phase_dir
  local helper_status

  duration_ms="$(phase_clamp_duration_ms "$6")"
  wall_duration_ms="$(phase_clamp_duration_ms "$7")"

  phase_dir="$(prepare_phase_artifact_dir "$phase")"

  set +e
  CARTULARY_PHASE_LABEL="$phase" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$duration_ms" \
  CARTULARY_PHASE_WALL_DURATION_MS="$wall_duration_ms" \
  CARTULARY_PHASE_EXIT_STATUS="$exit_status" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" "${helper_command}"
  helper_status=$?
  set -e

  if [[ "${helper_status}" -eq 0 ]]; then
    return 0
  fi

  emit_target_summary fail || true
  return "${helper_status}"
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

  phase_capture_start PHASE

  set +e
  if [[ "$output_mode" != "quiet" ]]; then
    "$@" > >(tee "$stdout_log") 2> >(tee "$stderr_log" >&2)
    status=$?
  else
    "$@" >"$stdout_log" 2>"$stderr_log"
    status=$?
  fi
  set -e

  phase_capture_finish PHASE
  start_time="${PHASE_START_TIME}"
  end_time="${PHASE_END_TIME}"
  duration_ms="${PHASE_DURATION_MS}"

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
  CARTULARY_PHASE_WALL_DURATION_MS="${PHASE_WALL_DURATION_MS}" \
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
