#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/phase-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/phase-runtime.sh"

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
run_phase_command "$phase" "${command[@]}"
