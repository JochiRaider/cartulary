#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-go-phase.sh \"<label>\" \"<pattern>\" -- <go test command...>" >&2
  exit 2
}

if [[ "$#" -lt 4 ]]; then
  usage
fi

phase="$1"
pattern="$2"
shift 2

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

command=("$@")
test_index=-1
for i in "${!command[@]}"; do
  if [[ "${command[$i]}" == "test" ]]; then
    test_index="$i"
    break
  fi
done

if [[ "$test_index" -lt 1 ]]; then
  echo "usage: run-go-phase.sh \"<label>\" \"<pattern>\" -- <go test command...>" >&2
  echo "expected a go test command after --" >&2
  exit 2
fi

prefix=("${command[@]:0:$test_index}")
suffix=("${command[@]:$((test_index + 1))}")
run_command=("${prefix[@]}" test -json -run "$pattern" "${suffix[@]}")
log_file="$(mktemp -t cartulary-phase-XXXX.log)"
stderr_file="$(mktemp -t cartulary-phase-XXXX.stderr.log)"
inventory_command=("${prefix[@]}" run ./tools/gotestinventory "$log_file")
output_mode="$(resolve_output_mode)"

echo "== ${phase} =="

set +e
if [[ "$output_mode" != "quiet" ]]; then
  "${run_command[@]}" > >(tee "$log_file") 2> >(tee "$stderr_file" >&2)
  run_status=$?
else
  "${run_command[@]}" >"$log_file" 2>"$stderr_file"
  run_status=$?
fi
set -e

set +e
inventory_output="$("${inventory_command[@]}" 2>&1)"
inventory_status=$?
set -e

if [[ -n "$inventory_output" ]]; then
  if [[ "$inventory_status" -eq 0 ]]; then
    printf '%s\n' "$inventory_output"
  else
    printf '%s\n' "$inventory_output" >&2
  fi
fi

if [[ "$inventory_status" -ne 0 && "$run_status" -eq 0 ]]; then
  echo "phase failed: ${phase} inventory" >&2
  echo "failing command: $(render_command "${inventory_command[@]}")" >&2
  echo "phase log: $log_file" >&2
  if [[ -s "$stderr_file" ]]; then
    echo "phase stderr log: $stderr_file" >&2
  else
    rm -f "$stderr_file"
  fi
  show_phase_log_excerpt "$log_file"
  exit "$inventory_status"
fi

if [[ "$run_status" -eq 0 ]]; then
  rm -f "$log_file"
  rm -f "$stderr_file"
  exit 0
fi

if [[ "$output_mode" == "quiet" ]]; then
  emit_phase_failure "$phase" "$log_file" "${run_command[@]}"
  if [[ -s "$stderr_file" ]]; then
    echo "phase stderr: $stderr_file" >&2
  else
    rm -f "$stderr_file"
  fi
fi

exit "$run_status"
