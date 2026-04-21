#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-playwright-manifest-phase.sh \"<label>\" <phase> <coverage> [<execution_dependency>] -- <playwright test command...>" >&2
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

mapfile -t manifest_files < <("$node_bin" "$manifest_script" playwright-files "$phase_manifest" "$coverage" "$execution_dependency")
grep_pattern="$("$node_bin" "$manifest_script" playwright-grep "$phase_manifest" "$coverage" "$execution_dependency")"
list_report="$(mktemp -t cartulary-playwright-list-XXXX.json)"
run_report="$(mktemp -t cartulary-playwright-run-XXXX.json)"
output_mode="$(resolve_output_mode)"

echo "== ${phase_label} =="

list_command=("${command[@]}" --list --reporter=json -g "$grep_pattern" "${manifest_files[@]}")
run_command=("${command[@]}" --reporter=json -g "$grep_pattern" "${manifest_files[@]}")

set +e
"${list_command[@]}" >"$list_report" 2>&1
list_status=$?
set -e
if [[ "$list_status" -ne 0 ]]; then
  echo "phase failed: ${phase_label} manifest list" >&2
  echo "failing command: $(render_command "${list_command[@]}")" >&2
  echo "----- playwright list output begin -----" >&2
  cat "$list_report" >&2
  echo "----- playwright list output end -----" >&2
  exit "$list_status"
fi

"$node_bin" "$manifest_script" playwright-verify-list "$phase_manifest" "$coverage" "$execution_dependency" "$list_report"

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
  verify_output="$("$node_bin" "$manifest_script" playwright-verify-run "$phase_manifest" "$coverage" "$execution_dependency" "$run_report" 2>&1)"
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
  if [[ "$verify_output" == playwright\ setup\ failure:* ]]; then
    echo "phase failed: ${phase_label}" >&2
  else
    echo "phase failed: ${phase_label} manifest verification" >&2
  fi
  echo "----- playwright run output begin -----" >&2
  cat "$run_report" >&2
  echo "----- playwright run output end -----" >&2
  exit "$verify_status"
fi

if [[ "$run_status" -eq 0 ]]; then
  rm -f "$list_report" "$run_report"
  exit 0
fi

if [[ "$output_mode" == "quiet" ]]; then
  echo "phase failed: ${phase_label}" >&2
  echo "failing command: $(render_command "${run_command[@]}")" >&2
  echo "----- playwright run output begin -----" >&2
  cat "$run_report" >&2
  echo "----- playwright run output end -----" >&2
fi

exit "$run_status"
