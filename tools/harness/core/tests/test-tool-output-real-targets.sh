#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
MAKE_BIN="${MAKE_BIN:-make}"
NODE_BIN="${NODE_BIN:-node}"
TMP_DIR="$(mktemp -d "${ROOT_DIR}/tmp/tool-output-real-targets.XXXXXX")"
RUN_PREFIX="opr$RANDOM$RANDOM"
RESULTS_ROOT="tmp/op"

cleanup() {
  rm -rf "$TMP_DIR"
  rm -rf "$ROOT_DIR/${RESULTS_ROOT}/${RUN_PREFIX}-"*
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

run_target() {
  local target="$1"
  local label="$2"
  shift 2
  local target_dir="${TMP_DIR}/${label}"
  local make_args=()
  if [[ -n "${RUN_TARGET_MAKE_ARGS:-}" ]]; then
    read -r -a make_args <<<"${RUN_TARGET_MAKE_ARGS}"
  fi
  mkdir -p "$target_dir"
  (
    cd "$ROOT_DIR"
    CARTULARY_OUTPUT_MODE=summary \
    CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
    CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target" "${make_args[@]}"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log" || {
    local status=$?
    echo "${target}: failed with status ${status}" >&2
    sed -n '1,80p' "${target_dir}/stdout.log" >&2
    sed -n '1,80p' "${target_dir}/stderr.log" >&2
    return "$status"
  }
  "$NODE_BIN" - "$ROOT_DIR" "$target" "${target_dir}/stdout.log" "${target_dir}/stderr.log" "$@" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [repoRoot, targetName, stdoutFile, stderrFile, ...requiredRoles] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(path.join(repoRoot, "tools/task_surface_manifest.json"), "utf8"));
const target = manifest.targets.find((entry) => entry.name === targetName);
if (!target?.output_policy?.success_budget) {
  throw new Error(`${targetName}: missing success budget`);
}
const budget = target.output_policy.success_budget;
const readText = (file) => fs.existsSync(file) ? fs.readFileSync(file, "utf8") : "";
const stdout = readText(stdoutFile);
const stderr = readText(stderrFile);
const lineCount = (text) => {
  if (text.length === 0) return 0;
  const trimmed = text.endsWith("\n") ? text.slice(0, -1) : text;
  return trimmed.length === 0 ? 0 : trimmed.split(/\r?\n/).length;
};
for (const [key, actual] of [
  ["stdout_lines", lineCount(stdout)],
  ["stdout_bytes", Buffer.byteLength(stdout)],
  ["stderr_lines", lineCount(stderr)],
  ["stderr_bytes", Buffer.byteLength(stderr)],
]) {
  const limit = budget[key];
  if (Number.isInteger(limit) && actual > limit) {
    throw new Error(`${targetName}: ${key} ${actual} exceeds budget ${limit}`);
  }
}
if (stderr !== "") {
  throw new Error(`${targetName}: expected empty successful stderr`);
}
const lineField = (line, name) => {
  const match = line.match(new RegExp(`${name}=([^ ]+)`));
  return match ? match[1] : null;
};
const resolveRepoPath = (artifactPath) => {
  if (path.isAbsolute(artifactPath)) {
    return artifactPath;
  }
  return path.resolve(repoRoot, artifactPath);
};
const resultLines = stdout
  .split(/\r?\n/)
  .filter((line) => line.startsWith("[RESULT] "))
  .filter((line) => lineField(line, "target") === targetName && lineField(line, "status") === "pass");
if (resultLines.length === 0) {
  throw new Error(`${targetName}: missing passing [RESULT] line`);
}
const refs = [];
let primarySummary = null;
const seenSummaries = new Set();
const addSummaryRefs = (summaryPath) => {
  if (seenSummaries.has(summaryPath) || !fs.existsSync(summaryPath)) {
    return;
  }
  seenSummaries.add(summaryPath);
  const summary = JSON.parse(fs.readFileSync(summaryPath, "utf8"));
  if (summary.schema_id !== "cartulary.tool_run_summary.v3") {
    throw new Error(`${targetName}: unexpected schema ${summary.schema_id}`);
  }
  if (summary.target !== targetName || summary.status !== "pass") {
    throw new Error(`${targetName}: unexpected summary target/status`);
  }
  primarySummary ??= summary;
  refs.push(...(summary.summary_artifacts ?? []), ...(summary.log_artifacts ?? []));
};
for (const resultLine of resultLines) {
  const runRoot = lineField(resultLine, "run_root");
  const summaryJson = lineField(resultLine, "summary_json");
  if (!runRoot || !summaryJson) {
    throw new Error(`${targetName}: missing run_root or summary_json in ${resultLine}`);
  }
  const runRootPath = resolveRepoPath(runRoot);
  const allowsRunLevelSummary =
    ["run_and_target_summaries", "scheduler_and_tool_run_summaries"].includes(
      target.output_policy?.artifact_policy,
    );
  if (runRoot.endsWith(`/${targetName}`)) {
    throw new Error(`${targetName}: run_root must identify the run, not the target directory: ${runRoot}`);
  }
  const summaryPath = path.isAbsolute(summaryJson)
    ? summaryJson
    : path.resolve(runRootPath, summaryJson);
  const relativeSummary = path.relative(runRootPath, summaryPath).replaceAll("\\", "/");
  if (relativeSummary.startsWith("../") || relativeSummary === ".." || path.isAbsolute(relativeSummary)) {
    throw new Error(`${targetName}: summary_json must stay under run_root, got ${summaryJson}`);
  }
  const canonicalTargetSummaryPath = path.resolve(runRootPath, targetName, "tool-run-summary.json");
  const canonicalRunSummaryPath = path.resolve(runRootPath, "tool-run-summary.json");
  if (!allowsRunLevelSummary && summaryPath !== canonicalTargetSummaryPath) {
    throw new Error(`${targetName}: expected canonical target-level summary ${canonicalTargetSummaryPath}, got ${summaryPath}`);
  }
  if (
    allowsRunLevelSummary &&
    summaryPath !== canonicalTargetSummaryPath &&
    summaryPath !== canonicalRunSummaryPath
  ) {
    throw new Error(`${targetName}: expected target or run summary under ${runRootPath}, got ${summaryPath}`);
  }
  if (!fs.existsSync(summaryPath)) {
    throw new Error(`${targetName}: missing tool-run summary ${summaryPath}`);
  }
  addSummaryRefs(summaryPath);
}
const roles = new Set(refs.map((artifact) => artifact.role));
for (const role of requiredRoles) {
  if (!roles.has(role)) {
    throw new Error(`${targetName}: missing artifact role ${role}`);
  }
}
for (const artifact of refs) {
  if (!artifact.path) {
    throw new Error(`${targetName}: artifact ${artifact.role} has no path`);
  }
  const artifactPath = resolveRepoPath(artifact.path);
  if (!fs.existsSync(artifactPath)) {
    throw new Error(`${targetName}: artifact ${artifact.role} missing at ${artifact.path}`);
  }
}
if (["phase-slice", "service-backed-slice"].includes(targetName)) {
  if (!primarySummary) {
    throw new Error(`${targetName}: missing primary summary`);
  }
  const totalWork = (primarySummary.work_units ?? []).reduce((sum, unit) => sum + (unit.total ?? 0), 0);
  const totalCount =
    (primarySummary.counts?.tests ?? 0) +
    (primarySummary.counts?.non_test ?? 0) +
    (primarySummary.counts?.packages ?? 0);
  if (totalWork <= 0) {
    throw new Error(`${targetName}: expected nonzero work unit data`);
  }
  if (totalCount <= 0) {
    throw new Error(`${targetName}: expected nonzero count data`);
  }
}
EOF
}

run_machine_target() {
  local target="$1"
  local label="$2"
  local target_dir="${TMP_DIR}/${label}"
  mkdir -p "$target_dir"
  (
    cd "$ROOT_DIR"
    CARTULARY_OUTPUT_MODE=machine \
    CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
    CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log"
  "$NODE_BIN" - "$target" "${target_dir}/stdout.log" "${target_dir}/stderr.log" <<'EOF'
const fs = require("node:fs");
const [targetName, stdoutFile, stderrFile] = process.argv.slice(2);
const stdout = fs.readFileSync(stdoutFile, "utf8");
const stderr = fs.readFileSync(stderrFile, "utf8");
if (stderr !== "") {
  throw new Error(`${targetName}: machine stderr must be empty`);
}
const lines = stdout.split(/\r?\n/).filter(Boolean);
if (lines.length !== 1) {
  throw new Error(`${targetName}: machine mode must emit one JSON line, got ${lines.length}`);
}
const summary = JSON.parse(lines[0]);
if (
  summary.schema_id !== "cartulary.tool_run_summary.v3" ||
  summary.target !== targetName ||
  summary.status !== "pass"
) {
  throw new Error(`${targetName}: unexpected machine summary`);
}
if (stdout.includes("[RESULT]") || stdout.includes("[ARTIFACTS]") || stdout.includes("[PROGRESS]")) {
  throw new Error(`${targetName}: machine mode must not include human lines`);
}
EOF
}

run_invalid_usage_check() {
  local target_dir="${TMP_DIR}/invalid-explain-target"
  mkdir -p "$target_dir"
  set +e
  (
    cd "$ROOT_DIR"
    "$MAKE_BIN" --no-print-directory explain-target TARGET=not-a-real-target DETAIL=summary
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log"
  local status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "invalid explain-target unexpectedly passed"
  fi
  "$NODE_BIN" - "${target_dir}/stderr.log" <<'EOF'
const fs = require("node:fs");
const [stderrFile] = process.argv.slice(2);
const stderr = fs.readFileSync(stderrFile, "utf8");
if (!stderr.includes('expected="TARGET=<target> [DETAIL=summary|rows|artifacts]"')) {
  throw new Error("invalid target output must include expected argument shape");
}
const match = stderr.match(/nearest=([^\n]+)/);
if (!match) {
  throw new Error("invalid target output must include nearest candidates");
}
const candidates = match[1].split(",").filter(Boolean);
if (candidates.length > 10) {
  throw new Error(`invalid target output returned too many candidates: ${candidates.length}`);
}
EOF
}

run_target json-shape-check json-shape-check tool_run_summary phase_summary
run_target lint-shell lint-shell tool_run_summary phase_summary shellcheck_inventory
run_target backend-unit backend-unit tool_run_summary target_summary target_timing
run_machine_target backend-unit backend-unit-machine
run_target lint lint tool_run_summary target_summary target_timing
run_target build build tool_run_summary target_summary target_timing
RUN_TARGET_MAKE_ARGS="PHASE=phase0" run_target phase-slice phase-slice-phase0 tool_run_summary target_summary target_timing scheduler_summary scheduler_events scheduler_progress scheduler_logs
RUN_TARGET_MAKE_ARGS="PHASE=phase0" run_target service-backed-slice service-backed-slice-phase0 tool_run_summary target_summary target_timing scheduler_summary scheduler_events scheduler_progress scheduler_logs
run_machine_target build build-machine
run_target test-fast test-fast tool_run_summary run_summary
run_target check check tool_run_summary target_summary scheduler_summary scheduler_events scheduler_progress scheduler_logs
run_machine_target check check-machine
run_invalid_usage_check

echo "real target output policy checks passed"
