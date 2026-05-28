#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=scripts/lib/run-phase-common.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-playwright-webserver-batch.sh <webserver-backed|functional|support> -- <playwright test command...>" >&2
  echo "       run-playwright-webserver-batch.sh functional-shard <shard-name> <shard-index> <shard-count> -- <playwright test command...>" >&2
  exit 2
}

if [[ "$#" -lt 3 ]]; then
  usage
fi

mode="$1"
shift
single_shard_name=""
single_shard_index=""
single_shard_count=""

if [[ "$mode" == "functional-shard" ]]; then
  if [[ "$#" -lt 5 ]]; then
    usage
  fi
  single_shard_name="$1"
  single_shard_index="$2"
  single_shard_count="$3"
  shift 3
  if [[ ! "$single_shard_index" =~ ^[0-9]+$ ]] || (( single_shard_index < 0 )); then
    echo "functional-shard index must be a non-negative integer" >&2
    exit 2
  fi
  if [[ ! "$single_shard_count" =~ ^[0-9]+$ ]] || (( single_shard_count < 1 )); then
    echo "functional-shard count must be a positive integer" >&2
    exit 2
  fi
  if (( single_shard_index >= single_shard_count )); then
    echo "functional-shard index must be less than shard count" >&2
    exit 2
  fi
fi

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

case "$mode" in
  webserver-backed | functional | support | functional-shard)
    ;;
  *)
    usage
    ;;
esac

command=("$@")
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/scripts/lib/phase-manifest.mjs"
shard_plan_script="$repo_root/scripts/lib/browser-shard-plan.mjs"
output_mode="$(resolve_output_mode)"
batch_label="playwright-${mode}-batch"
if [[ "$mode" == "functional-shard" ]]; then
  batch_label="playwright-${mode}-${single_shard_name}-batch"
fi
batch_dir="$(prepare_target_support_dir "$batch_label")"
run_report="${batch_dir}/runner.json"
support_report="${batch_dir}/support-runner.json"
stdout_log="${batch_dir}/stdout.log"
stderr_log="${batch_dir}/stderr.log"
support_stdout_log="${batch_dir}/support-stdout.log"
support_stderr_log="${batch_dir}/support-stderr.log"
output_dir="${batch_dir}/playwright-output"
shard_plan="${batch_dir}/functional-shards.json"
phase_filter="${CARTULARY_PHASE_SLICE_PHASE:-}"

resolve_functional_shard_limit() {
  local configured="${BROWSER_E2E_FUNCTIONAL_SHARDS:-auto}"
  if [[ -z "$configured" || "$configured" == "auto" ]]; then
    PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-2}" "$node_bin" <<'EOF'
const workers = Number.parseInt(process.env.PLAYWRIGHT_WORKERS ?? "2", 10);
process.stdout.write(String(Number.isInteger(workers) && workers > 0 ? workers : 2));
EOF
    return 0
  fi
  if [[ "$configured" =~ ^[0-9]+$ ]] && (( configured >= 1 )); then
    printf '%s\n' "$configured"
    return 0
  fi
  echo "BROWSER_E2E_FUNCTIONAL_SHARDS must be auto or a positive integer" >&2
  return 2
}

if [[ "$mode" == "functional-shard" ]]; then
  functional_shard_limit="$single_shard_count"
else
  functional_shard_limit="$(resolve_functional_shard_limit)"
fi
playwright_worker_count="${CARTULARY_PLAYWRIGHT_WORKER_COUNT:-$functional_shard_limit}"

resolve_functional_worker_offset() {
  local shard_index="$1"
  if [[ "$mode" == "functional-shard" && -n "${CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET:-}" ]]; then
    printf '%s\n' "$CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET"
    return 0
  fi
  printf '%s\n' "$shard_index"
}

if [[ "$mode" != "support" ]]; then
  shard_plan_command=("$node_bin" "$shard_plan_script" plan --max-shards "$functional_shard_limit")
  if [[ -n "$phase_filter" ]]; then
    shard_plan_command+=(--phase "$phase_filter")
  fi
  "${shard_plan_command[@]}" >"$shard_plan"
