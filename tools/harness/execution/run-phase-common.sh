#!/usr/bin/env bash

RUN_PHASE_COMMON_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_PHASE_REPO_ROOT="$(cd "${RUN_PHASE_COMMON_DIR}/../../.." && pwd)"
TEST_OUTPUT_HELPER="${RUN_PHASE_REPO_ROOT}/tools/harness/output/test-output.sh"

phase_now_utc() {
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ 2>/dev/null || true)"
  if [[ "${timestamp}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$ ]]; then
    printf '%s\n' "${timestamp}"
    return
  fi

  if [[ -n "${NODE_BIN:-}" && -x "${NODE_BIN}" ]]; then
    "${NODE_BIN}" -e 'process.stdout.write(new Date().toISOString() + "\n")'
    return
  fi

  if [[ -x "${RUN_PHASE_REPO_ROOT}/tmp/node-runtime/bin/node" ]]; then
    "${RUN_PHASE_REPO_ROOT}/tmp/node-runtime/bin/node" -e 'process.stdout.write(new Date().toISOString() + "\n")'
    return
  fi

  if command -v node >/dev/null 2>&1; then
    node -e 'process.stdout.write(new Date().toISOString() + "\n")'
    return
  fi

  printf '%s.000Z\n' "$(date -u +%Y-%m-%dT%H:%M:%S)"
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

emit_target_timing_span() {
  if [[ "$#" -ne 7 ]]; then
    echo "emit_target_timing_span requires <bucket> <label> <start-time> <end-time> <duration-ms> <status> <exit-status>" >&2
    return 2
  fi

  if [[ -z "${CARTULARY_TEST_TARGET:-}" ]]; then
    return 0
  fi

  local bucket="$1"
  local label="$2"
  local start_time="$3"
  local end_time="$4"
  local duration_ms
  local status="$6"

  duration_ms="$(phase_clamp_duration_ms "$5")"

  CARTULARY_TIMING_BUCKET="$bucket" \
  CARTULARY_TIMING_LABEL="$label" \
  CARTULARY_TIMING_START_TIME="$start_time" \
  CARTULARY_TIMING_END_TIME="$end_time" \
  CARTULARY_TIMING_DURATION_MS="$duration_ms" \
  CARTULARY_TIMING_STATUS="$status" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" timing-span >/dev/null 2>&1 || true
}

run_timing_span() {
  if [[ "$#" -lt 3 ]]; then
    echo "run_timing_span requires <bucket> <label> <command...>" >&2
    return 2
  fi

  local bucket="$1"
  local label="$2"
  shift 2

  local start_time
  local end_time
  local start_ms
  local end_ms
  local duration_ms
  local status
  local span_status

  start_time="$(phase_now_utc)"
  start_ms="$(phase_now_monotonic_ms)"

  set +e
  "$@"
  status=$?
  set -e

  end_time="$(phase_now_utc)"
  end_ms="$(phase_now_monotonic_ms)"
  duration_ms="$(phase_elapsed_ms "${start_ms}" "${end_ms}")"
  span_status="pass"
  if [[ "${status}" -ne 0 ]]; then
    span_status="fail"
  fi
  emit_target_timing_span "$bucket" "$label" "$start_time" "$end_time" "$duration_ms" "$span_status" "$status"
  return "$status"
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

render_command_argv_json() {
  if [[ -n "${NODE_BIN:-}" && -x "${NODE_BIN:-}" ]]; then
    "${NODE_BIN}" -e 'process.stdout.write(JSON.stringify(process.argv.slice(1)))' "$@"
    return
  fi
  if command -v node >/dev/null 2>&1; then
    node -e 'process.stdout.write(JSON.stringify(process.argv.slice(1)))' "$@"
    return
  fi
  printf '[]'
}

resolve_output_mode() {
  local output_mode="${CARTULARY_OUTPUT_MODE:-}"
  if [[ -z "${output_mode}" ]]; then
    if [[ "${VERBOSE:-0}" == "1" ]]; then
      output_mode="verbose"
    elif [[ "${CI_VERBOSE:-0}" == "1" || "${CI:-0}" == "1" || "${CARTULARY_TEST_TARGET:-}" == "ci" ]]; then
      output_mode="ci"
    else
      output_mode="summary"
    fi
  fi
  case "$output_mode" in
    quiet | summary | ci | machine)
      output_mode="quiet"
      ;;
    verbose | debug)
      output_mode="normal"
      ;;
    *)
      echo "invalid CARTULARY_OUTPUT_MODE ${output_mode}; expected quiet, summary, ci, verbose, debug, or machine" >&2
      return 2
      ;;
  esac
  printf '%s\n' "$output_mode"
}

