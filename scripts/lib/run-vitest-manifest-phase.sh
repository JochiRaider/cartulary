#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-vitest-manifest-phase.sh \"<label>\" <phase> <coverage> [<execution_dependency>] -- <vitest run command...>" >&2
  exit 2
}

if [[ "$#" -lt 5 ]]; then
  usage
fi

phase_label="$1"
phase_manifest="$2"
coverage="$3"
shift 3

execution_dependency=""
if [[ "$1" != "--" ]]; then
  execution_dependency="$1"
  shift
fi

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

command=("$@")
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/scripts/lib/phase-manifest.mjs"

mapfile -t manifest_files < <("$node_bin" "$manifest_script" vitest-files "$phase_manifest" "$coverage" "$execution_dependency")
grep_pattern="$("$node_bin" "$manifest_script" vitest-grep "$phase_manifest" "$coverage" "$execution_dependency")"
run_report="$(mktemp -t cartulary-vitest-run-XXXX.json)"
output_mode="$(resolve_output_mode)"

echo "== ${phase_label} =="

run_command=("${command[@]}" --reporter=json -t "$grep_pattern" "${manifest_files[@]}")

set +e
if [[ "$output_mode" != "quiet" ]]; then
  "${run_command[@]}" | tee "$run_report"
  run_status=$?
else
  "${run_command[@]}" >"$run_report" 2>&1
  run_status=$?
fi
set -e

verify_output=""
verify_status=0
if [[ "$run_status" -eq 0 ]]; then
  set +e
  verify_output="$("$node_bin" "$manifest_script" vitest-verify-run "$phase_manifest" "$coverage" "$execution_dependency" "$run_report" 2>&1)"
  verify_status=$?
  set -e
fi

if [[ -n "$verify_output" ]]; then
  if [[ "$verify_status" -eq 0 ]]; then
    printf '%s\n' "$verify_output"
  else
    printf '%s\n' "$verify_output" >&2
  fi
fi

if [[ "$verify_status" -ne 0 && "$run_status" -eq 0 ]]; then
  echo "phase failed: ${phase_label} manifest verification" >&2
  echo "----- vitest run output begin -----" >&2
  cat "$run_report" >&2
  echo "----- vitest run output end -----" >&2
  exit "$verify_status"
fi

if [[ "$run_status" -eq 0 ]]; then
  rm -f "$run_report"
  exit 0
fi

if [[ "$output_mode" == "quiet" ]]; then
  echo "phase failed: ${phase_label}" >&2
  echo "failing command: $(render_command "${run_command[@]}")" >&2
  echo "----- vitest run output begin -----" >&2
  cat "$run_report" >&2
  echo "----- vitest run output end -----" >&2
fi

exit "$run_status"
