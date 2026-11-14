#!/usr/bin/env bash

# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"

CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED=0
CARTULARY_LIFECYCLE_SHUTDOWN_SIGNAL=""
CARTULARY_LIFECYCLE_ACQUISITION_DEPTH=0
declare -Ag CARTULARY_LIFECYCLE_PIDS=()
declare -Ag CARTULARY_LIFECYCLE_RESULTS=()

lifecycle_reset_shutdown_state() {
  CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED=0
  CARTULARY_LIFECYCLE_SHUTDOWN_SIGNAL=""
}

lifecycle_request_shutdown() {
  local signal_name="${1:-TERM}"

  if [[ "${CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED}" -eq 0 ]]; then
    CARTULARY_LIFECYCLE_SHUTDOWN_SIGNAL="${signal_name}"
  fi
  CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED=1
}

lifecycle_exit_on_signal() {
  local signal_name="${1:-TERM}"
  local exit_status=0

  lifecycle_request_shutdown "${signal_name}"
  # Publish or settle an in-flight native acquisition before the exit trap can
  # release dependencies. Parent-loss cleanup remains the native fallback.
  if [[ "${CARTULARY_LIFECYCLE_ACQUISITION_DEPTH}" -gt 0 ]]; then return 0; fi
  exit_status="$(lifecycle_signal_exit_status)"
  exit "${exit_status}"
}

lifecycle_install_signal_traps() {
  trap 'lifecycle_exit_on_signal INT' INT
  trap 'lifecycle_exit_on_signal TERM' TERM
}

lifecycle_shutdown_requested() {
  [[ "${CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED}" -eq 1 ]]
}

lifecycle_signal_name() {
  printf '%s\n' "${CARTULARY_LIFECYCLE_SHUTDOWN_SIGNAL:-TERM}"
}

lifecycle_signal_exit_status() {
  case "$(lifecycle_signal_name)" in
    INT)
      printf '%s\n' 130
      ;;
    TERM)
      printf '%s\n' 143
      ;;
    *)
      printf '%s\n' 1
      ;;
  esac
}

# Detached native owners and workloads must not retain a step capture pipe.
lifecycle_close_step_capture_descriptors() {
  if [[ -n "${STEP_CAPTURE_STDOUT_FD:-}" ]]; then exec {STEP_CAPTURE_STDOUT_FD}>&-; fi
  if [[ -n "${STEP_CAPTURE_STDERR_FD:-}" ]]; then exec {STEP_CAPTURE_STDERR_FD}>&-; fi
}

lifecycle_native() {
  local lifecycle_root
  lifecycle_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)" || return $?
  "${CARTULARY_HARNESS_RUNTIME_BIN:-${lifecycle_root}/${CARTULARY_LAYOUT_TOOLBIN}/cartulary-harness}" owned-command "$@"
}

start_owned_command() {
  local lifecycle_status=0
  CARTULARY_LIFECYCLE_ACQUISITION_DEPTH=$((CARTULARY_LIFECYCLE_ACQUISITION_DEPTH + 1))
  lifecycle_acquire_owned_command "$@" || lifecycle_status=$?
  CARTULARY_LIFECYCLE_ACQUISITION_DEPTH=$((CARTULARY_LIFECYCLE_ACQUISITION_DEPTH - 1))
  if lifecycle_shutdown_requested; then return "$(lifecycle_signal_exit_status)"; fi
  return "${lifecycle_status}"
}

lifecycle_acquire_owned_command() {
  local outvar="$1" log_file="$2" acquired_handle observed_pid status=0 lifecycle_root
  shift 2
  if [[ "$#" -eq 0 ]]; then return 2; fi
  lifecycle_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)" || return $?
  if [[ -n "${log_file}" ]]; then
    "${NODE_BIN:-node}" "${lifecycle_root}/tools/harness/contract/harness-contract-cli.mjs" secure-files "${log_file}" || return $?
  fi
  acquired_handle="$(lifecycle_native allocate)" || return $?
  # The native owner survives parent loss and observes its parent's exact kernel
  # identity. It alone launches, settles and reaps the workload's process group.
  if [[ -n "${log_file}" ]]; then
    ( lifecycle_close_step_capture_descriptors
      exec setsid "${CARTULARY_HARNESS_RUNTIME_BIN:-${lifecycle_root}/${CARTULARY_LAYOUT_TOOLBIN}/cartulary-harness}" owned-command serve "${acquired_handle}" "$@"
    ) >"${log_file}" 2>&1 &
  else
    ( lifecycle_close_step_capture_descriptors
      exec setsid "${CARTULARY_HARNESS_RUNTIME_BIN:-${lifecycle_root}/${CARTULARY_LAYOUT_TOOLBIN}/cartulary-harness}" owned-command serve "${acquired_handle}" "$@"
    ) &
  fi
  if observed_pid="$(lifecycle_native ready "${acquired_handle}")"; then
    CARTULARY_LIFECYCLE_PIDS["${acquired_handle}"]="${observed_pid}"
    printf -v "${outvar}" '%s' "${acquired_handle}"
  else
    status=$?
    # Require the native receipt; EOF or a recycled numeric PID is no authority.
    if lifecycle_native stop "${acquired_handle}"; then
      lifecycle_native release "${acquired_handle}" || status=$?
    fi
    echo "owned command acquisition failed" >&2
    return "${status}"
  fi
}

# PID is a projection for port/artifact observations, never a control key.
owned_command_pid() {
  local command_handle="${1:-}"
  [[ -n "${command_handle}" ]] || return 12
  [[ -n "${CARTULARY_LIFECYCLE_PIDS[$command_handle]:-}" ]] || return 12
  printf '%s\n' "${CARTULARY_LIFECYCLE_PIDS[$command_handle]}"
}

owned_command_running() {
  local command_handle="${1:-}"
  [[ -n "${command_handle}" ]] || return 1
  if [[ -n "${CARTULARY_LIFECYCLE_RESULTS[$command_handle]+present}" ]]; then return 1; fi
  [[ -n "${CARTULARY_LIFECYCLE_PIDS[$command_handle]:-}" ]] || return 12
  lifecycle_native running "${command_handle}"
}

owned_command_result() {
  local command_handle="${1:-}"
  [[ -n "${command_handle}" ]] || return 12
  if [[ -n "${CARTULARY_LIFECYCLE_RESULTS[$command_handle]+present}" ]]; then
    return "${CARTULARY_LIFECYCLE_RESULTS[$command_handle]}"
  fi
  [[ -n "${CARTULARY_LIFECYCLE_PIDS[$command_handle]:-}" ]] || return 12
  lifecycle_native result "${command_handle}"
}

stop_owned_command() {
  local command_handle="${1:-}" native_status=0
  [[ -n "${command_handle}" ]] || return 0
  if [[ -n "${CARTULARY_LIFECYCLE_RESULTS[$command_handle]+present}" ]]; then return 0; fi
  [[ -n "${CARTULARY_LIFECYCLE_PIDS[$command_handle]:-}" ]] || return 12
  lifecycle_native stop "${command_handle}" || return $?
  lifecycle_native result "${command_handle}" || native_status=$?
  lifecycle_native release "${command_handle}" || return $?
  CARTULARY_LIFECYCLE_RESULTS["${command_handle}"]="${native_status}"
  unset 'CARTULARY_LIFECYCLE_PIDS[$command_handle]'
}
