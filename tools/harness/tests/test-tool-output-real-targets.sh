#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
MAKE_BIN="${MAKE_BIN:-make}"
NODE_BIN="${NODE_BIN:-node}"
TMP_DIR="$(mktemp -d "${ROOT_DIR}/tmp/tool-output-real-targets.XXXXXX")"
RUN_PREFIX="$(printf 'r%x' "$$")"
RESULTS_ROOT="tmp/op"
preserve_results=0

cleanup() {
  rm -rf "$TMP_DIR"
  if [[ "$preserve_results" -eq 0 ]]; then
    rm -rf "$ROOT_DIR/${RESULTS_ROOT}/${RUN_PREFIX}-"*
  else
    echo "retained failing real-target roots at $ROOT_DIR/${RESULTS_ROOT}/${RUN_PREFIX}-*" >&2
  fi
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
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u PLAYWRIGHT_WORKERS -u JSON \
      -u MAKEFLAGS -u MFLAGS \
      CARTULARY_OUTPUT_MODE=summary \
      CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
      CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target" "${make_args[@]}"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log" || {
    local status=$?
    preserve_results=1
    echo "${target}: failed with status ${status}" >&2
    sed -n '1,80p' "${target_dir}/stdout.log" >&2
    sed -n '1,80p' "${target_dir}/stderr.log" >&2
    return "$status"
  }
  # Preserve a successful target's canonical root if post-run contract
  # validation fails; otherwise the diagnostic would erase the evidence it
  # needs to explain.
  preserve_results=1
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
if (target.output_policy.summary_schema === "cartulary.harness_run_summary.v1") {
  const graphLines = stdout
    .split(/\r?\n/)
    .filter((line) => line.startsWith("[GRAPH] "))
    .filter((line) => lineField(line, "target") === targetName && lineField(line, "status") === "pass");
  if (graphLines.length !== 1) {
    throw new Error(`${targetName}: expected one passing [GRAPH] line`);
  }
  const runRoot = lineField(graphLines[0], "run_root");
  if (!runRoot) throw new Error(`${targetName}: graph output omitted run_root`);
  const runRootPath = resolveRepoPath(runRoot);
  const files = {
    manifest: path.join(runRootPath, "run-manifest.json"),
    events: path.join(runRootPath, "unit-events.ndjson"),
    run: path.join(runRootPath, "run-summary.json"),
    target: path.join(runRootPath, "target-summaries", `${targetName}.json`),
  };
  for (const [role, file] of Object.entries(files)) {
    if (!fs.existsSync(file)) throw new Error(`${targetName}: missing canonical ${role} artifact`);
  }
  const runSummary = JSON.parse(fs.readFileSync(files.run, "utf8"));
  const targetSummary = JSON.parse(fs.readFileSync(files.target, "utf8"));
  const runManifest = JSON.parse(fs.readFileSync(files.manifest, "utf8"));
  if (
    runSummary.schema_id !== "cartulary.harness_run_summary.v1" ||
    runSummary.target !== targetName ||
    runSummary.status !== "pass" ||
    targetSummary.schema_id !== "cartulary.harness_target_summary.v1" ||
    targetSummary.target !== targetName ||
    targetSummary.status !== "pass" ||
    runManifest.schema_id !== "cartulary.harness_run_manifest.v1" ||
    runManifest.target !== targetName
  ) {
    throw new Error(`${targetName}: invalid canonical graph summaries ${JSON.stringify({
      run_summary: { schema_id: runSummary.schema_id, target: runSummary.target, status: runSummary.status },
      target_summary: { schema_id: targetSummary.schema_id, target: targetSummary.target, status: targetSummary.status },
      run_manifest: { schema_id: runManifest.schema_id, target: runManifest.target },
    })}`);
  }
  const eventLines = fs.readFileSync(files.events, "utf8").trim().split(/\r?\n/).filter(Boolean);
  if (eventLines.length === 0 || eventLines.some((line) => JSON.parse(line).schema_id !== "cartulary.harness_unit_event.v1")) {
    throw new Error(`${targetName}: invalid canonical unit event stream`);
  }
  process.exit(0);
}
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
const addSummaryRefs = (summaryPath, runRootPath) => {
  if (seenSummaries.has(summaryPath) || !fs.existsSync(summaryPath)) {
    return;
  }
  seenSummaries.add(summaryPath);
  const summary = JSON.parse(fs.readFileSync(summaryPath, "utf8"));
  if (summary.schema_id !== "cartulary.tool_run_summary.v5") {
    throw new Error(`${targetName}: unexpected schema ${summary.schema_id}`);
  }
  if (summary.target !== targetName || summary.status !== "pass") {
    throw new Error(`${targetName}: unexpected summary target/status`);
  }
  primarySummary ??= summary;
  refs.push(
    ...(summary.summary_artifacts ?? []).map((artifact) => ({ ...artifact, runRootPath })),
    ...(summary.log_artifacts ?? []).map((artifact) => ({ ...artifact, runRootPath })),
  );
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
  addSummaryRefs(summaryPath, runRootPath);
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
  const artifactPath = path.resolve(artifact.runRootPath, artifact.path);
  const relativeArtifact = path.relative(artifact.runRootPath, artifactPath).replaceAll("\\", "/");
  if (relativeArtifact.startsWith("../") || relativeArtifact === ".." || path.isAbsolute(relativeArtifact)) {
    throw new Error(`${targetName}: artifact ${artifact.role} escapes run root at ${artifact.path}`);
  }
  if (!fs.existsSync(artifactPath)) {
    throw new Error(`${targetName}: artifact ${artifact.role} missing at ${artifact.path}`);
  }
}

EOF
  preserve_results=0
}

