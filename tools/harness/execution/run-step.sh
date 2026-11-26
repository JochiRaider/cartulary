#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/step-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/step-runtime.sh"

usage() {
  echo "usage: run-step.sh [--target <target>] [--timing-bucket <bucket>] \"<label>\" -- <command...>" >&2
  exit 2
}

options=()
while [[ "${1:-}" == --target || "${1:-}" == --timing-bucket ]]; do
  [[ "$#" -ge 2 ]] || usage
  options+=("$1" "$2")
  shift 2
done

[[ "$#" -ge 3 ]] || usage
step="$1"
shift
[[ "$1" == "--" ]] || usage
shift
[[ "$#" -gt 0 ]] || usage

command=("$@")
run_step_command "${options[@]}" "$step" "${command[@]}"
