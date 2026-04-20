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
list_command=("${command[@]}" -list "$pattern")
run_command=("${command[@]}" -run "$pattern")

set +e
list_output="$("${list_command[@]}" 2>&1)"
list_status=$?
set -e

echo "== ${phase} =="

if [[ "$list_status" -ne 0 ]]; then
  echo "phase failed: ${phase} inventory" >&2
  echo "failing command: $(render_command "${list_command[@]}")" >&2
  echo "----- phase output begin -----" >&2
  printf '%s\n' "$list_output" >&2
  echo "----- phase output end -----" >&2
  exit "$list_status"
fi

matched_packages=0
matched_tests=0
inventory=""
pending_tests=()

while IFS= read -r line; do
  if [[ "$line" =~ ^(Test|Benchmark|Fuzz) ]]; then
    pending_tests+=("$line")
    continue
  fi

  if [[ "$line" =~ ^(ok|\?|FAIL)[[:space:]]+([^[:space:]]+) ]]; then
    if [[ ${#pending_tests[@]} -eq 0 ]]; then
      continue
    fi

    pkg="${BASH_REMATCH[2]}"
    matched_packages=$((matched_packages + 1))
    inventory+=$'  '"$pkg"$'\n'
    for test_name in "${pending_tests[@]}"; do
      matched_tests=$((matched_tests + 1))
      inventory+=$'    '"$test_name"$'\n'
    done
    pending_tests=()
  fi
done <<<"$list_output"

if [[ ${#pending_tests[@]} -ne 0 ]]; then
  echo "phase inventory parse failed: unmatched test names remained without a package status line" >&2
  echo "inventory command: $(render_command "${list_command[@]}")" >&2
  echo "----- phase output begin -----" >&2
  printf '%s\n' "$list_output" >&2
  echo "----- phase output end -----" >&2
  exit 1
fi

if [[ "$matched_tests" -eq 0 ]]; then
  echo "phase matched zero tests" >&2
  echo "inventory command: $(render_command "${list_command[@]}")" >&2
  exit 1
fi

echo "matched go tests: ${matched_tests} across ${matched_packages} packages"
printf '%s' "$inventory"

RUN_PHASE_SHOW_BANNER=0 run_phase_command "$phase" "${run_command[@]}"
