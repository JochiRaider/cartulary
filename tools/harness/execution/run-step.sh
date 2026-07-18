#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/step-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/step-runtime.sh"

if [[ "$#" -lt 3 ]]; then
  echo "usage: run-step.sh \"<label>\" -- <command...>" >&2
  exit 2
fi

step="$1"
shift

if [[ "$1" != "--" ]]; then
  echo "usage: run-step.sh \"<label>\" -- <command...>" >&2
  exit 2
fi
shift

if [[ "$#" -eq 0 ]]; then
  echo "usage: run-step.sh \"<label>\" -- <command...>" >&2
  exit 2
fi

command=("$@")
run_step_command "$step" "${command[@]}"
