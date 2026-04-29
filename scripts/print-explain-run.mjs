#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { failureHeadlineForSummary } from "./lib/failure-taxonomy.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const validDetails = new Set(["summary", "children", "logs"]);

function usage() {
  process.stderr.write(
    "usage: print-explain-run.mjs --results-dir <root|run-dir> [--run-id <id>] [--target <target>] [--detail summary|children|logs]\n",
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
  if (options.detail === "logs" && !options.target) {
    throw new Error("DETAIL=logs requires TARGET=<target>");
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

function newestRunDir(resultsRoot) {
  const candidates = readdirSync(resultsRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(resultsRoot, entry.name))
    .filter((dir) => existsSync(path.join(dir, "run-summary.json")))
    .sort((left, right) => statSync(right).mtimeMs - statSync(left).mtimeMs);
  return candidates[0] ?? "";
}

function resolveRunDir(options) {
  const resultsDir = resolvePath(options.resultsDir);
  if (existsSync(path.join(resultsDir, "run-summary.json"))) {
    return resultsDir;
  }
  if (options.target && existsSync(path.join(resultsDir, options.target, "target-summary.json"))) {
    return resultsDir;
  }
  if (options.runId) {
    return path.join(resultsDir, options.runId);
  }
  const newest = newestRunDir(resultsDir);
  if (newest) {
    return newest;
  }
  throw new Error(`no run-summary.json found under ${resultsDir}; pass RUN_ID for a results root`);
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
}

function writeSchedulerSummary(summary) {
  if (!summary) {
    return;
  }
  const slowest = (summary.slowest_work_units ?? [])
    .map((entry) => `${entry.label}(${formatDuration(entry.duration_ms)})`)
    .join(",") || "none";
  process.stdout.write(
    `[SCHEDULER] ${summary.target} status=${summary.status}${failureClassField(summary)} completed_work_units=${summary.completed_work_units}/${summary.total_work_units} failed=${summary.failed_work_unit ?? "none"} slowest=${slowest} logs=${summary.artifacts?.scheduler_logs_dir ?? ""}\n`,
  );
  writeFailureHeadline(summary.target, summary);
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
    process.stdout.write("[CHILDREN] none\n");
    return;
  }
  for (const summary of targets) {
    const totals = summary.totals ?? summary;
    const c = counts(totals);
    process.stdout.write(
      `[TARGET] ${summary.target} status=${summary.status}${failureClassField(summary)} tests=${c.tests ?? 0} failed=${c.failed ?? 0} ${coverageCounts(totals)} duration=${duration(totals)} artifacts=${summary.own?.artifacts?.dir ?? summary.artifacts?.dir ?? ""}\n`,
    );
    writeFailureHeadline(summary.target, summary);
  }
}

function writeLogs(runDir, target) {
  const logDir = path.join(runDir, target, "scheduler-logs");
  if (!existsSync(logDir)) {
    throw new Error(`no scheduler logs found for ${target}: ${logDir}`);
  }
  const files = readdirSync(logDir)
    .filter((name) => name.endsWith(".log"))
    .sort((left, right) => left.localeCompare(right));
  if (files.length === 0) {
    process.stdout.write(`[LOGS] ${target} none\n`);
    return;
  }
  for (const name of files) {
    const file = path.join(logDir, name);
    const content = readFileSync(file, "utf8");
    process.stdout.write(`[LOG] ${target} ${relToRepo(file)}\n`);
    process.stdout.write(content);
    if (!content.endsWith("\n")) {
      process.stdout.write("\n");
    }
  }
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const runDir = resolveRunDir(options);
  const runSummary = loadRunSummary(runDir);
  const targetSummary = loadTargetSummary(runDir, options.target);
  const schedulerSummary = loadSchedulerSummary(runDir, options.target);

  if (options.detail === "summary") {
    writeRunSummary(runDir, runSummary);
    writeTargetSummary(runDir, targetSummary);
    writeSchedulerSummary(schedulerSummary);
    return;
  }
  if (options.detail === "children") {
    if (options.target) {
      writeChildren(targetSummary);
    } else {
      writeRunChildren(runSummary);
    }
    return;
  }
  writeLogs(runDir, options.target);
}

try {
  main();
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exit(1);
}
