#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-go-manifest-phase.sh \"<label>\" <phase> <section> <coverage> -- <go test command...>" >&2
  exit 2
}

if [[ "$#" -lt 6 ]]; then
  usage
fi

phase_label="$1"
phase_manifest="$2"
section="$3"
coverage="$4"
shift 4

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
  echo "expected a go test command after --" >&2
  exit 2
fi

prefix=("${command[@]:0:$test_index}")
suffix=("${command[@]:$((test_index + 1))}")
package_patterns=()
for arg in "${suffix[@]}"; do
  if [[ "$arg" == -* ]]; then
    break
  fi
  package_patterns+=("$arg")
done

if [[ "${#package_patterns[@]}" -eq 0 ]]; then
  echo "manifest go phase requires at least one package pattern before test flags" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/scripts/lib/phase-manifest.mjs"

pattern="$("$node_bin" "$manifest_script" go-regex "$phase_manifest" "$section" "$coverage" "${package_patterns[@]}")"
run_command=("${prefix[@]}" test -json -run "$pattern" "${suffix[@]}")
log_file="$(mktemp -t cartulary-phase-XXXX.log)"
output_mode="$(resolve_output_mode)"

echo "== ${phase_label} =="

set +e
if [[ "$output_mode" != "quiet" ]]; then
  "${run_command[@]}" > >(tee "$log_file") 2>&1
  run_status=$?
else
  "${run_command[@]}" >"$log_file" 2>&1
  run_status=$?
fi
set -e

set +e
inventory_output="$("$node_bin" "$manifest_script" go-verify-log "$phase_manifest" "$section" "$coverage" "$log_file" "${package_patterns[@]}" 2>&1)"
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
  echo "phase failed: ${phase_label} manifest verification" >&2
  echo "phase log: $log_file" >&2
  show_phase_log_excerpt "$log_file"
  exit "$inventory_status"
fi

if [[ "$run_status" -eq 0 ]]; then
  rm -f "$log_file"
  exit 0
fi

if [[ "$output_mode" == "quiet" ]]; then
  emit_phase_failure "$phase_label" "$log_file" "${run_command[@]}"
fi

exit "$run_status"