else
  cat >"$shard_plan" <<JSON
{
  "schema_id": "cartulary.browser_e2e_shard_plan.v2",
  "phase": "${phase_filter}",
  "entry_count": 0,
  "file_count": 0,
  "files": [],
  "entries": [],
  "shards": []
}
JSON
fi

shard_names=()
if [[ "$mode" != "support" ]]; then
  mapfile -t shard_names < <(
    "$node_bin" - "$shard_plan" <<'EOF'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
for (const shard of plan.shards ?? []) {
  process.stdout.write(`${shard.name}\n`);
}
EOF
  )
  if [[ "$mode" == "functional-shard" ]]; then
    shard_found=0
    for shard in "${shard_names[@]}"; do
      if [[ "$shard" == "$single_shard_name" ]]; then
        shard_found=1
        break
      fi
    done
    if [[ "$shard_found" -ne 1 ]]; then
      echo "browser functional shard plan did not contain ${single_shard_name}" >&2
      exit 1
    fi
    shard_names=("$single_shard_name")
  fi
fi

if [[ "$mode" != "support" && "${#shard_names[@]}" -eq 0 ]]; then
  echo "browser functional shard plan produced no shards" >&2
  exit 1
fi

functional_phases=()
if [[ "$mode" != "support" ]]; then
  mapfile -t functional_phases < <(
    "$node_bin" - "$shard_plan" "$single_shard_name" <<'EOF'
const fs = require("node:fs");
const [planPath, selectedShardName] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planPath, "utf8"));
const phases = new Set();
for (const shard of plan.shards ?? []) {
  if (selectedShardName && shard.name !== selectedShardName) {
    continue;
  }
  for (const phase of shard.phases ?? []) {
    phases.add(phase);
  }
}
for (const phase of [...phases].sort((left, right) => left.localeCompare(right, undefined, { numeric: true }))) {
  process.stdout.write(`${phase}\n`);
}
EOF
  )
fi

if [[ -n "$phase_filter" ]]; then
  support_count="$("$node_bin" "$manifest_script" playwright-count "$phase_filter" supplemental browser_support)"
  if [[ "$support_count" != "0" ]]; then
    support_phases=("$phase_filter")
  else
    support_phases=()
  fi
else
  mapfile -t support_phases < <(
    "$node_bin" "$manifest_script" playwright-phases supplemental browser_support 2>/dev/null || true
  )
fi

support_selection_specs=()
for support_phase in "${support_phases[@]}"; do
  support_selection_specs+=("${support_phase}:supplemental:browser_support")
done

if [[ "$mode" != "support" ]]; then
  all_functional_grep="$(
    "$node_bin" - "$shard_plan" <<'EOF'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const escapeRegex = (value) => value.replace(/[\\^$.*+?()[\]{}|]/g, "\\$&");
const titles = [];
for (const shard of plan.shards ?? []) {
  for (const entry of shard.entries ?? []) {
    for (const title of entry.titles ?? [entry.title]) {
      titles.push(title);
    }
  }
}
process.stdout.write(`^(?:${titles.map(escapeRegex).join("|")})$`);
EOF
  )"
  all_functional_files="$(
    "$node_bin" - "$shard_plan" <<'EOF'
const fs = require("node:fs");
const plan = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const files = new Set();
for (const file of plan.files ?? []) {
  files.add(file);
}
for (const entry of plan.entries ?? []) {
  files.add(entry.file);
}
process.stdout.write(`${[...files].sort().join("\n")}\n`);
EOF
  )"
else
  all_functional_grep="(?!)"
  all_functional_files="__no_browser_functional__.spec.ts"
fi
if [[ "${#support_selection_specs[@]}" -gt 0 ]]; then
  all_support_grep="$("$node_bin" "$manifest_script" playwright-grep-many "${support_selection_specs[@]}")"
  all_support_files="$("$node_bin" "$manifest_script" playwright-files-many "${support_selection_specs[@]}")"
