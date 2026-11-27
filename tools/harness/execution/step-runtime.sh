#!/usr/bin/env bash
# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"

STEP_RUNTIME_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_STEP_REPO_ROOT="$(cd "${STEP_RUNTIME_DIR}/../../.." && pwd)"
TEST_OUTPUT_HELPER="${RUN_STEP_REPO_ROOT}/tools/harness/output/test-output.sh"

step_now_utc() {
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

  if [[ -x "${RUN_STEP_REPO_ROOT}/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node" ]]; then
    "${RUN_STEP_REPO_ROOT}/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node" -e 'process.stdout.write(new Date().toISOString() + "\n")'
    return
  fi

  if command -v node >/dev/null 2>&1; then
    node -e 'process.stdout.write(new Date().toISOString() + "\n")'
    return
  fi

  printf '%s.000Z\n' "$(date -u +%Y-%m-%dT%H:%M:%S)"
}

step_now_monotonic_ms() {
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

step_clamp_duration_ms() {
  local value="${1:-0}"
  if [[ ! "${value}" =~ ^-?[0-9]+$ ]] || (( value < 0 )); then
    printf '0\n'
    return
  fi
  printf '%s\n' "${value}"
}

step_elapsed_ms() {
  local start_ms="${1:-0}"
  local end_ms="${2:-0}"
  if [[ ! "${start_ms}" =~ ^-?[0-9]+$ ]] || [[ ! "${end_ms}" =~ ^-?[0-9]+$ ]]; then
    printf '0\n'
    return
  fi
  step_clamp_duration_ms "$((end_ms - start_ms))"
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

  duration_ms="$(step_clamp_duration_ms "$5")"

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

  start_time="$(step_now_utc)"
  start_ms="$(step_now_monotonic_ms)"

  set +e
  "$@"
  status=$?
  set -e

  end_time="$(step_now_utc)"
  end_ms="$(step_now_monotonic_ms)"
  duration_ms="$(step_elapsed_ms "${start_ms}" "${end_ms}")"
  span_status="pass"
  if [[ "${status}" -ne 0 ]]; then
    span_status="fail"
  fi
  emit_target_timing_span "$bucket" "$label" "$start_time" "$end_time" "$duration_ms" "$span_status" "$status"
  return "$status"
}

step_capture_start() {
  if [[ "$#" -ne 1 ]]; then
    echo "step_capture_start requires <prefix>" >&2
    return 2
  fi

  local prefix="$1"
  printf -v "${prefix}_START_TIME" '%s' "$(step_now_utc)"
  printf -v "${prefix}_START_MONOTONIC_MS" '%s' "$(step_now_monotonic_ms)"
}

step_capture_finish() {
  if [[ "$#" -ne 1 ]]; then
    echo "step_capture_finish requires <prefix>" >&2
    return 2
  fi

  local prefix="$1"
  local start_monotonic_var="${prefix}_START_MONOTONIC_MS"
  local end_time
  local end_monotonic
  local duration_ms

  end_time="$(step_now_utc)"
  end_monotonic="$(step_now_monotonic_ms)"
  duration_ms="$(step_elapsed_ms "${!start_monotonic_var:-0}" "${end_monotonic}")"

  printf -v "${prefix}_END_TIME" '%s' "${end_time}"
  printf -v "${prefix}_DURATION_MS" '%s' "${duration_ms}"
  printf -v "${prefix}_WALL_DURATION_MS" '%s' "${duration_ms}"
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

slugify_step_label() {
  local label="$1"
  printf '%s' "$label" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/--+/-/g'
}

resolve_harness_node() {
  if [[ -n "${NODE_BIN:-}" && -x "${NODE_BIN}" ]]; then
    printf '%s\n' "${NODE_BIN}"
    return
  fi
  if [[ -x "${RUN_STEP_REPO_ROOT}/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node" ]]; then
    printf '%s\n' "${RUN_STEP_REPO_ROOT}/${CARTULARY_LAYOUT_NODE_RUNTIME}/bin/node"
    return
  fi
  printf '%s\n' "node"
}

step_secure_mkdir() {
  if [[ "$#" -lt 1 ]]; then
    echo "step_secure_mkdir requires <dir...>" >&2
    return 2
  fi
  "$(resolve_harness_node)" "${RUN_STEP_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" secure-directories "$@"
}

step_secure_files() {
  "$(resolve_harness_node)" "${RUN_STEP_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" secure-files "$@"
}

# Capture workers own their pipes until EOF. Callers retain shell-function state
# and explicitly join both workers before publishing or replacing any artifact.
step_begin_capture() {
  local output_mode="$1" stdout_log="$2" stderr_log="$3"
  step_secure_files "$stdout_log" "$stderr_log" || return $?
  exec {STEP_CAPTURE_STDOUT_FD}> >(
    set -o pipefail
    if [[ "$output_mode" == "quiet" ]]; then
      step_redact_stream >"$stdout_log"
    else
      step_redact_stream | tee "$stdout_log"
    fi
  )
  STEP_CAPTURE_STDOUT_PID=$!
  exec {STEP_CAPTURE_STDERR_FD}> >(
    exec {STEP_CAPTURE_STDOUT_FD}>&-
    set -o pipefail
    if [[ "$output_mode" == "quiet" ]]; then
      step_redact_stream >"$stderr_log"
    else
      step_redact_stream | tee "$stderr_log" >&2
    fi
  )
  STEP_CAPTURE_STDERR_PID=$!
}

step_close_capture() {
  exec {STEP_CAPTURE_STDOUT_FD}>&-
  exec {STEP_CAPTURE_STDERR_FD}>&-
}

step_finish_capture() {
  local capture_status=0
  wait "$STEP_CAPTURE_STDOUT_PID" || capture_status=11
  wait "$STEP_CAPTURE_STDERR_PID" || capture_status=11
  if [[ "$capture_status" -ne 0 ]]; then
    echo "artifact capture or redaction failed" >&2
  fi
  return "$capture_status"
}

step_capture_command() {
  local output_mode="$1" stdout_log="$2" stderr_log="$3"
  shift 3
  local STEP_CAPTURE_STDOUT_FD STEP_CAPTURE_STDERR_FD STEP_CAPTURE_STDOUT_PID STEP_CAPTURE_STDERR_PID
  local command_status capture_status=0
  step_begin_capture "$output_mode" "$stdout_log" "$stderr_log" || return $?
  "$@" 1>&"$STEP_CAPTURE_STDOUT_FD" 2>&"$STEP_CAPTURE_STDERR_FD"
  command_status=$?
  step_close_capture
  step_finish_capture || capture_status=$?
  if [[ "$command_status" -ne 0 ]]; then return "$command_status"; fi
  return "$capture_status"
}

step_redact_stream() {
  local node_bin
  node_bin="$(resolve_harness_node)"
  "${node_bin}" "${RUN_STEP_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" redact
}

step_redact_file() {
  if [[ "$#" -ne 1 ]]; then
    echo "step_redact_file requires <file>" >&2
    return 2
  fi
  local file="$1"
  if [[ ! -e "$file" ]]; then
    return 0
  fi
  local tmp_file
  # mktemp creates the staging file exclusively with mode 0600.
  tmp_file="$(mktemp "${file}.redacted.XXXXXX")" || return $?
  if ! step_redact_stream <"$file" >"$tmp_file"; then
    rm -f -- "$tmp_file"
    return 11
  fi
  mv -- "$tmp_file" "$file" || { rm -f -- "$tmp_file"; return 11; }

}

ensure_harness_artifact_identity() {
  local target
  target="$(resolve_test_target)"
  if [[ -n "${CARTULARY_HARNESS_IDENTITY_PREPARED:-}" ]]; then
    local node_bin
    local output
    node_bin="$(resolve_harness_node)"
    output="$("${node_bin}" "${RUN_STEP_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" retained-artifact-env --existing-run-root "${target}")" || return "$?"
    CARTULARY_TEST_RESULTS_DIR="$(printf '%s\n' "${output}" | sed -n '1p')"
    CARTULARY_TEST_RUN_ID="$(printf '%s\n' "${output}" | sed -n '2p')"
    export CARTULARY_TEST_RESULTS_DIR CARTULARY_TEST_RUN_ID
    return 0
  fi

  if [[ "${target}" == "adhoc" ]]; then
    return 0
  fi

  local node_bin
  local output
  node_bin="$(resolve_harness_node)"
  output="$("${node_bin}" "${RUN_STEP_REPO_ROOT}/tools/harness/contract/harness-contract-cli.mjs" retained-artifact-env --existing-run-root "${target}")" || return "$?"
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
      printf '%s\n' "${RUN_STEP_REPO_ROOT}/${CARTULARY_TEST_RESULTS_DIR}"
    fi
    return
  fi
  printf '%s\n' "${RUN_STEP_REPO_ROOT}/.cartulary/test-results"
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
  step_secure_mkdir "${results_root}/${run_id}" "${results_root}/${run_id}/${target}"
  printf '%s\n' "${results_root}/${run_id}/${target}"
}

prepare_target_support_dir() {
  local name="${1:-support}"
  local target_dir
  target_dir="$(ensure_target_artifact_dir)"
  step_secure_mkdir "${target_dir}/${name}"
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
  step_secure_mkdir "${results_root}/${run_id}" "${results_root}/${run_id}/_shared" "${results_root}/${run_id}/_shared/${name}"
  printf '%s\n' "${results_root}/${run_id}/_shared/${name}"
}

prepare_step_artifact_dir() {
  local step="$1"
  local target_dir
  local slug
  target_dir="$(ensure_target_artifact_dir)"
  target_dir="${RUN_STEP_ARTIFACT_ROOT:-${target_dir}}"
  slug="$(slugify_step_label "$step")"
  step_secure_mkdir "${target_dir}/${slug}"
  printf '%s\n' "${target_dir}/${slug}"
}

emit_target_summary() {
  local status="${1:-pass}"
  if [[ -z "${CARTULARY_TEST_TARGET:-}" ]]; then
    return 0
  fi
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" target-summary "${CARTULARY_TEST_TARGET}" "${status}"
}

# step_publish_summary <helper-command> [field=value...] runs the step summary
# helper with a step's label, directory, command, timing, exit status and logs.
# It is the only writer of the step summary names (TH-HARNESS Section 5.5,
# single writer): wrappers pass fields, never the names. A field that is not
# given is not set; one given empty is set empty. duration_ms sets both the
# logical and the executed duration. Names specific to one step kind reach the
# helper from the caller's environment.
step_publish_summary() {
  local helper_command="$1"
  shift
  local -a assignments=()
  local argument
  local value
  for argument in "$@"; do
    value="${argument#*=}"
    case "${argument%%=*}" in
      label) assignments+=("CARTULARY_STEP_LABEL=${value}") ;;
      dir) assignments+=("CARTULARY_STEP_DIR=${value}") ;;
      command) assignments+=("CARTULARY_STEP_COMMAND=${value}") ;;
      command_argv) assignments+=("CARTULARY_STEP_COMMAND_ARGV=${value}") ;;
      start_time) assignments+=("CARTULARY_STEP_START_TIME=${value}") ;;
      end_time) assignments+=("CARTULARY_STEP_END_TIME=${value}") ;;
      duration_ms) assignments+=("CARTULARY_STEP_LOGICAL_DURATION_MS=${value}" "CARTULARY_STEP_EXECUTED_DURATION_MS=${value}") ;;
      wall_duration_ms) assignments+=("CARTULARY_STEP_WALL_DURATION_MS=${value}") ;;
      exit_status) assignments+=("CARTULARY_STEP_EXIT_STATUS=${value}") ;;
      timing_bucket) assignments+=("CARTULARY_STEP_TIMING_BUCKET=${value}") ;;
      interrupt_signal) assignments+=("CARTULARY_STEP_INTERRUPT_SIGNAL=${value}") ;;
      runner_log) assignments+=("CARTULARY_STEP_RUNNER_LOG=${value}") ;;
      stdout_log) assignments+=("CARTULARY_STEP_STDOUT_LOG=${value}") ;;
      stderr_log) assignments+=("CARTULARY_STEP_STDERR_LOG=${value}") ;;
      *)
        echo "step_publish_summary: unknown field ${argument%%=*}" >&2
        return 2
        ;;
    esac
  done
  env "${assignments[@]}" NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" "${helper_command}"
}