json_escape_string() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/\\n}"
  value="${value//$'\r'/\\r}"
  value="${value//$'\t'/\\t}"
  printf '%s' "$value"
}

write_vitest_watchdog_report() {
  if [[ "$#" -ne 8 ]]; then
    echo "write_vitest_watchdog_report requires <file> <label> <timeout-seconds> <grace-seconds> <pid> <started-at> <timed-out-at> <killed-at>" >&2
    return 2
  fi

  local file="$1"
  local label="$2"
  local timeout_seconds="$3"
  local grace_seconds="$4"
  local pid="$5"
  local started_at="$6"
  local timed_out_at="$7"
  local killed_at="$8"
  phase_secure_mkdir "$(dirname "$file")"
  cat >"$file" <<JSON
{
  "schema_id": "cartulary.vitest_watchdog.v1",
  "status": "timed_out",
  "label": "$(json_escape_string "$label")",
  "timeout_seconds": ${timeout_seconds},
  "kill_grace_seconds": ${grace_seconds},
  "pid": ${pid},
  "started_at": "$(json_escape_string "$started_at")",
  "timed_out_at": "$(json_escape_string "$timed_out_at")",
  "killed_at": "$(json_escape_string "$killed_at")"
}
JSON
  chmod 600 "$file" 2>/dev/null || true
}