else
  all_support_grep="(?!)"
  all_support_files="__no_browser_support__.spec.ts"
fi

if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== browser-e2e-${mode} duration-balanced batch =="
fi

phase_capture_start BATCH

shard_grep() {
  local shard="$1"
  "$node_bin" - "$shard_plan" "$shard" <<'EOF'
const fs = require("node:fs");
const [planPath, shardName] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planPath, "utf8"));
const shard = (plan.shards ?? []).find((entry) => entry.name === shardName);
if (!shard) {
  throw new Error(`missing shard ${shardName}`);
}
process.stdout.write(shard.grep);
EOF
}

shard_files() {
  local shard="$1"
  "$node_bin" - "$shard_plan" "$shard" <<'EOF'
const fs = require("node:fs");
const [planPath, shardName] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planPath, "utf8"));
const shard = (plan.shards ?? []).find((entry) => entry.name === shardName);
if (!shard) {
  throw new Error(`missing shard ${shardName}`);
}
process.stdout.write(`${(shard.files ?? []).join("\n")}\n`);
EOF
}

write_missing_report() {
  local report="$1"
  local message="$2"
  if [[ -s "$report" ]]; then
    return 0
  fi
  cat >"$report" <<JSON
{
  "suites": [],
  "errors": [
    {
      "message": "$message"
    }
  ]
}
JSON
}

run_functional_shard() {
  local shard="$1"
  local shard_index="$2"
  local grep
  local files
  local shard_report
  local shard_stdout
  local shard_stderr
  local shard_output_dir
  local selected_ids
  local worker_offset
  local -a run_command
  local status

  grep="$(shard_grep "$shard")"
  files="$(shard_files "$shard")"
  worker_offset="$(resolve_functional_worker_offset "$shard_index")"
  selected_ids="$("$node_bin" - "$shard_plan" "$shard" <<'EOF'
const fs = require("node:fs");
const [planPath, shardName] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planPath, "utf8"));
const shard = (plan.shards ?? []).find((entry) => entry.name === shardName);
if (!shard) {
  throw new Error(`missing shard ${shardName}`);
}
process.stdout.write(`${(shard.entries ?? []).map((entry) => entry.id).join("\n")}\n`);
EOF
  )"
  shard_report="${batch_dir}/${shard}.json"
  shard_stdout="${batch_dir}/${shard}.stdout.log"
  shard_stderr="${batch_dir}/${shard}.stderr.log"
  shard_output_dir="${output_dir}/${shard}"
  if [[ "$output_mode" == "quiet" ]]; then
    run_command=("${command[@]}" --reporter=json --output "$shard_output_dir" --project functional)
  else
    run_command=("${command[@]}" "--reporter=dot,json" --output "$shard_output_dir" --project functional)
  fi

  set +e
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP="$grep" \
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES="$files" \
  CARTULARY_PLAYWRIGHT_SUPPORT_GREP="$all_support_grep" \
  CARTULARY_PLAYWRIGHT_SUPPORT_FILES="$all_support_files" \
  CARTULARY_PLAYWRIGHT_WORKER_COUNT="$playwright_worker_count" \
  CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET="$worker_offset" \
  CARTULARY_MANIFEST_SELECTED_IDS="$selected_ids" \
  PLAYWRIGHT_WORKERS=1 \
  PLAYWRIGHT_JSON_OUTPUT_FILE="$shard_report" \
    "${run_command[@]}" >"$shard_stdout" 2>"$shard_stderr"
  status=$?
  set -e

  write_missing_report "$shard_report" "playwright shard ${shard} did not produce a JSON report"
  return "$status"
}