emit_report_step_summary() {
  if [[ "$#" -ne 8 ]]; then
    echo "emit_report_step_summary requires <helper-command> <label> <command-text> <start-time> <end-time> <duration-ms> <wall-duration-ms> <exit-status>" >&2
    return 2
  fi

  local helper_command="$1"
  local step="$2"
  local command_text="$3"
  local start_time="$4"
  local end_time="$5"
  local duration_ms
  local wall_duration_ms
  local exit_status="$8"
  local step_dir
  local helper_status

  duration_ms="$(step_clamp_duration_ms "$6")"
  wall_duration_ms="$(step_clamp_duration_ms "$7")"

  step_dir="$(prepare_step_artifact_dir "$step")"

  set +e
  step_publish_summary "${helper_command}" label="$step" dir="$step_dir" command="$command_text" \
    start_time="$start_time" end_time="$end_time" duration_ms="$duration_ms" wall_duration_ms="$wall_duration_ms" \
    exit_status="$exit_status" timing_bucket="${CARTULARY_STEP_TIMING_BUCKET:-}" \
    interrupt_signal="${CARTULARY_STEP_INTERRUPT_SIGNAL:-}"
  helper_status=$?
  set -e

  if [[ "${helper_status}" -eq 0 ]]; then
    return 0
  fi

  emit_target_summary fail || true
  return "${helper_status}"
}