run_vitest_command_with_watchdog() {
  if [[ "$#" -lt 6 ]]; then
    echo "run_vitest_command_with_watchdog requires <label> <phase-dir> <stdout-log> <stderr-log> <output-mode> <command...>" >&2
    return 2
  fi

  local label="$1"
  local phase_dir="$2"
  local stdout_log="$3"
  local stderr_log="$4"
  local output_mode="$5"
  local status
  shift 5

  local timeout_seconds="${CARTULARY_VITEST_WATCHDOG_SECONDS:-300}"
  local grace_seconds="${CARTULARY_WATCHDOG_KILL_GRACE_SECONDS:-10}"
  if ! [[ "$timeout_seconds" =~ ^[0-9]+$ ]]; then
    echo "invalid CARTULARY_VITEST_WATCHDOG_SECONDS=${timeout_seconds}" >&2
    return 2
  fi
  if ! [[ "$grace_seconds" =~ ^[0-9]+$ ]]; then
    echo "invalid CARTULARY_WATCHDOG_KILL_GRACE_SECONDS=${grace_seconds}" >&2
    return 2
  fi

  phase_secure_mkdir "$(dirname "$stdout_log")" "$(dirname "$stderr_log")" "$phase_dir"
  CARTULARY_VITEST_WATCHDOG_LOG="${phase_dir}/watchdog.json"
  export CARTULARY_VITEST_WATCHDOG_LOG
  RUN_VITEST_WATCHDOG_TIMED_OUT=0
  export RUN_VITEST_WATCHDOG_TIMED_OUT

  if [[ "$timeout_seconds" -eq 0 ]]; then
    if [[ "$output_mode" != "quiet" ]]; then
      "$@" > >(phase_redact_stream | tee "$stdout_log") 2> >(phase_redact_stream | tee "$stderr_log" >&2)
      status=$?
    else
      "$@" >"$stdout_log" 2>"$stderr_log"
      status=$?
    fi
    phase_redact_file "$stdout_log"
    phase_redact_file "$stderr_log"
    return "$status"
  fi

  local started_at
  local start_ms
  started_at="$(phase_now_utc)"
  start_ms="$(phase_now_monotonic_ms)"

  if command -v setsid >/dev/null 2>&1; then
    # shellcheck disable=SC2016
    setsid bash -c '
      set -euo pipefail
      output_mode="$1"
      stdout_log="$2"
      stderr_log="$3"
      node_bin="$4"
      contract_script="$5"
      shift 5
      if [[ "$output_mode" != "quiet" ]]; then
        exec "$@" > >("$node_bin" "$contract_script" redact | tee "$stdout_log") 2> >("$node_bin" "$contract_script" redact | tee "$stderr_log" >&2)
      fi
      exec "$@" >"$stdout_log" 2>"$stderr_log"
    ' bash "$output_mode" "$stdout_log" "$stderr_log" "$(resolve_harness_node)" "${RUN_PHASE_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" "$@" &
  else
    bash -c '
      set -euo pipefail
      output_mode="$1"
      stdout_log="$2"
      stderr_log="$3"
      node_bin="$4"
      contract_script="$5"
      shift 5
      if [[ "$output_mode" != "quiet" ]]; then
        exec "$@" > >("$node_bin" "$contract_script" redact | tee "$stdout_log") 2> >("$node_bin" "$contract_script" redact | tee "$stderr_log" >&2)
      fi
      exec "$@" >"$stdout_log" 2>"$stderr_log"
    ' bash "$output_mode" "$stdout_log" "$stderr_log" "$(resolve_harness_node)" "${RUN_PHASE_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" "$@" &
  fi
  local child_pid=$!
  local child_group="-$child_pid"
  local kill_target="$child_pid"
  if command -v setsid >/dev/null 2>&1; then
    kill_target="$child_group"
  fi

  while kill -0 "$child_pid" >/dev/null 2>&1; do
    local now_ms
    now_ms="$(phase_now_monotonic_ms)"
    if (( now_ms - start_ms >= timeout_seconds * 1000 )); then
      local timed_out_at
      local killed_at
      timed_out_at="$(phase_now_utc)"
      kill -TERM "$kill_target" >/dev/null 2>&1 || true
      local grace_start_ms
      grace_start_ms="$(phase_now_monotonic_ms)"
      while kill -0 "$child_pid" >/dev/null 2>&1; do
        local grace_now_ms
        grace_now_ms="$(phase_now_monotonic_ms)"
        if (( grace_now_ms - grace_start_ms >= grace_seconds * 1000 )); then
          kill -KILL "$kill_target" >/dev/null 2>&1 || true
          break
        fi
        sleep 0.2
      done
      wait "$child_pid" >/dev/null 2>&1 || true
      killed_at="$(phase_now_utc)"
      write_vitest_watchdog_report "$CARTULARY_VITEST_WATCHDOG_LOG" "$label" "$timeout_seconds" "$grace_seconds" "$child_pid" "$started_at" "$timed_out_at" "$killed_at"
      RUN_VITEST_WATCHDOG_TIMED_OUT=1
      export RUN_VITEST_WATCHDOG_TIMED_OUT
      phase_redact_file "$stdout_log"
      phase_redact_file "$stderr_log"
      chmod 600 "$CARTULARY_VITEST_WATCHDOG_LOG" 2>/dev/null || true
      return 124
    fi
    sleep 0.5
  done

  wait "$child_pid"
  status=$?
  phase_redact_file "$stdout_log"
  phase_redact_file "$stderr_log"
  return "$status"
}

slugify_phase_label() {
  local label="$1"
  printf '%s' "$label" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/--+/-/g'
}

resolve_harness_node() {
  if [[ -n "${NODE_BIN:-}" && -x "${NODE_BIN}" ]]; then
    printf '%s\n' "${NODE_BIN}"
    return
  fi
  if [[ -x "${RUN_PHASE_REPO_ROOT}/tmp/node-runtime/bin/node" ]]; then
    printf '%s\n' "${RUN_PHASE_REPO_ROOT}/tmp/node-runtime/bin/node"
    return
  fi
  printf '%s\n' "node"
}

phase_secure_mkdir() {
  if [[ "$#" -lt 1 ]]; then
    echo "phase_secure_mkdir requires <dir...>" >&2
    return 2
  fi
  local dir
  for dir in "$@"; do
    mkdir -p "$dir"
    chmod 700 "$dir" 2>/dev/null || true
  done
}

phase_redact_stream() {
  local node_bin
  node_bin="$(resolve_harness_node)"
  "${node_bin}" "${RUN_PHASE_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" redact
}

phase_redact_file() {
  if [[ "$#" -ne 1 ]]; then
    echo "phase_redact_file requires <file>" >&2
    return 2
  fi
  local file="$1"
  if [[ ! -e "$file" ]]; then
    return 0
  fi
  local tmp_file
  tmp_file="${file}.redacted.$$"
  phase_redact_stream <"$file" >"$tmp_file"
  mv "$tmp_file" "$file"
  chmod 600 "$file" 2>/dev/null || true
}

