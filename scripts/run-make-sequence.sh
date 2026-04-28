#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TEST_OUTPUT_SCRIPT="${TEST_OUTPUT_SCRIPT:-${ROOT_DIR}/scripts/lib/test-output.sh}"
MAKE_BIN="${MAKE:-make}"

label=""
summary_targets_csv=""
summary_groups_spec=""
sequence=""
manifest_path="${TASK_SURFACE_MANIFEST:-${ROOT_DIR}/tools/task_surface_manifest.json}"
steps=()

usage() {
  echo "usage: run-make-sequence.sh --sequence <name> [--manifest <path>]" >&2
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --sequence)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      sequence="$2"
      shift 2
      ;;
    --manifest)
      [[ "$#" -ge 2 ]] || { usage; exit 2; }
      manifest_path="$2"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "${sequence}" ]]; then
  usage
  exit 2
fi

label="${sequence}"
node_cmd="${NODE_BIN:-node}"
if [[ -n "${node_cmd}" && ! -x "${node_cmd}" ]]; then
  node_cmd="node"
fi
mapfile -t resolved_sequence < <(
  "$node_cmd" --input-type=module - "${manifest_path}" "${sequence}" <<'EOF'
import { loadTaskSurfaceManifest, sequenceDefinition, summaryProfileArgs } from "./scripts/lib/task-surface.mjs";

const [manifestPath, sequenceName] = process.argv.slice(2);
const { manifest } = loadTaskSurfaceManifest(manifestPath);
const sequence = sequenceDefinition(manifest, sequenceName);
const profile = summaryProfileArgs(manifest, sequence.summaryProfile);
const lines = [profile.summaryTargets.join(","), profile.groupsSpec];
for (const step of sequence.steps) {
  if (step.type === "step") {
    lines.push(`step:${step.target}`);
    continue;
  }
  const jobs = step.jobs ?? (step.jobsVariable ? process.env[step.jobsVariable] : "");
  if (!/^[1-9][0-9]*$/.test(String(jobs))) {
    throw new Error(`sequence ${sequenceName} parallel step ${step.target} has invalid jobs value`);
  }
  lines.push(`parallel:${step.target}:${jobs}`);
}
process.stdout.write(`${lines.join("\n")}\n`);
EOF
)
summary_targets_csv="${resolved_sequence[0]:-}"
summary_groups_spec="${resolved_sequence[1]:-}"
steps=("${resolved_sequence[@]:2}")

if [[ -z "${summary_targets_csv}" || "${#steps[@]}" -eq 0 ]]; then
  usage
  exit 2
fi

summary_targets=()
IFS=',' read -r -a summary_targets <<<"${summary_targets_csv}"
for i in "${!summary_targets[@]}"; do
  summary_targets[$i]="${summary_targets[$i]//[[:space:]]/}"
done

dry_run=0
case " ${MAKEFLAGS:-} " in
  *" n"*|*" --just-print"*|*" --dry-run"*) dry_run=1 ;;
esac

completed=0
total="${#steps[@]}"
target_count="${#summary_targets[@]}"
max_jobs=1
helper_units=()
completed_helper_units=()

is_summary_target() {
  local candidate="$1"
  local summary_target
  for summary_target in "${summary_targets[@]}"; do
    if [[ "${summary_target}" == "${candidate}" ]]; then
      return 0
    fi
  done
  return 1
}

for step in "${steps[@]}"; do
  step_kind="${step%%:*}"
  step_rest="${step#*:}"
  step_target="${step_rest%%:*}"
  if ! is_summary_target "${step_target}"; then
    helper_units+=("${step_target}")
  fi
  if [[ "${step_kind}" != "parallel" ]]; then
    continue
  fi
  step_jobs="${step_rest#*:}"
  if [[ "${step_jobs}" =~ ^[0-9]+$ ]] && (( step_jobs > max_jobs )); then
    max_jobs="${step_jobs}"
  fi
done

emit_run_start() {
  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" run-start \
    "${label}" --steps "${total}" --summary-targets "${target_count}" --helper-units "${#helper_units[@]}" --jobs "${max_jobs}"
}

emit_step_start() {
  local encoded="$1"
  local index="$2"
  local kind="${encoded%%:*}"
  local rest="${encoded#*:}"
  local target
  local jobs=1
  local mode=serial

  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  case "${kind}" in
    step)
      target="${rest}"
      ;;
    parallel)
      target="${rest%%:*}"
      jobs="${rest#*:}"
      mode=parallel
      ;;
    *)
      target="${rest}"
      ;;
  esac

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" step-start \
    "${label}" "${index}" "${total}" "${target}" --mode "${mode}" --jobs "${jobs}"
}

run_summary() {
  local status="$1"
  local aborted_after="$2"

  if [[ "${dry_run}" -eq 1 ]]; then
    return 0
  fi

  local summary_args=()
  if [[ -n "${summary_groups_spec}" ]]; then
    summary_args+=(--summary-groups "${summary_groups_spec}")
  fi
  if [[ "${#helper_units[@]}" -gt 0 ]]; then
    local helper_units_csv completed_helper_units_csv
    helper_units_csv="$(IFS=,; echo "${helper_units[*]}")"
    completed_helper_units_csv="$(IFS=,; echo "${completed_helper_units[*]}")"
    summary_args+=(--helper-units "${helper_units_csv}" --completed-helper-units "${completed_helper_units_csv}")
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_SCRIPT}" run-summary \
    "${label}" "${status}" "${completed}" "${total}" "${aborted_after}" \
    "${summary_args[@]}" "${summary_targets[@]}"
}

run_step() {
  local encoded="$1"
  local kind="${encoded%%:*}"
  local rest="${encoded#*:}"
  local target
  local jobs

  case "${kind}" in
    step)
      target="${rest}"
      "${MAKE_BIN}" --no-print-directory "${target}"
      ;;
    parallel)
      target="${rest%%:*}"
      jobs="${rest#*:}"
      if [[ -z "${target}" || -z "${jobs}" || "${jobs}" == "${rest}" ]]; then
        echo "invalid parallel step ${rest}; expected <target>:<jobs>" >&2
        return 2
      fi
      "${MAKE_BIN}" --no-print-directory --output-sync=target "-j${jobs}" "${target}"
      ;;
    *)
      echo "unknown step kind ${kind}" >&2
      return 2
      ;;
  esac
}

emit_run_start

for index in "${!steps[@]}"; do
  step="${steps[$index]}"
  step_target="${step#*:}"
  step_target="${step_target%%:*}"
  step_number=$((index + 1))

  emit_step_start "${step}" "${step_number}"

  set +e
  run_step "${step}"
  status=$?
  set -e

  if [[ "${status}" -eq 0 ]]; then
    completed=$((completed + 1))
    if ! is_summary_target "${step_target}"; then
      completed_helper_units+=("${step_target}")
    fi
    continue
  fi

  run_summary fail "${step_target}" || true
  exit "${status}"
done

run_summary pass -
