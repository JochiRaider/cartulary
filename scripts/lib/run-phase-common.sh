#!/usr/bin/env bash

render_command() {
  local rendered=""
  local arg
  for arg in "$@"; do
    printf -v rendered '%s%q ' "$rendered" "$arg"
  done
  printf '%s' "${rendered% }"
}

resolve_output_mode() {
  local output_mode="${CARTULARY_OUTPUT_MODE:-normal}"
  if [[ "${VERBOSE:-0}" == "1" || "${CI_VERBOSE:-0}" == "1" ]]; then
    output_mode="normal"
  fi
  printf '%s\n' "$output_mode"
}

show_phase_log_excerpt() {
  local log_file="$1"
  local line_count
  line_count="$(wc -l <"$log_file")"

  if [[ "$line_count" -le 200 ]]; then
    echo "----- phase output begin -----" >&2
    cat "$log_file" >&2
    echo "----- phase output end -----" >&2
    return
  fi

  echo "----- phase output first 40 lines begin -----" >&2
  sed -n '1,40p' "$log_file" >&2
  echo "----- phase output first 40 lines end -----" >&2
  echo "----- phase output last 160 lines begin -----" >&2
  tail -n 160 "$log_file" >&2
  echo "----- phase output last 160 lines end -----" >&2
}

emit_phase_failure() {
  local phase="$1"
  local log_file="$2"
  shift 2

  echo "phase failed: ${phase}" >&2
  echo "failing command: $(render_command "$@")" >&2
  echo "phase log: $log_file" >&2
  show_phase_log_excerpt "$log_file"
}

run_phase_command() {
  if [[ "$#" -lt 2 ]]; then
    echo "run_phase_command requires <label> <command...>" >&2
    return 2
  fi

  local phase="$1"
  shift

  local output_mode
  output_mode="$(resolve_output_mode)"

  if [[ "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
    echo "== ${phase} =="
  fi

  if [[ "$output_mode" != "quiet" ]]; then
    set +e
    "$@"
    local status=$?
    set -e
    return "$status"
  fi

  local log_file
  log_file="$(mktemp -t cartulary-phase-XXXX.log)"

  set +e
  "$@" >"$log_file" 2>&1
  local status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    if [[ "${CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG:-0}" == "1" && -s "$log_file" ]]; then
      cat "$log_file"
    fi
    rm -f "$log_file"
    return 0
  fi

  emit_phase_failure "$phase" "$log_file" "$@"

  return "$status"
}
