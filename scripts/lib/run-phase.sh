#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -lt 3 ]]; then
  echo "usage: run-phase.sh \"<label>\" -- <command...>" >&2
  exit 2
fi

phase="$1"
shift

if [[ "$1" != "--" ]]; then
  echo "usage: run-phase.sh \"<label>\" -- <command...>" >&2
  exit 2
fi
shift

if [[ "$#" -eq 0 ]]; then
  echo "usage: run-phase.sh \"<label>\" -- <command...>" >&2
  exit 2
fi

command=("$@")

render_command() {
  local rendered=""
  local arg
  for arg in "$@"; do
    printf -v rendered '%s%q ' "$rendered" "$arg"
  done
  printf '%s' "${rendered% }"
}

output_mode="${CARTULARY_OUTPUT_MODE:-normal}"
if [[ "${VERBOSE:-0}" == "1" || "${CI_VERBOSE:-0}" == "1" ]]; then
  output_mode="normal"
fi

echo "== ${phase} =="

if [[ "$output_mode" != "quiet" ]]; then
  exec "${command[@]}"
fi

log_file="$(mktemp -t cartulary-phase-XXXX.log)"

set +e
"${command[@]}" >"$log_file" 2>&1
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  if [[ "${CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG:-0}" == "1" && -s "$log_file" ]]; then
    cat "$log_file"
  fi
  rm -f "$log_file"
  exit 0
fi

line_count="$(wc -l <"$log_file")"

echo "phase failed: ${phase}" >&2
echo "failing command: $(render_command "${command[@]}")" >&2
echo "phase log: $log_file" >&2

if [[ "$line_count" -le 200 ]]; then
  echo "----- phase output begin -----" >&2
  cat "$log_file" >&2
  echo "----- phase output end -----" >&2
else
  echo "----- phase output first 40 lines begin -----" >&2
  sed -n '1,40p' "$log_file" >&2
  echo "----- phase output first 40 lines end -----" >&2
  echo "----- phase output last 160 lines begin -----" >&2
  tail -n 160 "$log_file" >&2
  echo "----- phase output last 160 lines end -----" >&2
fi

exit "$status"