run_step_command() {
  # --target and --timing-bucket set the step's report target and timing
  # bucket for the step and everything it runs, nested steps included, and
  # are restored when it returns. Callers pass them here rather than setting
  # the step runtime's names themselves (TH-HARNESS Section 5.5).
  while [[ "${1:-}" == --target || "${1:-}" == --timing-bucket ]]; do
    if [[ "$#" -lt 2 ]]; then
      echo "run_step_command: $1 requires a value" >&2
      return 2
    fi
    case "$1" in
      --target) local -x CARTULARY_TEST_TARGET="$2" ;;
      --timing-bucket) local -x CARTULARY_STEP_TIMING_BUCKET="$2" ;;
    esac
    shift 2
  done
  if [[ "$#" -lt 2 ]]; then
    echo "run_step_command requires [--target <target>] [--timing-bucket <bucket>] <label> <command...>" >&2
    return 2
  fi

  local step="$1"
  shift

  local output_mode
  local step_dir
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

  ensure_harness_artifact_identity
  output_mode="$(resolve_output_mode)"
  step_dir="$(prepare_step_artifact_dir "$step")"
  stdout_log="${step_dir}/stdout.log"
  stderr_log="${step_dir}/stderr.log"
  command_text="$(render_command "$@")"
  command_argv_json="$(render_command_argv_json "$@")"

  if [[ "$output_mode" != "quiet" && "${RUN_STEP_SHOW_BANNER:-1}" == "1" ]]; then
    echo "== ${step} =="
  fi

  step_capture_start STEP

  set +e
  CARTULARY_STEP_ARTIFACT_DIR="$step_dir" step_capture_command "$output_mode" "$stdout_log" "$stderr_log" "$@"
  status=$?
  set -e

  step_capture_finish STEP
  start_time="${STEP_START_TIME}"
  end_time="${STEP_END_TIME}"
  duration_ms="${STEP_DURATION_MS}"

  set +e
  step_publish_summary shell-step label="$step" dir="$step_dir" command="$command_text" command_argv="$command_argv_json" \
    start_time="$start_time" end_time="$end_time" duration_ms="$duration_ms" wall_duration_ms="${STEP_WALL_DURATION_MS}" \
    exit_status="$status" timing_bucket="${CARTULARY_STEP_TIMING_BUCKET:-}" stdout_log="$stdout_log" stderr_log="$stderr_log"
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