functional_status=0
if [[ "$mode" != "support" ]]; then
  active_shards=0
  for shard_index in "${!shard_names[@]}"; do
    actual_shard_index="$shard_index"
    if [[ "$mode" == "functional-shard" ]]; then
      actual_shard_index="$single_shard_index"
    fi
    run_functional_shard "${shard_names[$shard_index]}" "$actual_shard_index" &
    active_shards=$((active_shards + 1))
    if (( active_shards >= functional_shard_limit )); then
      if ! wait -n; then
        functional_status=1
      fi
      active_shards=$((active_shards - 1))
    fi
  done
  while (( active_shards > 0 )); do
    if ! wait -n; then
      functional_status=1
    fi
    active_shards=$((active_shards - 1))
  done

  if compgen -G "${batch_dir}/browser-functional-shard-*.stdout.log" >/dev/null; then
    cat "${batch_dir}"/browser-functional-shard-*.stdout.log >"$stdout_log"
  else
    : >"$stdout_log"
  fi
  if compgen -G "${batch_dir}/browser-functional-shard-*.stderr.log" >/dev/null; then
    cat "${batch_dir}"/browser-functional-shard-*.stderr.log >"$stderr_log"
  else
    : >"$stderr_log"
  fi
  if [[ "$output_mode" != "quiet" ]]; then
    cat "$stdout_log"
    cat "$stderr_log" >&2
  fi

  "$node_bin" "$shard_plan_script" merge-reports "$run_report" "${batch_dir}"/browser-functional-shard-*.json
else
  : >"$stdout_log"
  : >"$stderr_log"
fi

support_status=0
if [[ ("$mode" == "webserver-backed" || "$mode" == "support") && "${#support_selection_specs[@]}" -gt 0 ]]; then
  support_run_command=("${command[@]}" --reporter=json --output "${output_dir}/support" --project support)
  if [[ "$output_mode" != "quiet" ]]; then
    support_run_command=("${command[@]}" "--reporter=dot,json" --output "${output_dir}/support" --project support)
  fi

  set +e
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP="$all_functional_grep" \
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES="$all_functional_files" \
  CARTULARY_PLAYWRIGHT_SUPPORT_GREP="$all_support_grep" \
  CARTULARY_PLAYWRIGHT_SUPPORT_FILES="$all_support_files" \
  PLAYWRIGHT_JSON_OUTPUT_FILE="$support_report" \
    "${support_run_command[@]}" >"$support_stdout_log" 2>"$support_stderr_log"
  support_status=$?
  set -e
  write_missing_report "$support_report" "playwright support project did not produce a JSON report"
  if [[ "$output_mode" != "quiet" ]]; then
    cat "$support_stdout_log"
    cat "$support_stderr_log" >&2
  fi
fi

phase_capture_finish BATCH
start_time="${BATCH_START_TIME}"
end_time="${BATCH_END_TIME}"
duration_ms="${BATCH_DURATION_MS}"
command_text="$(render_command "${command[@]}" --config-shards "$shard_plan")"

emit_playwright_manifest_slice() {
  local label="$1"
  local phase="$2"
  local accounting_mode="$3"
  local logical_duration_ms="$4"
  local executed_duration_ms="$5"
  local wall_duration_ms="$6"
  local selected_ids="$7"
  local phase_dir
  local selection_report
  local helper_status

  phase_dir="$(prepare_phase_artifact_dir "$label")"
  selection_report="${phase_dir}/manifest-selected-tests.json"
  "$node_bin" "$manifest_script" playwright-selection-report "$phase" authoritative browser_functional >"$selection_report"

  set +e
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_PHASE_ACCOUNTING_MODE="$accounting_mode" \
  CARTULARY_PHASE_LABEL="$label" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$logical_duration_ms" \
  CARTULARY_PHASE_LOGICAL_DURATION_MS="$logical_duration_ms" \
  CARTULARY_PHASE_EXECUTED_DURATION_MS="$executed_duration_ms" \
  CARTULARY_PHASE_WALL_DURATION_MS="$wall_duration_ms" \
  CARTULARY_PHASE_EXIT_STATUS="$functional_status" \
  CARTULARY_PHASE_RUNNER_LOG="$run_report" \
  CARTULARY_PLAYWRIGHT_SELECTION_REPORT="$selection_report" \
  CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
  CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
  CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
  CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
  CARTULARY_MANIFEST_PHASE="$phase" \
  CARTULARY_MANIFEST_COVERAGE=authoritative \
  CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=browser_functional \
  CARTULARY_MANIFEST_SELECTED_IDS="$selected_ids" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-manifest-phase
  helper_status=$?
  set -e
  return "$helper_status"
}

