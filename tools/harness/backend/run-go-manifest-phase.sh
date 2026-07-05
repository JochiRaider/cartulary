#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/phase-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)/tools/harness/execution/phase-runtime.sh"

usage() {
  echo "usage: run-go-manifest-phase.sh \"<label>\" <phase> <section> <coverage> [<execution_dependency>] -- <go test command...>" >&2
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
repo_root="$(cd "$script_dir/../../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/tools/harness/phase-accounting/phase-manifest.mjs"
if [[ -n "${CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION:-}" ]]; then
  echo "CARTULARY_ALLOW_EMPTY_MANIFEST_SELECTION is retired; use tools/phase_policy_exceptions.json for temporary empty manifest selection exceptions" >&2
  exit 2
fi
match_count="$("$node_bin" "$manifest_script" go-count "$phase_manifest" "$section" "$coverage" "$execution_dependency" "${package_patterns[@]}")"
if [[ "$match_count" == "0" ]]; then
  allow_output=""
  allow_status=0
  set +e
  allow_output="$(
    "$node_bin" "$manifest_script" empty-go-manifest-selection-allowed \
      "$phase_manifest" "$section" "$coverage" "$execution_dependency" "${package_patterns[@]}" 2>&1
  )"
  allow_status=$?
  set -e
  if [[ "$allow_status" -eq 0 ]]; then
    exit 0
  fi
  if [[ -n "$allow_output" ]]; then
    printf '%s\n' "$allow_output" >&2
    exit "$allow_status"
  fi
  echo "no ${coverage} go tests found for ${phase_manifest} ${section} in ${package_patterns[*]}" >&2
  exit 1
fi

pattern="$("$node_bin" "$manifest_script" go-regex "$phase_manifest" "$section" "$coverage" "$execution_dependency" "${package_patterns[@]}")"
run_command=("${prefix[@]}" test -json -run "$pattern" "${suffix[@]}")
output_mode="$(resolve_output_mode)"
phase_dir="$(prepare_phase_artifact_dir "$phase_label")"
log_file="${phase_dir}/runner.jsonl"
stderr_file="${phase_dir}/stderr.log"
command_text="$(render_command "${run_command[@]}")"
phase_capture_start PHASE

if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${phase_label} =="
fi

set +e
if [[ "$output_mode" != "quiet" ]]; then
  "${run_command[@]}" > >(tee "$log_file" | stream_go_json_output) 2> >(tee "$stderr_file" >&2)
  run_status=$?
else
  "${run_command[@]}" >"$log_file" 2>"$stderr_file"
  run_status=$?
fi
set -e

phase_capture_finish PHASE
start_time="${PHASE_START_TIME}"
end_time="${PHASE_END_TIME}"
duration_ms="${PHASE_DURATION_MS}"

set +e
CARTULARY_PHASE_LABEL="$phase_label" \
CARTULARY_PHASE_DIR="$phase_dir" \
CARTULARY_PHASE_COMMAND="$command_text" \
CARTULARY_PHASE_START_TIME="$start_time" \
CARTULARY_PHASE_END_TIME="$end_time" \
CARTULARY_PHASE_DURATION_MS="$duration_ms" \
CARTULARY_PHASE_WALL_DURATION_MS="${PHASE_WALL_DURATION_MS}" \
CARTULARY_PHASE_EXIT_STATUS="$run_status" \
CARTULARY_PHASE_RUNNER_LOG="$log_file" \
CARTULARY_PHASE_STDERR_LOG="$stderr_file" \
CARTULARY_MANIFEST_PHASE="$phase_manifest" \
CARTULARY_MANIFEST_SECTION="$section" \
CARTULARY_MANIFEST_COVERAGE="$coverage" \
CARTULARY_MANIFEST_EXECUTION_DEPENDENCY="$execution_dependency" \
CARTULARY_GO_PACKAGE_PATTERNS="$(printf '%s\n' "${package_patterns[@]}")" \
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" go-manifest-phase
helper_status=$?
set -e

if [[ "$run_status" -eq 0 && "$helper_status" -eq 0 ]]; then
  exit 0
fi

emit_target_summary fail || true

if [[ "$run_status" -ne 0 ]]; then
  exit "$run_status"
fi

exit "$helper_status"
