#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { failureHeadlineForSummary } from "../contract/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const validDetails = new Set(["summary", "children", "logs", "progress", "accounting"]);
const coverageBuckets = [
  "authoritative",
  "support",
  "raw",
  "tooling_support",
  "unowned_regression",
  "unmapped",
];

function usage() {
  process.stderr.write(
    "usage: print-explain-run.mjs --results-dir <root|run-dir> [--run-id <id>] [--target <target>] [--detail summary|children|logs|progress|accounting]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = {
    resultsDir: "",
    runId: "",
    target: "",
    detail: "summary",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--results-dir") {
      options.resultsDir = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--run-id") {
      options.runId = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--target") {
      options.target = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--detail") {
      options.detail = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.resultsDir || !validDetails.has(options.detail)) {
    usage();
  }
  if ((options.detail === "logs" || options.detail === "progress") && !options.target) {
    throw new Error(`DETAIL=${options.detail} requires TARGET=<target>`);
  }
  return options;
}

function resolvePath(value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function relToRepo(value) {
  const relative = path.relative(repoRoot, value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return value.replaceAll("\\", "/");
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function toolSummaryTargets(runDir) {
  if (!existsSync(runDir)) {
    return [];
  }
  return readdirSync(runDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .filter((target) => existsSync(path.join(runDir, target, "tool-run-summary.json")))
    .sort((left, right) => left.localeCompare(right));
}

function hasRunArtifacts(dir) {
  if (existsSync(path.join(dir, "run-summary.json"))) {
    return true;
  }
  return toolSummaryTargets(dir).length > 0;
}

function newestRunDir(resultsRoot) {
  const candidates = readdirSync(resultsRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(resultsRoot, entry.name))
    .filter((dir) => hasRunArtifacts(dir))
    .sort((left, right) => statSync(right).mtimeMs - statSync(left).mtimeMs);
  return candidates[0] ?? "";
}

function defaultToolTarget(runDir) {
  const targets = toolSummaryTargets(runDir);
  if (targets.includes("agent-finalize")) {
    return "agent-finalize";
  }
  return targets.length === 1 ? targets[0] : "";
}

function resolveRunContext(options) {
  const resultsDir = resolvePath(options.resultsDir);
  if (existsSync(path.join(resultsDir, "run-summary.json"))) {
    return { runDir: resultsDir, targetFromPath: "" };
  }
  if (existsSync(path.join(resultsDir, "tool-run-summary.json"))) {
    const toolSummary = readJSON(path.join(resultsDir, "tool-run-summary.json"));
    return { runDir: path.dirname(resultsDir), targetFromPath: toolSummary.target ?? path.basename(resultsDir) };
  }
  if (options.target && existsSync(path.join(resultsDir, options.target, "target-summary.json"))) {
    return { runDir: resultsDir, targetFromPath: "" };
  }
  if (options.target && existsSync(path.join(resultsDir, options.target, "tool-run-summary.json"))) {
    return { runDir: resultsDir, targetFromPath: options.target };
  }
  if (toolSummaryTargets(resultsDir).length > 0) {
    return { runDir: resultsDir, targetFromPath: "" };
  }
  if (options.runId) {
    return { runDir: path.join(resultsDir, options.runId), targetFromPath: "" };
  }
  const newest = newestRunDir(resultsDir);
  if (newest) {
    return { runDir: newest, targetFromPath: "" };
  }
  throw new Error(`no run-summary.json or target/tool-run-summary.json found under ${resultsDir}; pass RUN_ID for a results root`);
}

function formatDuration(ms) {
  if (!Number.isFinite(ms) || ms <= 0) {
    return "0ms";
  }
  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  }
  return `${(ms / 1000).toFixed(2)}s`;
}

function counts(summary) {
  return summary?.counts ?? {};
}

function coverageCounts(summary) {
  const c = counts(summary);
  return `authoritative=${c.authoritative ?? 0} support=${c.support ?? 0} raw=${c.raw ?? 0} tooling_support=${c.tooling_support ?? 0} unowned_regression=${c.unowned_regression ?? 0} unmapped=${c.unmapped ?? 0}`;
}

function duration(summary) {
  return formatDuration(
    summary?.critical_path_wall_duration_ms ?? summary?.wall_duration_ms ?? summary?.logical_duration_ms ?? 0,
  );
}

function slowestTarget(summary) {
  const slowest = summary?.slowest_target;
  if (!slowest) {
    return "none";
  }
  return `${slowest.target}(${formatDuration(slowest.critical_path_wall_duration_ms)})`;
}

function slowestChild(targetSummary) {
  const children = targetSummary?.children?.present ?? [];
  let slowest = null;
  for (const child of children) {
    const totals = child.totals ?? child;
    const value =
      totals.critical_path_wall_duration_ms ?? totals.wall_duration_ms ?? totals.logical_duration_ms ?? 0;
    if (!slowest || value > slowest.duration_ms || (value === slowest.duration_ms && child.target < slowest.target)) {
      slowest = { target: child.target, duration_ms: value };
    }
  }
  return slowest ? `${slowest.target}(${formatDuration(slowest.duration_ms)})` : "none";
}

function failureClassField(summary) {
  return summary?.failure_class ? ` failure_class=${summary.failure_class}` : "";
}

function writeFailureHeadline(label, summary) {
  const headline = failureHeadlineForSummary(summary);
  if (headline) {
    process.stdout.write(`[FAILURE] ${label} ${headline}\n`);
  }
}

function failureLabels(summary) {
  return (summary?.failures ?? [])
    .map((failure) => failure.label || failure.row_id || failure.target || "")
    .filter(Boolean);
}

function loadRunSummary(runDir) {
  const file = path.join(runDir, "run-summary.json");
  return existsSync(file) ? readJSON(file) : null;
}

function loadTargetSummary(runDir, target) {
  if (!target) {
    return null;
  }
  const file = path.join(runDir, target, "target-summary.json");
  return existsSync(file) ? readJSON(file) : null;
}

function loadSchedulerSummary(runDir, target) {
  if (!target) {
    return null;
  }
  const file = path.join(runDir, target, "scheduler-summary.json");
  return existsSync(file) ? readJSON(file) : null;
}

function loadToolSummary(runDir, target) {
  if (!target) {
    return null;
  }
  const file = path.join(runDir, target, "tool-run-summary.json");
  return existsSync(file) ? readJSON(file) : null;
}

function artifactPathByRole(summary, role) {
  return (summary?.summary_artifacts ?? []).find((artifact) => artifact.role === role)?.path ?? "";
}

function loadFinalizeSummary(runDir, toolSummary) {
  const configured = artifactPathByRole(toolSummary, "finalize_summary");
  const file = configured ? absoluteArtifactPath(configured) : path.join(runDir, "agent-finalize", "finalize-summary.json");
  return existsSync(file) ? { file, summary: readJSON(file) } : { file, summary: null };
}

function writeBrowserStartupDiagnostics(toolSummary) {
  const configured = artifactPathByRole(toolSummary, "browser_startup_diagnostics");
  if (!configured) {
    return;
  }
  const file = absoluteArtifactPath(configured);
  if (!existsSync(file)) {
    process.stdout.write(`[BROWSER-STARTUP] missing artifacts=${configured}\n`);
    return;
  }
  const summary = readJSON(file);
  const failure = summary.failure_class
    ? ` failure_class=${summary.failure_class} reason=${summary.failure_reason ?? "unknown_failure"}`
    : "";
  process.stdout.write(
    `[BROWSER-STARTUP] status=${summary.status} phase=${summary.startup_phase} frontend_mode=${summary.frontend_mode} command_kind=${summary.frontend_command_kind}${failure} message=${summary.message ?? ""} artifacts=${configured}\n`,
  );
}

function writeRunSummary(runDir, runSummary) {
  if (!runSummary) {
    process.stdout.write(`[RUN] missing artifacts=${relToRepo(runDir)}\n`);
    return;
  }
  const c = counts(runSummary);
  const workUnits = runSummary.work_units ?? { completed: 0, total: 0 };
  const summaryTargets = runSummary.summary_targets ?? { expected: [], missing: [] };
  const evidenceTargets = runSummary.evidence_targets ?? { present: [] };
  const helperUnits = runSummary.helper_units ?? { total: 0 };
  const expectedEvidenceTargets = Math.max(
    0,
    summaryTargets.expected.length - (summaryTargets.skipped_after_failure?.length ?? 0),
  );
  process.stdout.write(
    `[RUN] ${runSummary.label} status=${runSummary.status}${failureClassField(runSummary)} work_units=${workUnits.completed}/${workUnits.total} summary_targets=${summaryTargets.expected.length} evidence_targets=${evidenceTargets.present.length}/${expectedEvidenceTargets} helper_units=${helperUnits.total} tests=${c.tests ?? 0} failed=${c.failed ?? 0} ${coverageCounts(runSummary)} duration=${duration(runSummary)} slowest_target=${slowestTarget(runSummary)} artifacts=${runSummary.artifacts?.dir ?? relToRepo(runDir)}\n`,
  );
  writeFailureHeadline(runSummary.label, runSummary);
  const missing = summaryTargets.missing ?? [];
  if (missing.length > 0) {
    process.stdout.write(`[RUN-MISSING] ${missing.join(",")}\n`);
  }
}

function writeToolSummary(runDir, target, toolSummary) {
  if (!toolSummary) {
    const targets = toolSummaryTargets(runDir);
    process.stdout.write(`[TOOL] missing target=${target || "none"} available=${targets.join(",") || "none"} artifacts=${relToRepo(runDir)}\n`);
    return;
  }
  const c = counts(toolSummary);
  process.stdout.write(
    `[TOOL] ${toolSummary.target} status=${toolSummary.status}${failureClassField(toolSummary)} exit_code=${toolSummary.exit_code} tests=${c.tests ?? 0} failed=${c.failed ?? 0} duration=${formatDuration(toolSummary.duration_ms ?? 0)} output_mode=${toolSummary.output_mode} summaries=${toolSummary.summary_artifacts?.length ?? 0} logs=${toolSummary.log_artifacts?.length ?? 0} artifacts=${toolSummary.run_root ?? relToRepo(runDir)}\n`,
  );
  writeFailureHeadline(toolSummary.target, toolSummary);
  writeBrowserStartupDiagnostics(toolSummary);
}

function writeFinalizeSummary(runDir, toolSummary) {
  const { file, summary } = loadFinalizeSummary(runDir, toolSummary);
  if (!summary) {
    process.stdout.write(`[FINALIZE] missing artifacts=${relToRepo(file)}\n`);
    return;
  }
  const actions = summary.actions ?? [];
  process.stdout.write(
    `[FINALIZE] ${summary.target} status=${summary.status} results_dir_status=${summary.results_dir_status} generated=${summary.generated?.status ?? "unknown"} updated_files=${summary.generated?.updated_file_count ?? 0} duration=${summary.duration?.status ?? "unknown"} run_checks=${summary.run_checks?.status ?? "unknown"} actions=${actions.length} failures=${summary.failures?.length ?? 0} artifacts=${relToRepo(file)}\n`,
  );
  for (const action of actions) {
    const substeps = action.substeps ?? [];
    const failed = substeps.filter((substep) => substep.status === "fail").map((substep) => substep.id);
    const skipped = substeps.filter((substep) => substep.status === "skipped").map((substep) => substep.id);
    process.stdout.write(
      `[FINALIZE-ACTION] ${action.action_id} status=${action.status} execution_state=${action.execution_state ?? "unknown"} cache_state=${action.cache?.state ?? "none"} cache_reason=${action.cache?.reason_code ?? "none"} substeps=${substeps.length} failed=${failed.join(",") || "none"} skipped=${skipped.join(",") || "none"} duration=${formatDuration(action.duration_ms ?? 0)}\n`,
    );
  }
  for (const failure of summary.failures ?? []) {
    process.stdout.write(
      `[FINALIZE-FAILURE] action=${failure.action_id} substep=${failure.substep_id ?? "none"} target=${failure.target ?? "none"} failure_class=${failure.failure_class} failure_reason=${failure.failure_reason} headline=${failure.headline} artifact=${failure.summary_json ?? "none"}\n`,
    );
  }
}

function writeTargetSummary(runDir, targetSummary) {
  if (!targetSummary) {
    return;
  }
  const totals = targetSummary.totals ?? targetSummary;
  const c = counts(totals);
  const children = targetSummary.children;
  const childFields = children
    ? ` children=${children.present?.length ?? 0}/${children.expected?.length ?? 0} failed_children=${(children.failed_targets ?? []).join(",") || "none"} missing_children=${(children.missing ?? []).join(",") || "none"} skipped_children=${(children.skipped ?? []).map((child) => child.target).join(",") || "none"} slowest_child=${slowestChild(targetSummary)}`
    : "";
  process.stdout.write(
    `[TARGET] ${targetSummary.target} status=${targetSummary.status}${failureClassField(targetSummary)} kind=${targetSummary.kind ?? "leaf"} tests=${c.tests ?? 0} failed=${c.failed ?? 0} ${coverageCounts(totals)} duration=${duration(totals)}${childFields} artifacts=${targetSummary.own?.artifacts?.dir ?? relToRepo(path.join(runDir, targetSummary.target))}\n`,
  );
  writeFailureHeadline(targetSummary.target, targetSummary);
  writeFrontendRowAccountingDigest(targetSummary);
}

function writeFrontendRowAccountingDigest(targetSummary) {
  const configured =
    targetSummary.artifacts?.frontend_row_accounting ??
    targetSummary.own?.artifacts?.frontend_row_accounting ??
    "";
  if (!configured) {
    return;
  }
  const file = absoluteArtifactPath(configured);
  if (!existsSync(file)) {
    process.stdout.write(
      `[FRONTEND-ROWS] missing target=${targetSummary.target} artifacts=${configured}\n`,
    );
    return;
  }
  const accounting = readJSON(file);
  const blockers = failureLabels(targetSummary).join(";") || "none";
  for (const row of accounting.row_results ?? []) {
    const passed = row.closing_scenario_titles ?? [];
    if (row.failure_reason !== "target_failed" || passed.length === 0) {
      continue;
    }
    process.stdout.write(
      `[FRONTEND-ROW-BLOCKED] target=${targetSummary.target} row=${row.row_id} phase=${row.phase_id} closure=${row.closure_status} reason=${row.failure_reason} passed_scenarios=${passed.length} blocker=${blockers} artifacts=${configured}\n`,
    );
  }
}

function writeSchedulerSummary(summary) {
  if (!summary) {
    return;
  }
  const slowest = (summary.slowest_work_units ?? [])
    .map((entry) => `${entry.label}(${formatDuration(entry.duration_ms)})`)
    .join(",") || "none";
  const terminal = summary.critical_path_terminal_unit?.label
    ? `${summary.critical_path_terminal_unit.label}(${formatDuration(summary.critical_path_terminal_unit.duration_ms ?? 0)})`
    : "none";
  const blockers = (summary.critical_path_blockers ?? summary.top_blockers ?? [])
    .slice(0, 3)
    .map((entry) => `${entry.kind}:${entry.name}(${entry.count})`)
    .join(",") || "none";
  process.stdout.write(
    `[SCHEDULER] ${summary.target} status=${summary.status}${failureClassField(summary)} completed_work_units=${summary.completed_work_units}/${summary.total_work_units} failed=${summary.failed_work_unit ?? "none"} critical_path=${formatDuration(summary.critical_path_wall_duration_ms ?? 0)} terminal=${terminal} blockers=${blockers} slowest=${slowest} logs=${summary.artifacts?.scheduler_logs_dir ?? ""} progress=${summary.artifacts?.progress_summary_log ?? ""}\n`,
  );
  writeFailureHeadline(summary.target, summary);
  writeSchedulerProgressDigest(summary);
}

function formatSlowestObservation(entry) {
  const scope = entry.source === "nested"
    ? `${entry.work_unit || entry.nested_target || "nested"}`
    : entry.source || "outer";
  return `${scope}:${entry.label}(${formatDuration(entry.duration_ms)})`;
}

function writeSchedulerProgressDigest(summary) {
  const progressLog = summary.artifacts?.progress_summary_log ?? "";
  if (progressLog) {
    process.stdout.write(`[PROGRESS-LOG] ${summary.target} ${progressLog}\n`);
  }
  const snapshots = (summary.progress_snapshots ?? []).slice(-3);
  for (const snapshot of snapshots) {
    if (snapshot.line) {
      process.stdout.write(`[PROGRESS-SNAPSHOT] ${snapshot.line}\n`);
      continue;
    }
    process.stdout.write(
      `[PROGRESS-SNAPSHOT] ${summary.target} completed=${snapshot.completed ?? 0}/${snapshot.total_work_units ?? 0} running=${snapshot.running ?? 0} pending=${snapshot.pending ?? 0} blocked=${snapshot.blocked ?? 0} slowest=${snapshot.slowest_running?.label ?? "none"}\n`,
    );
  }
  const slowest = (summary.slowest_running_observations ?? []).map(formatSlowestObservation);
  if (slowest.length > 0) {
    process.stdout.write(`[SLOWEST-RUNNING] ${summary.target} ${slowest.join(",")}\n`);
  }
}

function helperArtifacts(runSummary, target = "") {
  const helpers = runSummary?.helper_units?.artifacts ?? [];
  if (!target) {
    return helpers;
  }
  return helpers.filter((entry) => entry.target === target);
}

function writeHelperLines(runSummary, target = "") {
  const helpers = helperArtifacts(runSummary, target);
  const sameRunRefs = new Map(
    (runSummary?.helper_units?.same_run_artifact_refs ?? []).map((ref) => [
      ref.target,
      ref,
    ]),
  );
  for (const helper of helpers) {
    const phaseSummaries = helper.phase_summaries ?? [];
    const failed = phaseSummaries.some((summary) => summary.status && summary.status !== "pass");
    process.stdout.write(
      `[HELPER] ${helper.target} status=${failed ? "fail" : "pass"} phases=${phaseSummaries.length} latest=${helper.latest || "none"}\n`,
    );
    const sameRunRef = sameRunRefs.get(helper.target);
    const sameRunRefPath = helper.same_run_artifact_ref || sameRunRef?.artifact || "";
    if (sameRunRefPath) {
      const refFile = absoluteArtifactPath(sameRunRefPath);
      if (existsSync(refFile)) {
        const refSummary = readJSON(refFile);
        process.stdout.write(
          `[HELPER-REF] ${helper.target} accounting=${refSummary.accounting_mode ?? "unknown"} scheduler_reused=${String(refSummary.scheduler_reused)} producer_artifacts=${refSummary.producer_artifacts?.length ?? 0} output_digest=${refSummary.output_digest_sha256 ?? "none"} artifact=${sameRunRefPath}\n`,
        );
      } else {
        process.stdout.write(
          `[HELPER-REF] ${helper.target} missing artifact=${sameRunRefPath}\n`,
        );
      }
    }
    for (const phase of phaseSummaries) {
      process.stdout.write(
        `[HELPER-PHASE] ${helper.target} label=${phase.label || "unknown"} status=${phase.status || "unknown"} artifact=${phase.artifact} runner_json=${phase.runner_json || "none"} stdout_log=${phase.stdout_log || "none"} stderr_log=${phase.stderr_log || "none"}\n`,
      );
    }
  }
}

function writeChildren(targetSummary) {
  const children = targetSummary?.children?.present ?? [];
  const skipped = targetSummary?.children?.skipped ?? [];
  if (children.length === 0 && skipped.length === 0) {
    process.stdout.write("[CHILDREN] none\n");
    return;
  }
  for (const child of children) {
    const totals = child.totals ?? child;
    const c = counts(totals);
    process.stdout.write(
      `[CHILD] ${child.target} status=${child.status}${failureClassField(child)} tests=${c.tests ?? 0} failed=${c.failed ?? 0} ${coverageCounts(totals)} duration=${duration(totals)} artifacts=${child.artifacts?.dir ?? child.own?.artifacts?.dir ?? ""}\n`,
    );
    writeFailureHeadline(child.target, child);
  }
  for (const child of skipped) {
    process.stdout.write(
      `[CHILD-SKIPPED] ${child.target} reason=${child.reason} failed_dependency=${child.failed_dependency || "unknown"} work_unit=${child.work_unit}\n`,
    );
  }
}

function writeRunChildren(runSummary) {
  const targets = runSummary?.evidence_targets?.summaries ?? [];
  if (targets.length === 0) {
    if (helperArtifacts(runSummary).length === 0) {
      process.stdout.write("[CHILDREN] none\n");
      return;
    }
  } else {
    for (const summary of targets) {
      const totals = summary.totals ?? summary;
      const c = counts(totals);
      process.stdout.write(
        `[TARGET] ${summary.target} status=${summary.status}${failureClassField(summary)} tests=${c.tests ?? 0} failed=${c.failed ?? 0} ${coverageCounts(totals)} duration=${duration(totals)} artifacts=${summary.own?.artifacts?.dir ?? summary.artifacts?.dir ?? ""}\n`,
      );
      writeFailureHeadline(summary.target, summary);
    }
  }
  writeHelperLines(runSummary);
}

function writeToolChildren(runDir, target, toolSummary) {
  if (target === "agent-finalize" && toolSummary) {
    const { summary } = loadFinalizeSummary(runDir, toolSummary);
    if (!summary) {
      process.stdout.write("[CHILDREN] none\n");
      return;
    }
    for (const action of summary.actions ?? []) {
      for (const substep of action.substeps ?? []) {
        process.stdout.write(
          `[FINALIZE-SUBSTEP] action=${action.action_id} id=${substep.id} target=${substep.target ?? "none"} status=${substep.status} skipped_reason=${substep.skipped_reason ?? "none"} summary_json=${substep.summary_json ?? "none"} stdout_log=${substep.stdout_log ?? "none"} stderr_log=${substep.stderr_log ?? "none"}\n`,
        );
      }
    }
    return;
  }
  const refs = [...(toolSummary?.evidence_targets ?? []), ...(toolSummary?.helper_units ?? [])];
  if (refs.length === 0) {
    process.stdout.write("[CHILDREN] none\n");
    return;
  }
  for (const ref of refs) {
    process.stdout.write(`[TOOL-REF] ${ref.target} status=${ref.status ?? "unknown"} run_root=${ref.run_root ?? relToRepo(runDir)}\n`);
  }
}

function absoluteArtifactPath(value) {
  if (!value) {
    return "";
  }
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function writeLogFile(target, file) {
  const absolute = absoluteArtifactPath(file);
  if (!absolute || !existsSync(absolute)) {
    return false;
  }
  const content = readFileSync(absolute, "utf8");
  process.stdout.write(`[LOG] ${target} ${relToRepo(absolute)}\n`);
  process.stdout.write(content);
  if (!content.endsWith("\n")) {
    process.stdout.write("\n");
  }
  return true;
}

function writeLogs(runDir, target, runSummary) {
  const logDir = path.join(runDir, target, "scheduler-logs");
  let wrote = false;
  if (existsSync(logDir)) {
    const files = readdirSync(logDir)
      .filter((name) => name.endsWith(".log"))
      .sort((left, right) => left.localeCompare(right));
    for (const name of files) {
      wrote = writeLogFile(target, path.join(logDir, name)) || wrote;
    }
  }
  for (const helper of helperArtifacts(runSummary, target)) {
    for (const phase of helper.phase_summaries ?? []) {
      wrote = writeLogFile(target, phase.stdout_log) || wrote;
      wrote = writeLogFile(target, phase.stderr_log) || wrote;
    }
  }
  if (!wrote) {
    process.stdout.write(`[LOGS] ${target} none\n`);
  }
}

function writeToolLogs(runDir, target, toolSummary) {
  let wrote = false;
  for (const artifact of toolSummary?.log_artifacts ?? []) {
    wrote = writeLogFile(target, artifact.path) || wrote;
  }
  if (target === "agent-finalize" && toolSummary) {
    const { summary } = loadFinalizeSummary(runDir, toolSummary);
    for (const action of summary?.actions ?? []) {
      for (const substep of action.substeps ?? []) {
        wrote = writeLogFile(target, substep.stdout_log) || wrote;
        wrote = writeLogFile(target, substep.stderr_log) || wrote;
      }
    }
  }
  if (!wrote) {
    process.stdout.write(`[LOGS] ${target} none\n`);
  }
}

function writeProgress(runDir, target, schedulerSummary) {
  const configured = schedulerSummary?.artifacts?.progress_summary_log ?? "";
  const fallback = path.join(runDir, target, "progress-summary.log");
  const file = configured ? absoluteArtifactPath(configured) : fallback;
  if (!file || !existsSync(file)) {
    process.stdout.write(`[PROGRESS-LOG] ${target} none\n`);
    return;
  }
  const content = readFileSync(file, "utf8");
  process.stdout.write(`[PROGRESS-LOG] ${target} ${relToRepo(file)}\n`);
  process.stdout.write(content);
  if (!content.endsWith("\n")) {
    process.stdout.write("\n");
  }
}

function accountingInventory(summary) {
  return summary?.totals?.inventory_by_coverage ?? summary?.inventory_by_coverage ?? {};
}

function accountingCounts(summary) {
  return summary?.totals?.counts ?? summary?.counts ?? {};
}

function ownerSource(item) {
  if (!item) {
    return "unknown";
  }
  if (item.coverage === "unmapped") {
    return "none";
  }
  if (item.coverage === "raw") {
    return "execution_topology";
  }
  if (item.id) {
    return String(item.id).startsWith("FE-") ? "frontend_phase_map" : "phase_map";
  }
  if (item.coverage === "tooling_support" || item.coverage === "unowned_regression" || item.coverage === "support") {
    return "classification_manifest";
  }
  return "summary";
}

function rowLikeInventoryTitle(item) {
  const value = String(item?.symbol_or_title ?? "");
  return (
    /\bFE-[A-Z]+-P\d+-\d+\b/.test(value) ||
    /\b[UIE]-\d+(?:-[A-Z0-9]+)*-\d+\b/.test(value)
  );
}

function writeAccountingForSummary(target, summary) {
  if (!summary) {
    process.stdout.write(`[ACCOUNTING] target=${target || "none"} missing\n`);
    return;
  }
  const c = accountingCounts(summary);
  const inventory = accountingInventory(summary);
  process.stdout.write(
    `[ACCOUNTING] target=${target || summary.target || "run"} tests=${c.tests ?? 0} ${coverageCounts({ counts: c })}\n`,
  );
  for (const coverage of coverageBuckets) {
    const items = inventory[coverage] ?? [];
    const byFile = new Map();
    for (const item of items) {
      const file = item.package_or_file || "(unknown)";
      if (!byFile.has(file)) {
        byFile.set(file, []);
      }
      byFile.get(file).push(item);
    }
    const rowLike = items.filter(rowLikeInventoryTitle).length;
    process.stdout.write(
      `[ACCOUNTING-BUCKET] target=${target || summary.target || "run"} coverage=${coverage} entries=${items.length} files=${byFile.size} row_like_titles=${rowLike}\n`,
    );
    const files = [...byFile.entries()].sort((left, right) => {
      if (right[1].length !== left[1].length) {
        return right[1].length - left[1].length;
      }
      return left[0].localeCompare(right[0]);
    });
    for (const [file, fileItems] of files) {
      const sources = [...new Set(fileItems.map(ownerSource))].sort();
      const sample = String(fileItems[0]?.symbol_or_title ?? "").replaceAll("\n", " ");
      process.stdout.write(
        `[ACCOUNTING-FILE] target=${target || summary.target || "run"} coverage=${coverage} file=${file} entries=${fileItems.length} owner_source=${sources.join("+") || "unknown"} sample=${JSON.stringify(sample)}\n`,
      );
    }
  }
}

function writeAccountingDetail(runSummary, targetSummary) {
  if (targetSummary) {
    writeAccountingForSummary(targetSummary.target, targetSummary);
    return;
  }
  const summaries = runSummary?.evidence_targets?.summaries ?? [];
  if (summaries.length === 0) {
    process.stdout.write("[ACCOUNTING] none\n");
    return;
  }
  for (const summary of summaries) {
    writeAccountingForSummary(summary.target, summary);
  }
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const { runDir, targetFromPath } = resolveRunContext(options);
  const runSummary = loadRunSummary(runDir);
  const target = options.target || targetFromPath || (runSummary ? "" : defaultToolTarget(runDir));
  const targetSummary = loadTargetSummary(runDir, target);
  const schedulerSummary = loadSchedulerSummary(runDir, target);
  const toolSummary = loadToolSummary(runDir, target);

  if (options.detail === "summary") {
    if (runSummary) {
      writeRunSummary(runDir, runSummary);
    } else if (toolSummary) {
      process.stdout.write(`[RUN] tool-summary-only target=${target || "none"} artifacts=${relToRepo(runDir)}\n`);
    } else {
      writeRunSummary(runDir, runSummary);
    }
    writeTargetSummary(runDir, targetSummary);
    writeSchedulerSummary(schedulerSummary);
    if (toolSummary) {
      writeToolSummary(runDir, target, toolSummary);
      if (target === "agent-finalize") {
        writeFinalizeSummary(runDir, toolSummary);
      }
    }
    writeHelperLines(runSummary, target);
    return;
  }
  if (options.detail === "children") {
    if (targetSummary) {
      writeChildren(targetSummary);
    } else if (toolSummary) {
      writeToolChildren(runDir, target, toolSummary);
    } else {
      writeRunChildren(runSummary);
    }
    return;
  }
  if (options.detail === "progress") {
    writeProgress(runDir, target, schedulerSummary);
    return;
  }
  if (options.detail === "accounting") {
    writeAccountingDetail(runSummary, targetSummary);
    return;
  }
  if (toolSummary) {
    writeToolLogs(runDir, target, toolSummary);
    return;
  }
  writeLogs(runDir, target, runSummary);
}

try {
  main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