emit_playwright_support_slice() {
  local phase="$1"
  local label="browser-e2e-support ${phase} supplemental"
  local phase_dir
  local selection_report
  local helper_status

  phase_dir="$(prepare_phase_artifact_dir "$label")"
  selection_report="${phase_dir}/manifest-selected-tests.json"
  "$node_bin" "$manifest_script" playwright-selection-report "$phase" supplemental browser_support >"$selection_report"

  set +e
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_PHASE_ACCOUNTING_MODE=derived \
  CARTULARY_PHASE_LABEL="$label" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS=0 \
  CARTULARY_PHASE_LOGICAL_DURATION_MS=0 \
  CARTULARY_PHASE_EXECUTED_DURATION_MS=0 \
  CARTULARY_PHASE_WALL_DURATION_MS=0 \
  CARTULARY_PHASE_EXIT_STATUS="$support_status" \
  CARTULARY_PHASE_RUNNER_LOG="$support_report" \
  CARTULARY_PLAYWRIGHT_SELECTION_REPORT="$selection_report" \
  CARTULARY_PHASE_STDOUT_LOG="$support_stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$support_stderr_log" \
  CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
  CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
  CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
  CARTULARY_MANIFEST_PHASE="$phase" \
  CARTULARY_MANIFEST_COVERAGE=supplemental \
  CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=browser_support \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-manifest-phase
  helper_status=$?
  set -e
  return "$helper_status"
}

overall_status=0
for phase in "${functional_phases[@]}"; do
  if [[ "$mode" == "support" ]]; then
    continue
  fi
  label="browser-e2e-functional ${phase} authoritative"
  if [[ "$mode" == "functional-shard" ]]; then
    label="${label} ${single_shard_name}"
  fi
  selected_phase_ids="$("$node_bin" - "$shard_plan" "$phase" "$single_shard_name" <<'EOF'
const fs = require("node:fs");
const [planPath, phase, selectedShardName] = process.argv.slice(2);
const plan = JSON.parse(fs.readFileSync(planPath, "utf8"));
const ids = [];
for (const shard of plan.shards ?? []) {
  if (selectedShardName && shard.name !== selectedShardName) {
    continue;
  }
  for (const entry of shard.entries ?? []) {
    if (entry.phase === phase) {
      ids.push(entry.id);
    }
  }
}
process.stdout.write(`${ids.join("\n")}\n`);
EOF
  )"
  accounting_mode=actual
  logical_ms="$duration_ms"
  executed_ms="$duration_ms"
  wall_ms="${BATCH_WALL_DURATION_MS}"
  if ! emit_playwright_manifest_slice "$label" "$phase" "$accounting_mode" "$logical_ms" "$executed_ms" "$wall_ms" "$selected_phase_ids"; then
    overall_status=1
  fi
done

if [[ "$mode" == "webserver-backed" || "$mode" == "support" ]]; then
  for phase in "${support_phases[@]}"; do
    if ! emit_playwright_support_slice "$phase"; then
      overall_status=1
    fi
  done
fi

if [[ "$functional_status" -ne 0 && "$overall_status" -eq 0 ]]; then
  overall_status="$functional_status"
fi
if [[ "$support_status" -ne 0 && "$overall_status" -eq 0 ]]; then
  overall_status="$support_status"
fi

if [[ "$overall_status" -eq 0 ]]; then
  exit 0
fi

emit_target_summary fail || true
exit "$overall_status"