run_machine_target() {
  local target="$1"
  local label="$2"
  local target_dir="${TMP_DIR}/${label}"
  mkdir -p "$target_dir"
  (
    cd "$ROOT_DIR"
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u PLAYWRIGHT_WORKERS -u JSON \
      -u MAKEFLAGS -u MFLAGS \
      CARTULARY_OUTPUT_MODE=machine \
      CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
      CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log" || {
    local status=$?
    preserve_results=1
    echo "${target}: machine target failed with status ${status}" >&2
    sed -n '1,80p' "${target_dir}/stdout.log" >&2
    sed -n '1,80p' "${target_dir}/stderr.log" >&2
    return "$status"
  }
  "$NODE_BIN" - "$target" "${target_dir}/stdout.log" "${target_dir}/stderr.log" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
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
const manifest = JSON.parse(fs.readFileSync(path.join(process.cwd(), "tools/task_surface_manifest.json"), "utf8"));
const target = manifest.targets.find((entry) => entry.name === targetName);
const expectedSchema = target?.output_policy?.summary_schema ?? "cartulary.tool_run_summary.v5";
if (summary.schema_id !== expectedSchema || summary.target !== targetName || summary.status !== "pass") {
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
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u PLAYWRIGHT_WORKERS -u JSON \
      -u MAKEFLAGS -u MFLAGS \
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
if (!stderr.includes("failure_reason=usage_error")) {
  throw new Error("invalid target output must be classified as a usage error");
}
if (!stderr.includes("TARGET must name a declared target")) {
  throw new Error("invalid target output must identify the target-selection error");
}
EOF
}

run_target json-shape-check json-shape-check tool_run_summary target_summary target_timing
# lint-shell is covered by harness-smoke-lint-shell; in summary mode its success
# output is retained in artifacts and does not emit a standalone [RESULT] line.
run_target backend-unit backend-unit
run_machine_target backend-unit backend-unit-machine
run_target lint lint tool_run_summary run_summary
run_target build build
run_machine_target build build-machine
run_target test-fast test-fast
run_target check check
run_machine_target check check-machine
run_invalid_usage_check

echo "real target output policy checks passed"