ensure_harness_artifact_identity() {
  if [[ -n "${CARTULARY_HARNESS_IDENTITY_PREPARED:-}" ]]; then
    return 0
  fi

  local target
  target="$(resolve_test_target)"
  if [[ "${target}" == "adhoc" ]]; then
    CARTULARY_HARNESS_IDENTITY_PREPARED=1
    export CARTULARY_HARNESS_IDENTITY_PREPARED
    return 0
  fi

  local node_bin
  local output
  node_bin="$(resolve_harness_node)"
  output="$(CARTULARY_SUPPRESS_CHILD_SUCCESS=1 "${node_bin}" "${RUN_PHASE_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" retained-artifact-env "${target}")" || return "$?"
  CARTULARY_TEST_RESULTS_DIR="$(printf '%s\n' "${output}" | sed -n '1p')"
  CARTULARY_TEST_RUN_ID="$(printf '%s\n' "${output}" | sed -n '2p')"
  CARTULARY_HARNESS_IDENTITY_PREPARED=1
  export CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID CARTULARY_HARNESS_IDENTITY_PREPARED
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
  ensure_harness_artifact_identity
  results_root="$(resolve_results_root)"
  run_id="$(resolve_test_run_id)"
  target="$(resolve_test_target)"
  phase_secure_mkdir "${results_root}/${run_id}" "${results_root}/${run_id}/${target}"
  printf '%s\n' "${results_root}/${run_id}/${target}"
}

prepare_target_support_dir() {
  local name="${1:-support}"
  local target_dir
  target_dir="$(ensure_target_artifact_dir)"
  phase_secure_mkdir "${target_dir}/${name}"
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

  ensure_harness_artifact_identity
  results_root="$(resolve_results_root)"
  run_id="$(resolve_test_run_id)"
  phase_secure_mkdir "${results_root}/${run_id}" "${results_root}/${run_id}/_shared" "${results_root}/${run_id}/_shared/${name}"
  printf '%s\n' "${results_root}/${run_id}/_shared/${name}"
}

prepare_phase_artifact_dir() {
  local phase="$1"
  local target_dir
  local slug
  target_dir="$(ensure_target_artifact_dir)"
  slug="$(slugify_phase_label "$phase")"
  phase_secure_mkdir "${target_dir}/${slug}"
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
  CARTULARY_PHASE_TIMING_BUCKET="${CARTULARY_PHASE_TIMING_BUCKET:-}" \
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
  local command_argv_json
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
  command_argv_json="$(render_command_argv_json "$@")"

  if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
    echo "== ${phase} =="
  fi

  phase_capture_start PHASE

  set +e
  if [[ "$output_mode" != "quiet" ]]; then
    CARTULARY_PHASE_ARTIFACT_DIR="$phase_dir" "$@" > >(phase_redact_stream | tee "$stdout_log") 2> >(phase_redact_stream | tee "$stderr_log" >&2)
    status=$?
  else
    CARTULARY_PHASE_ARTIFACT_DIR="$phase_dir" "$@" >"$stdout_log" 2>"$stderr_log"
    status=$?
  fi
  set -e
  phase_redact_file "$stdout_log"
  phase_redact_file "$stderr_log"

  phase_capture_finish PHASE
  start_time="${PHASE_START_TIME}"
  end_time="${PHASE_END_TIME}"
  duration_ms="${PHASE_DURATION_MS}"

  if [[ "$status" -eq 0 && "$output_mode" == "quiet" && "${CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG:-0}" == "1" && "${CARTULARY_ENABLE_LEGACY_SUCCESS_LOG:-0}" == "1" && "${CARTULARY_SUPPRESS_CHILD_SUCCESS:-0}" != "1" ]]; then
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
  CARTULARY_PHASE_COMMAND_ARGV="$command_argv_json" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$duration_ms" \
  CARTULARY_PHASE_WALL_DURATION_MS="${PHASE_WALL_DURATION_MS}" \
  CARTULARY_PHASE_EXIT_STATUS="$status" \
  CARTULARY_PHASE_TIMING_BUCKET="${CARTULARY_PHASE_TIMING_BUCKET:-}" \
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
