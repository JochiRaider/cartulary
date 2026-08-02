#!/usr/bin/env bash

CARTULARY_LIFECYCLE_SHUTDOWN_REQUESTED=0
CARTULARY_LIFECYCLE_SHUTDOWN_SIGNAL=""
declare -Ag CARTULARY_LIFECYCLE_GROUP_MONITORS=()

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

start_process_group() {
  local outvar="$1"
  local log_file="$2"
  local parent_pid="$$"
  local group_id=""
  local monitor_pid=""
  shift 2

  if [[ "$#" -eq 0 ]]; then
    echo "start_process_group requires <outvar> <log_file> <command...>" >&2
    return 2
  fi

  if [[ -n "${log_file}" ]]; then
    setsid "$@" >"${log_file}" 2>&1 &
  else
    setsid "$@" &
  fi

  group_id="$!"
  # shellcheck disable=SC2016
  setsid bash -c '
    parent_pid="$1"
    group_id="$2"

    while kill -0 "$parent_pid" >/dev/null 2>&1 && kill -0 "$group_id" >/dev/null 2>&1; do
      sleep 0.2
    done

    if kill -0 "$parent_pid" >/dev/null 2>&1 || ! kill -0 "$group_id" >/dev/null 2>&1; then
      exit 0
    fi

    kill -TERM -- "-${group_id}" >/dev/null 2>&1 || true
    for _ in $(seq 1 50); do
      if ! kill -0 "$group_id" >/dev/null 2>&1; then
        exit 0
      fi
      sleep 0.2
    done

    kill -KILL -- "-${group_id}" >/dev/null 2>&1 || true
  ' bash "${parent_pid}" "${group_id}" >/dev/null 2>&1 &
  monitor_pid="$!"
  CARTULARY_LIFECYCLE_GROUP_MONITORS["${group_id}"]="${monitor_pid}"

  printf -v "${outvar}" '%s' "${group_id}"
}

process_group_running() {
  local group_id="${1:-}"

  [[ -n "${group_id}" ]] && kill -0 -- "-${group_id}" >/dev/null 2>&1
}

wait_for_process_group_exit() {
  local group_id="${1:-}"
  local attempts="${2:-50}"
  local interval_seconds="${3:-0.2}"

  if [[ -z "${group_id}" ]]; then
    return 0
  fi

  for _ in $(seq 1 "${attempts}"); do
    if ! process_group_running "${group_id}"; then
      return 0
    fi
    sleep "${interval_seconds}"
  done

  return 1
}

stop_process_group() {
  local group_id="${1:-}"
  local attempts="${2:-50}"
  local interval_seconds="${3:-0.2}"
  local monitor_pid=""

  if [[ -z "${group_id}" ]]; then
    return 0
  fi

  monitor_pid="${CARTULARY_LIFECYCLE_GROUP_MONITORS[$group_id]:-}"

  if [[ -n "${monitor_pid}" ]]; then
    kill "${monitor_pid}" >/dev/null 2>&1 || true
    wait "${monitor_pid}" >/dev/null 2>&1 || true
    unset "CARTULARY_LIFECYCLE_GROUP_MONITORS[$group_id]"
  fi

  if process_group_running "${group_id}"; then
    kill -TERM -- "-${group_id}" >/dev/null 2>&1 || true
    if ! wait_for_process_group_exit "${group_id}" "${attempts}" "${interval_seconds}"; then
      kill -KILL -- "-${group_id}" >/dev/null 2>&1 || true
      wait_for_process_group_exit "${group_id}" "${attempts}" "${interval_seconds}" || true
    fi
  fi

  wait "${group_id}" >/dev/null 2>&1 || true
}
